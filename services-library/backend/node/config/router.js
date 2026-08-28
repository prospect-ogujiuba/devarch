'use strict';

const http = require('node:http');

const listenPort = Number.parseInt(process.env.PORT || '3000', 10);
const targetPort = Number.parseInt(process.env.DEVARCH_NODE_TARGET_PORT || '3000', 10);
const requestTimeout = Number.parseInt(process.env.DEVARCH_NODE_REQUEST_TIMEOUT_MS || '120000', 10);
const appPattern = /^[a-z0-9](?:[a-z0-9-]{0,56}[a-z0-9])?$/;

function appNameFromHost(hostHeader) {
  if (typeof hostHeader !== 'string' || hostHeader.length === 0) return null;
  const hostname = hostHeader.split(':', 1)[0];
  if (hostname !== hostname.toLowerCase() || !hostname.endsWith('.test')) return null;
  const appName = hostname.slice(0, -5);
  return appPattern.test(appName) ? appName : null;
}

function targetForHost(hostHeader) {
  const appName = appNameFromHost(hostHeader);
  return appName ? `node-${appName}` : null;
}

function sendError(response, statusCode, message) {
  if (response.headersSent) {
    response.destroy();
    return;
  }
  response.writeHead(statusCode, { 'content-type': 'text/plain; charset=utf-8' });
  response.end(`${message}\n`);
}

function isLocalHealthCheck(request) {
  if (request.url !== '/__devarch/health') return false;
  const hostname = (request.headers.host || '').split(':', 1)[0];
  return hostname === '127.0.0.1' || hostname === 'localhost' || hostname === 'node';
}

function endpointForTarget(target, options) {
  if (typeof options.resolveTarget === 'function') {
    return options.resolveTarget(target.slice(5), target);
  }
  return { hostname: target, port: options.targetPort || targetPort };
}

function proxyRequest(request, response, options = {}) {
  if (isLocalHealthCheck(request)) {
    response.writeHead(200, { 'content-type': 'text/plain; charset=utf-8' });
    response.end('ok\n');
    return;
  }

  const target = targetForHost(request.headers.host);
  if (!target) {
    sendError(response, 400, 'A valid <app>.test Host header is required');
    return;
  }

  const endpoint = endpointForTarget(target, options);
  const upstream = http.request({
    hostname: endpoint.hostname,
    port: endpoint.port,
    method: request.method,
    path: request.url,
    headers: request.headers,
    timeout: requestTimeout,
  }, (upstreamResponse) => {
    response.writeHead(upstreamResponse.statusCode || 502, upstreamResponse.headers);
    upstreamResponse.pipe(response);
  });

  upstream.on('timeout', () => upstream.destroy(new Error('upstream timeout')));
  upstream.on('error', () => sendError(response, 503, `JavaScript app ${target.slice(5)} is not running`));
  request.on('aborted', () => upstream.destroy());
  request.pipe(upstream);
}

function writeRawResponseHead(socket, response) {
  socket.write(`HTTP/1.1 ${response.statusCode} ${response.statusMessage || ''}\r\n`);
  for (let index = 0; index < response.rawHeaders.length; index += 2) {
    socket.write(`${response.rawHeaders[index]}: ${response.rawHeaders[index + 1]}\r\n`);
  }
  socket.write('\r\n');
}

function proxyUpgrade(request, socket, head, options = {}) {
  const target = targetForHost(request.headers.host);
  if (!target) {
    socket.end('HTTP/1.1 400 Bad Request\r\nConnection: close\r\n\r\n');
    return;
  }

  const endpoint = endpointForTarget(target, options);
  const upstream = http.request({
    hostname: endpoint.hostname,
    port: endpoint.port,
    method: request.method,
    path: request.url,
    headers: request.headers,
  });

  upstream.setTimeout(requestTimeout, () => upstream.destroy(new Error('upstream timeout')));
  upstream.on('response', (upstreamResponse) => {
    writeRawResponseHead(socket, upstreamResponse);
    upstreamResponse.on('error', () => socket.destroy());
    upstreamResponse.pipe(socket);
  });
  upstream.on('upgrade', (upstreamResponse, upstreamSocket, upstreamHead) => {
    writeRawResponseHead(socket, upstreamResponse);
    if (upstreamHead.length) socket.write(upstreamHead);
    if (head.length) upstreamSocket.write(head);
    upstreamSocket.pipe(socket).pipe(upstreamSocket);
  });
  upstream.on('error', () => {
    if (!socket.destroyed) socket.end('HTTP/1.1 503 Service Unavailable\r\nConnection: close\r\n\r\n');
  });
  socket.on('error', () => upstream.destroy());
  socket.on('close', () => upstream.destroy());
  upstream.end();
}

function createServer(options = {}) {
  const server = http.createServer((request, response) => proxyRequest(request, response, options));
  server.on('upgrade', (request, socket, head) => proxyUpgrade(request, socket, head, options));
  return server;
}

if (require.main === module) {
  createServer().listen(listenPort, '0.0.0.0', () => {
    console.log(`[node-router] listening on 0.0.0.0:${listenPort}`);
  });
}

module.exports = { appNameFromHost, targetForHost, createServer };
