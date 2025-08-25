const http = require('http');

const server = http.createServer((req, res) => {
  res.writeHead(200, { 'Content-Type': 'text/plain' });
  res.end('Hello from API Service! This is the api-svc running on port 80.\nPath: ' + req.url + '\n');
});

const PORT = process.env.PORT || 80;
server.listen(PORT, () => {
  console.log(`API Service running on port ${PORT}`);
});