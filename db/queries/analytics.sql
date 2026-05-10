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
  AND o.ordered_at < CURRENT_DATE;

-- name: GetWeeklyRevenue :many
SELECT 
    TO_CHAR(date_trunc('day', ordered_at), 'Dy') as day_name,
    COALESCE(SUM(m.price * o.quantity), 0)::FLOAT as revenue,
    date_trunc('day', ordered_at) as day_date
FROM orders o
JOIN menu_items m ON o.menu_item_id = m.id
WHERE ordered_at >= CURRENT_DATE - INTERVAL '6 days'
GROUP BY day_date, day_name
ORDER BY day_date;

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
