package api

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type errorPayload struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeError(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(errorPayload{Error: errorBody{Code: code, Message: message}})
}

func handleDBError(c *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return writeError(c, fiber.StatusNotFound, "not_found", "resource not found")
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return writeError(c, fiber.StatusConflict, "unique_violation", "resource already exists")
		case "23503":
			return writeError(c, fiber.StatusConflict, "foreign_key_violation", "invalid reference")
		case "23502":
			return writeError(c, fiber.StatusBadRequest, "not_null_violation", "missing required field")
		default:
			return writeError(c, fiber.StatusBadRequest, "database_error", pgErr.Message)
		}
	}

	return writeError(c, fiber.StatusInternalServerError, "internal_error", "unexpected error")
}
