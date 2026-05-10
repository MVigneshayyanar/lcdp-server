async function testOrder() {
  try {
    const res = await fetch('http://localhost:8080/v1/orders', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer mock-waiter-token'
      },
      body: JSON.stringify({
    tableId: 2, // Table 1
    items: [
        { menuItemId: 6, quantity: 3 } // Truffle Fries
    ]
})
    });
    console.log('Status:', res.status);
    console.log('Body:', await res.text());
  } catch (e) {
    console.error('Error:', e.message);
  }
}
testOrder();
