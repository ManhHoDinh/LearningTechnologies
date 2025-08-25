const http = require('http');

const server = http.createServer((req, res) => {
  res.writeHead(200, { 'Content-Type': 'text/plain' });
  res.end('Hello from NGINX Service! This is the nginx-svc running on port 80.\n');
});

const PORT = process.env.PORT || 80;
server.listen(PORT, () => {
  console.log(`NGINX Service running on port ${PORT}`);
});