import assert from 'node:assert/strict';
import { test } from 'node:test';
import { getReadiness } from '../src/api.ts';

test('readiness uses the same-origin API gateway', async (t) => {
  const controller = new AbortController();
  t.mock.method(globalThis, 'fetch', async (url, options) => {
    assert.equal(url, '/api/ready');
    assert.equal(options.signal, controller.signal);
    return Response.json({ status: 'ok', database: 'ok' });
  });
  assert.deepEqual(await getReadiness(controller.signal), {
    status: 'ok',
    database: 'ok',
  });
});

test('readiness rejects unhealthy responses', async (t) => {
  t.mock.method(
    globalThis,
    'fetch',
    async () => new Response('', { status: 503 }),
  );
  await assert.rejects(getReadiness(), /503/);
});
