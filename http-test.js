const http = require('http');

const options = {
  hostname: 'localhost',
  port: 8080,
  path: '/v1/orders',
  method: 'GET',
  headers: {
    'Authorization': 'Bearer mock-waiter-token'
  }
};

const req = http.request(options, (res) => {
  console.log(`Status: ${res.statusCode}`);
  let data = '';
  res.on('data', (chunk) => { data += chunk; });
  res.on('end', () => {
    try {
      console.log('Body:', JSON.stringify(JSON.parse(data), null, 2));
    } catch (e) {
      console.log('Body:', data);
    }
  });
});

req.on('error', (e) => {
  console.error(`Problem with request: ${e.message}`);
});

req.end();
