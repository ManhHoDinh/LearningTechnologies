const http = require('http');

const server = http.createServer((req, res) => {
  res.writeHead(200, { 'Content-Type': 'text/plain' });
  res.end('Hello from WEB Service! This is the web-svc running on port 80.\nServing web content.\n');
});

const PORT = process.env.PORT || 80;
server.listen(PORT, () => {
  console.log(`WEB Service running on port ${PORT}`);
});