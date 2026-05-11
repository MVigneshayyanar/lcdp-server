-- name: GetTodayRevenue :one
SELECT COALESCE(SUM(m.price * o.quantity), 0)::FLOAT as revenue
FROM orders o
JOIN menu_items m ON o.menu_item_id = m.id
WHERE o.ordered_at >= CURRENT_DATE;

-- name: GetYesterdayRevenue :one
SELECT COALESCE(SUM(m.price * o.quantity), 0)::FLOAT as revenue
FROM orders o
JOIN menu_items m ON o.menu_item_id = m.id
WHERE o.ordered_at >= CURRENT_DATE - INTERVAL '1 day' 
  AND o.ordered_at < CURRENT_DATE;

-- name: GetTodayOrderCount :one
SELECT COUNT(*)::BIGINT as count
FROM orders
WHERE ordered_at >= CURRENT_DATE;

-- name: GetYesterdayOrderCount :one
SELECT COUNT(*)::BIGINT as count
FROM orders
WHERE ordered_at >= CURRENT_DATE - INTERVAL '1 day'
  AND ordered_at < CURRENT_DATE;

-- name: GetWeeklyRevenue :many
WITH days AS (
    SELECT generate_series(CURRENT_DATE - INTERVAL '6 days', CURRENT_DATE, '1 day')::date AS d
)
SELECT 
    TO_CHAR(d, 'Dy') as day_name,
    COALESCE(SUM(m.price * o.quantity), 0)::FLOAT as revenue,
    d as day_date
FROM days d
LEFT JOIN orders o ON o.ordered_at::date = d
LEFT JOIN menu_items m ON o.menu_item_id = m.id
GROUP BY d
ORDER BY d;

-- name: GetTopItems :many
SELECT 
    m.name, 
    SUM(o.quantity)::BIGINT as units_sold, 
    SUM(m.price * o.quantity)::FLOAT as revenue
FROM orders o
JOIN menu_items m ON o.menu_item_id = m.id
WHERE o.ordered_at >= CURRENT_DATE - INTERVAL '30 days'
GROUP BY m.name
ORDER BY units_sold DESC
LIMIT 5;

-- name: GetTotalSales30Days :one
SELECT COUNT(*)::BIGINT FROM orders WHERE ordered_at >= CURRENT_DATE - INTERVAL '30 days';

-- name: GetTopCategory :one
SELECT COALESCE(category, 'Mains')::TEXT
FROM menu_items m JOIN orders o ON o.menu_item_id = m.id 
GROUP BY category ORDER BY COUNT(*) DESC LIMIT 1;

-- name: GetPeakHour :one
SELECT COALESCE(TO_CHAR(ordered_at, 'HH24:00'), '19:00')::TEXT
FROM orders GROUP BY 1 ORDER BY COUNT(*) DESC LIMIT 1;

-- name: GetAvgItemsPerOrder :one
SELECT ROUND(COALESCE(AVG(quantity), 0)::NUMERIC, 2)::FLOAT FROM orders;

-- name: GetCategoryDistribution :many
SELECT COALESCE(category, 'Others')::TEXT as name, (COUNT(*) * 100.0 / NULLIF((SELECT COUNT(*) FROM orders), 0))::INT as percentage
FROM menu_items m JOIN orders o ON o.menu_item_id = m.id
GROUP BY category;

-- name: GetInventoryStats :one
SELECT 
    COUNT(*)::BIGINT as total_items,
    COUNT(*) FILTER (WHERE quantity <= min_stock / 2)::BIGINT as critical_items,
    COUNT(*) FILTER (WHERE quantity <= min_stock AND quantity > min_stock / 2)::BIGINT as low_stock_items
FROM inventory_items;

-- name: GetLowStockItems :many
SELECT name, ROUND(quantity::NUMERIC, 2)::FLOAT as current, unit, ROUND(min_stock::NUMERIC, 2)::FLOAT as threshold, 
    CASE WHEN quantity <= min_stock / 2 THEN 'critical' ELSE 'low' END as severity
FROM inventory_items 
WHERE quantity <= min_stock
ORDER BY quantity ASC LIMIT 10;
