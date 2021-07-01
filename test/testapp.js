//
// Simple NodeJS Test App
//

const http = require('http');
const os = require('os');

console.log("[] testapp server starting...");

var handler = function(request, response) {
    console.log("   Received request from (" + request.connection.remoteAddress + ")");    
    response.writeHead(200, {'Content-Type': 'text/plain'});
    response.end("   Hostname (" + os.hostname + ")\n");
};

var www = http.createServer(handler);
www.listen(8080);
