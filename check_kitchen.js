async function checkKitchen() {
  try {
    const res = await fetch('http://localhost:8080/v1/kitchen/orders', {
      headers: { 'Authorization': 'Bearer mock-manager-token' }
    });
    console.log('Kitchen Orders:', await res.json());
  } catch (e) {
    console.error('Error:', e.message);
  }
}
checkKitchen();
