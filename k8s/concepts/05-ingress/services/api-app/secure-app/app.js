const https = require('https');
const fs = require('fs');

const server = https.createServer((req, res) => {
  res.writeHead(200, { 'Content-Type': 'text/plain' });
  res.end('Hello from SECURE Service! This is the secure-svc running on port 443.\nThis is a secure HTTPS connection.\n');
});

const PORT = process.env.PORT || 443;
server.listen(PORT, () => {
  console.log(`SECURE Service running on port ${PORT}`);
});