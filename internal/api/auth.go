package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/Maestrominds/lcdp-restaurant/internal/sqlc"
)

type loginVerifyRequest struct {
	Phone string `json:"phone"`
	OTP   string `json:"otp"`
}

type loginResponse struct {
	Status string `json:"status"`
}

type verifyResponse struct {
	User             db.User   `json:"user"`
	SessionExpiresAt time.Time `json:"session_expires_at"`
}

func (h *Handler) GetLogin(c *fiber.Ctx) error {
	phone := strings.TrimSpace(c.Query("phone"))
	role := strings.ToLower(strings.TrimSpace(c.Query("role")))
	if phone == "" || role == "" {
		return writeError(c, fiber.StatusBadRequest, "missing_fields", "phone and role are required")
	}

	if !isValidUserRole(role) {
		return writeError(c, fiber.StatusBadRequest, "invalid_role", "role must be waiter, manager, or admin")
	}

	ctx := context.Background()
	user, err := h.DB.GetUserByPhone(ctx, phone)
	if err != nil {
		return handleDBError(c, err)
	}

	if user.Role != db.UserRole(role) {
		return writeError(c, fiber.StatusUnauthorized, "invalid_credentials", "phone and role do not match")
	}

	if err := h.sendTwoFactorOTP(ctx, phone); err != nil {
		if errors.Is(err, errOTPNotConfigured) {
			return c.Status(fiber.StatusOK).JSON(loginResponse{Status: "otp_unavailable"})
		}
		return h.handleOTPError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(loginResponse{Status: "sent"})
}

func (h *Handler) VerifyLogin(c *fiber.Ctx) error {
	var req loginVerifyRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_body", "invalid json body")
	}

	req.Phone = strings.TrimSpace(req.Phone)
	req.OTP = strings.TrimSpace(req.OTP)
	if req.Phone == "" {
		return writeError(c, fiber.StatusBadRequest, "missing_fields", "phone is required")
	}
	if req.OTP == "" && h.Config.TwoFactorAPIKey != "" {
		return writeError(c, fiber.StatusBadRequest, "missing_fields", "phone and otp are required")
	}

	ctx := context.Background()
	user, err := h.DB.GetUserByPhone(ctx, req.Phone)
	if err != nil {
		return handleDBError(c, err)
	}

	if h.Config.TwoFactorAPIKey != "" {
		if err := h.verifyTwoFactorOTP(ctx, req.Phone, req.OTP); err != nil {
			return h.handleOTPError(c, err)
		}
	}

	token, err := randomToken(32)
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, "session_failed", "unable to create session")
	}

	expiresAt := time.Now().Add(h.Config.SessionTTL)
	_, err = h.DB.CreateSession(ctx, db.CreateSessionParams{
		UserID:    user.ID,
		TokenHash: sha256Hex(token),
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return handleDBError(c, err)
	}

	cookie := &fiber.Cookie{
		Name:     h.Config.CookieName,
		Value:    token,
		HTTPOnly: true,
		Secure:   h.Config.CookieSecure,
		Expires:  expiresAt,
		Path:     "/",
		SameSite: parseSameSite(h.Config.CookieSameSite),
		MaxAge:   int(h.Config.SessionTTL.Seconds()),
	}

	if h.Config.CookieDomain != "" {
		cookie.Domain = h.Config.CookieDomain
	}

	c.Cookie(cookie)

	return c.Status(fiber.StatusOK).JSON(verifyResponse{User: user, SessionExpiresAt: expiresAt})
}

func (h *Handler) AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Try Cookie first
		token := strings.TrimSpace(c.Cookies(h.Config.CookieName))
		
		// If no cookie, try Authorization header (Bearer token)
		if token == "" {
			authHeader := c.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if token == "" {
			return writeError(c, fiber.StatusUnauthorized, "unauthorized", "missing session cookie or authorization header")
		}

		// Allow mock tokens for development
		if token == "mock-manager-token" || token == "mock-owner-token" || token == "mock-waiter-token" {
			c.Locals("user_id", int64(1))   // Set a default user ID
			c.Locals("user_role", "admin") // Set a default role
			return c.Next()
		}

		ctx := context.Background()
		session, err := h.DB.GetSessionByTokenHash(ctx, sha256Hex(token))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return writeError(c, fiber.StatusUnauthorized, "unauthorized", "invalid session")
			}
			return handleDBError(c, err)
		}

		if _, err := h.DB.UpdateSessionLastSeen(ctx, session.ID); err != nil {
			return handleDBError(c, err)
		}

		c.Locals("user_id", session.UserID)

		return c.Next()
	}
}

func parseSameSite(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "strict":
		return "Strict"
	case "none":
		return "None"
	default:
		return "Lax"
	}
}

var (
	errOTPNotConfigured = errors.New("otp provider not configured")
	errOTPInvalid       = errors.New("otp invalid")
	errOTPProvider      = errors.New("otp provider error")
)

type twoFactorResponse struct {
	Status  string `json:"Status"`
	Details string `json:"Details"`
}

func (h *Handler) sendTwoFactorOTP(ctx context.Context, phone string) error {
	if h.Config.TwoFactorAPIKey == "" {
		return errOTPNotConfigured
	}

	endpoint := fmt.Sprintf(
		"https://2factor.in/API/V1/%s/SMS/%s/AUTOGEN",
		url.PathEscape(h.Config.TwoFactorAPIKey),
		url.PathEscape(phone),
	)

	resp, err := h.callTwoFactor(ctx, endpoint)
	if err != nil {
		return errOTPProvider
	}

	if !strings.EqualFold(resp.Status, "Success") {
		return errOTPProvider
	}

	return nil
}

func (h *Handler) verifyTwoFactorOTP(ctx context.Context, phone, otp string) error {
	if h.Config.TwoFactorAPIKey == "" {
		return errOTPNotConfigured
	}

	endpoint := fmt.Sprintf(
		"https://2factor.in/API/V1/%s/SMS/VERIFY3/%s/%s",
		url.PathEscape(h.Config.TwoFactorAPIKey),
		url.PathEscape(phone),
		url.PathEscape(otp),
	)

	resp, err := h.callTwoFactor(ctx, endpoint)
	if err != nil {
		return errOTPProvider
	}

	if !strings.EqualFold(resp.Status, "Success") {
		return errOTPInvalid
	}

	return nil
}

func (h *Handler) callTwoFactor(ctx context.Context, endpoint string) (twoFactorResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return twoFactorResponse{}, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return twoFactorResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return twoFactorResponse{}, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return twoFactorResponse{}, fmt.Errorf("otp provider status %d", resp.StatusCode)
	}

	var payload twoFactorResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return twoFactorResponse{}, err
	}

	return payload, nil
}

func (h *Handler) handleOTPError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, errOTPNotConfigured):
		return writeError(c, fiber.StatusServiceUnavailable, "otp_unavailable", "otp provider not configured")
	case errors.Is(err, errOTPInvalid):
		return writeError(c, fiber.StatusUnauthorized, "invalid_otp", "otp is invalid or expired")
	default:
		return writeError(c, fiber.StatusBadGateway, "otp_failed", "otp provider error")
	}
}

func (h *Handler) ManagerLogin(c *fiber.Ctx) error {
	log.Println("[AUTH] Manager Login Attempt")
	type loginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req loginReq
	if err := c.BodyParser(&req); err != nil {
		log.Printf("[AUTH ERROR] Body parse: %v\n", err)
		return writeError(c, fiber.StatusBadRequest, "invalid_body", "invalid json body")
	}

	log.Printf("[AUTH] Email: %s\n", req.Email)

	// Mock validation for now
	if req.Email == "manager@cafedeparis.com" && req.Password == "admin123" {
		log.Println("[AUTH SUCCESS] Manager")
		return c.JSON(fiber.Map{"token": "mock-manager-token", "user": fiber.Map{"name": "Manager", "role": "manager"}})
	}
	log.Println("[AUTH FAILED] Invalid credentials")
	return writeError(c, fiber.StatusUnauthorized, "invalid_credentials", "invalid email or password")
}

func (h *Handler) OwnerLogin(c *fiber.Ctx) error {
	log.Println("[AUTH] Owner Login Attempt")
	type loginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req loginReq
	if err := c.BodyParser(&req); err != nil {
		log.Printf("[AUTH ERROR] Body parse: %v\n", err)
		return writeError(c, fiber.StatusBadRequest, "invalid_body", "invalid json body")
	}

	log.Printf("[AUTH] Email: %s\n", req.Email)

	// Mock validation for now
	if req.Email == "owner@cafedeparis.com" && req.Password == "admin123" {
		log.Println("[AUTH SUCCESS] Owner")
		return c.JSON(fiber.Map{"token": "mock-owner-token", "user": fiber.Map{"name": "Owner", "role": "owner"}})
	}
	log.Println("[AUTH FAILED] Invalid credentials")
	return writeError(c, fiber.StatusUnauthorized, "invalid_credentials", "invalid email or password")
}

func (h *Handler) SendOtpFlutter(c *fiber.Ctx) error {
	log.Println("[AUTH] Flutter OTP Send Request")
	type otpReq struct {
		Phone string `json:"phone"`
	}
	var req otpReq
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_body", "invalid json body")
	}

	// For dev: always succeed
	log.Printf("[AUTH] Sending OTP to %s\n", req.Phone)
	return c.JSON(fiber.Map{"status": "Success"})
}

func (h *Handler) VerifyOtpFlutter(c *fiber.Ctx) error {
	log.Println("[AUTH] Flutter OTP Verify Request")
	type verifyReq struct {
		Phone string `json:"phone"`
		OTP   string `json:"otp"`
		Code  string `json:"code"` // Flutter sends 'code'
	}
	var req verifyReq
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_body", "invalid json body")
	}

	code := req.OTP
	if code == "" {
		code = req.Code
	}

	log.Printf("[AUTH] Verifying OTP %s for %s\n", code, req.Phone)

	// For dev: accept '123456' or any 4/6 digit code for easier testing
	if code == "123456" || len(code) == 4 || len(code) == 6 {
		log.Println("[AUTH SUCCESS] Flutter Waiter")
		return c.JSON(fiber.Map{
			"token": "mock-waiter-token",
			"user": fiber.Map{
				"id":    1,
				"name":  "Waiter User",
				"role":  "waiter",
				"phone": req.Phone,
			},
		})
	}

	log.Println("[AUTH FAILED] Invalid OTP")
	return writeError(c, fiber.StatusUnauthorized, "invalid_otp", "invalid or expired otp")
}

func isValidUserRole(role string) bool {
	switch role {
	case "waiter", "manager", "admin", "owner":
		return true
	}
	return false
}
