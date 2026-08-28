'use strict';

const test = require('node:test');
const assert = require('node:assert/strict');
const http = require('node:http');
const net = require('node:net');
const { appNameFromHost, targetForHost, createServer } = require('./router');

function listen(server) {
  return new Promise((resolve, reject) => {
    server.once('error', reject);
    server.listen(0, '127.0.0.1', () => resolve(server.address().port));
  });
}

function close(server) {
  return new Promise((resolve) => server.close(resolve));
}

function rawRequest(port, request) {
  return new Promise((resolve, reject) => {
    let response = '';
    const socket = net.connect(port, '127.0.0.1', () => socket.write(request));
    socket.setEncoding('utf8');
    socket.setTimeout(2000, () => socket.destroy(new Error('response timed out')));
    socket.on('data', (chunk) => { response += chunk; });
    socket.on('end', () => resolve(response));
    socket.on('error', reject);
  });
}

test('maps a valid .test hostname to its isolated app container', () => {
  assert.equal(appNameFromHost('store-front.test'), 'store-front');
  assert.equal(targetForHost('store-front.test'), 'node-store-front');
});

test('accepts a host header containing a port', () => {
  assert.equal(targetForHost('docs.test:443'), 'node-docs');
});

test('accepts at most 58 app characters so node-<app> remains a DNS label', () => {
  assert.equal(appNameFromHost(`${'a'.repeat(58)}.test`), 'a'.repeat(58));
  assert.equal(appNameFromHost(`${'a'.repeat(59)}.test`), null);
});

test('rejects unrelated, nested, uppercase, and unsafe hostnames', () => {
  for (const host of ['example.com', 'nested.app.test', 'Bad.test', '-bad.test', 'bad_name.test', '']) {
    assert.equal(targetForHost(host), null, host);
  }
});

test('only reserves the health path for the container healthcheck host', async (t) => {
  const upstream = http.createServer((request, response) => response.end(`app:${request.url}`));
  const upstreamPort = await listen(upstream);
  const router = createServer({ resolveTarget: () => ({ hostname: '127.0.0.1', port: upstreamPort }) });
  const routerPort = await listen(router);
  t.after(async () => { await close(router); await close(upstream); });

  const body = await new Promise((resolve, reject) => {
    const request = http.get({ hostname: '127.0.0.1', port: routerPort, path: '/__devarch/health', headers: { host: 'demo.test' } }, (response) => {
      let data = '';
      response.setEncoding('utf8');
      response.on('data', (chunk) => { data += chunk; });
      response.on('end', () => resolve(data));
    });
    request.on('error', reject);
  });
  assert.equal(body, 'app:/__devarch/health');
});

test('forwards an upstream rejection of a WebSocket upgrade', async (t) => {
  const upstream = http.createServer();
  upstream.on('upgrade', (request, socket) => {
    socket.end('HTTP/1.1 401 Unauthorized\r\nContent-Type: text/plain\r\nContent-Length: 6\r\nConnection: close\r\n\r\ndenied');
  });
  const upstreamPort = await listen(upstream);
  const router = createServer({ resolveTarget: () => ({ hostname: '127.0.0.1', port: upstreamPort }) });
  const routerPort = await listen(router);
  t.after(async () => { await close(router); await close(upstream); });

  const response = await rawRequest(routerPort, 'GET /socket HTTP/1.1\r\nHost: demo.test\r\nConnection: Upgrade\r\nUpgrade: websocket\r\n\r\n');
  assert.match(response, /^HTTP\/1\.1 401 Unauthorized/);
  assert.match(response, /\r\n\r\ndenied$/);
});

test('forwards an accepted WebSocket upgrade', async (t) => {
  const upstream = http.createServer();
  upstream.on('upgrade', (request, socket) => {
    socket.end('HTTP/1.1 101 Switching Protocols\r\nConnection: Upgrade\r\nUpgrade: websocket\r\n\r\naccepted');
  });
  const upstreamPort = await listen(upstream);
  const router = createServer({ resolveTarget: () => ({ hostname: '127.0.0.1', port: upstreamPort }) });
  const routerPort = await listen(router);
  t.after(async () => { await close(router); await close(upstream); });

  const response = await rawRequest(routerPort, 'GET /socket HTTP/1.1\r\nHost: demo.test\r\nConnection: Upgrade\r\nUpgrade: websocket\r\n\r\n');
  assert.match(response, /^HTTP\/1\.1 101 Switching Protocols/);
  assert.match(response, /\r\n\r\naccepted$/);
});
