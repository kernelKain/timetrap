import assert from 'node:assert/strict'
import test from 'node:test'
import { ApiError, createScenario, getAnalysis, normalizeApiBaseUrl, setFetchImplementationForTests } from '../src/lib/api.ts'

const scenario = { name: 'Test', horizonMinutes: 10, objects: [{ clientId: 'a', name: 'A', kind: 'session', events: [{ type: 'issue', atMinute: 0 }] }], invariants: [{ type: 'max_validity', targetObjectId: 'a', maximumMinutes: 5 }] }
const scenarioEnvelope = { scenario: { id: '00000000-0000-4000-8000-000000000001', definition: scenario, createdAt: '2026-01-01T00:00:00Z', updatedAt: '2026-01-01T00:00:00Z' }, requestId: 'request-1' }
const jsonResponse = (body, status = 200, headers = {}) => new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json', ...headers } })

test.afterEach(() => setFetchImplementationForTests())
test('normalizes one or many trailing slashes', () => assert.equal(normalizeApiBaseUrl(' http://localhost:8080/// '), 'http://localhost:8080'))
test('rejects unusable API base URLs', () => assert.throws(() => normalizeApiBaseUrl('/api'), /absolute HTTP URL/))
test('decodes successful JSON', async () => { setFetchImplementationForTests(async () => jsonResponse(scenarioEnvelope, 201)); assert.equal((await createScenario(scenario)).scenario.id, scenarioEnvelope.scenario.id) })
test('decodes structured validation errors and request ID', async () => {
  setFetchImplementationForTests(async () => jsonResponse({ error: { code: 'VALIDATION_FAILED', message: 'Scenario validation failed.', fields: [{ field: 'name', message: 'is required' }] }, requestId: 'request-422' }, 422, { 'X-Request-ID': 'request-422' }))
  await assert.rejects(createScenario(scenario), (error) => error instanceof ApiError && error.status === 422 && error.code === 'VALIDATION_FAILED' && error.fields[0].field === 'name' && error.requestId === 'request-422')
})
test('sanitizes a non-JSON gateway error', async () => { setFetchImplementationForTests(async () => new Response('<html>proxy secret</html>', { status: 502, headers: { 'Content-Type': 'text/html' } })); await assert.rejects(getAnalysis('id'), (error) => error instanceof ApiError && !error.message.includes('proxy secret')) })
test('reports network failure without retrying POST', async () => { let calls = 0; setFetchImplementationForTests(async () => { calls++; throw new TypeError('offline') }); await assert.rejects(createScenario(scenario), (error) => error instanceof ApiError && error.code === 'NETWORK_ERROR'); assert.equal(calls, 1) })
test('reports a request timeout', async () => {
  setFetchImplementationForTests((_url, init) => new Promise((_resolve, reject) => init.signal.addEventListener('abort', () => reject(init.signal.reason), { once: true })))
  await assert.rejects(createScenario(scenario, { timeoutMs: 1 }), (error) => error instanceof ApiError && error.code === 'REQUEST_TIMEOUT')
})
test('preserves AbortError cancellation', async () => { const controller = new AbortController(); controller.abort(); setFetchImplementationForTests(async (_url, init) => { throw init.signal.reason }); await assert.rejects(createScenario(scenario, { signal: controller.signal }), (error) => error instanceof DOMException && error.name === 'AbortError') })
test('rejects unexpected success payloads', async () => { setFetchImplementationForTests(async () => jsonResponse({ ok: true })); await assert.rejects(getAnalysis('id'), (error) => error instanceof ApiError && error.code === 'INVALID_RESPONSE') })
