const { Client } = require('pg');
require('dotenv').config();

async function checkData() {
  const client = new Client({
    connectionString: process.env.DATABASE_URL,
    ssl: { rejectUnauthorized: false }
  });
  await client.connect();
  
  console.log("--- DINING TABLES ---");
  const tables = await client.query('SELECT id, number, name, status FROM dining_tables ORDER BY number');
  console.table(tables.rows);
  
  console.log("--- RECENT ORDERS ---");
  const orders = await client.query('SELECT o.id, t.name as table, m.name as item, o.quantity, o.status FROM orders o JOIN dining_tables t ON o.table_id = t.id JOIN menu_items m ON o.menu_item_id = m.id ORDER BY o.created_at DESC LIMIT 5');
  console.table(orders.rows);
  
  await client.end();
}

checkData();
