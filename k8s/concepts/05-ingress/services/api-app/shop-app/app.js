const http = require('http');

const server = http.createServer((req, res) => {
  res.writeHead(200, { 'Content-Type': 'text/plain' });
  res.end('Hello from SHOP Service! This is the shop-svc running on port 80.\nWelcome to our online shop!\n');
});

const PORT = process.env.PORT || 80;
server.listen(PORT, () => {
  console.log(`SHOP Service running on port ${PORT}`);
});