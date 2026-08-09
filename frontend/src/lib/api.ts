import type { AnalysisResponse, ApiErrorEnvelope, ApiFieldError, HealthResponse, ScenarioRequest, ScenarioResponse } from '../types/api'

const DEFAULT_API_BASE_URL = 'http://localhost:8080'
const DEFAULT_TIMEOUT_MS = 12_000

export interface RequestOptions { signal?: AbortSignal; timeoutMs?: number }
export class ApiError extends Error {
  readonly status: number
  readonly code: string
  readonly fields: ApiFieldError[]
  readonly requestId?: string
  constructor(message: string, status = 0, code = 'REQUEST_FAILED', fields: ApiFieldError[] = [], requestId?: string) {
    super(message); this.name = 'ApiError'; this.status = status; this.code = code; this.fields = fields; this.requestId = requestId
  }
}

export function normalizeApiBaseUrl(value: string): string {
  const trimmed = value.trim().replace(/\/+$/, '')
  let parsed: URL
  try { parsed = new URL(trimmed) } catch { throw new Error('VITE_API_BASE_URL must be a valid absolute HTTP URL.') }
  if (!['http:', 'https:'].includes(parsed.protocol) || parsed.username || parsed.password || parsed.search || parsed.hash) throw new Error('VITE_API_BASE_URL must be an HTTP(S) origin or base path without credentials, query, or fragment.')
  return trimmed
}

export const apiBaseUrl = normalizeApiBaseUrl(import.meta.env?.VITE_API_BASE_URL || DEFAULT_API_BASE_URL)

type Fetch = typeof fetch
let fetchImplementation: Fetch = (...args) => fetch(...args)
export function setFetchImplementationForTests(value?: Fetch) { fetchImplementation = value ?? ((...args) => fetch(...args)) }

function isRecord(value: unknown): value is Record<string, unknown> { return typeof value === 'object' && value !== null && !Array.isArray(value) }
function isString(value: unknown): value is string { return typeof value === 'string' && value.length > 0 }
function isScenarioResponse(value: unknown): value is ScenarioResponse { return isRecord(value) && isString(value.requestId) && isRecord(value.scenario) && isString(value.scenario.id) && isRecord(value.scenario.definition) }
function isAnalysisResponse(value: unknown): value is AnalysisResponse { return isRecord(value) && isString(value.requestId) && isRecord(value.analysis) && isString(value.analysis.id) && isString(value.analysis.scenarioId) && isRecord(value.analysis.result) && typeof value.analysis.result.safe === 'boolean' && Array.isArray(value.analysis.result.violations) }
function isHealthResponse(value: unknown): value is HealthResponse { return isRecord(value) && isString(value.status) && isString(value.service) && isString(value.database) && isString(value.timestamp) }
function isErrorEnvelope(value: unknown): value is ApiErrorEnvelope { return isRecord(value) && isRecord(value.error) && isString(value.error.code) && isString(value.error.message) && isString(value.requestId) }

async function request<T>(path: string, init: RequestInit, validate: (value: unknown) => value is T, options: RequestOptions = {}): Promise<T> {
  const timeoutController = new AbortController(); const timeoutMs = options.timeoutMs ?? DEFAULT_TIMEOUT_MS
  const timeout = setTimeout(() => timeoutController.abort('timeout'), timeoutMs)
  const signal = options.signal ? AbortSignal.any([options.signal, timeoutController.signal]) : timeoutController.signal
  try {
    let response: Response
    try { response = await fetchImplementation(`${apiBaseUrl}${path}`, { ...init, signal, headers: { Accept: 'application/json', ...init.headers } }) }
    catch {
      if (options.signal?.aborted) throw new DOMException('The request was aborted.', 'AbortError')
      if (timeoutController.signal.aborted) throw new ApiError('The request timed out. Check the API and retry.', 0, 'REQUEST_TIMEOUT')
      throw new ApiError('The API could not be reached. Check the connection and retry.', 0, 'NETWORK_ERROR')
    }
    const requestId = response.headers.get('X-Request-ID') || undefined
    const contentType = response.headers.get('Content-Type') || ''
    const json = contentType.toLowerCase().includes('application/json') ? await response.json().catch(() => undefined) : undefined
    if (!response.ok) {
      if (isErrorEnvelope(json)) throw new ApiError(json.error.message, response.status, json.error.code, Array.isArray(json.error.fields) ? json.error.fields.filter((field): field is ApiFieldError => isRecord(field) && typeof field.field === 'string' && typeof field.message === 'string') : [], json.requestId || requestId)
      throw new ApiError(response.status >= 500 ? 'The API gateway returned an unexpected response. Please retry.' : 'The API request failed.', response.status, 'UNEXPECTED_RESPONSE', [], requestId)
    }
    if (!validate(json)) throw new ApiError('The API returned an unexpected response shape.', response.status, 'INVALID_RESPONSE', [], requestId)
    return json
  } finally { clearTimeout(timeout) }
}

const jsonHeaders = { 'Content-Type': 'application/json' }
export const createScenario = (body: ScenarioRequest, options?: RequestOptions) => request('/api/v1/scenarios', { method: 'POST', headers: jsonHeaders, body: JSON.stringify(body) }, isScenarioResponse, options)
export const updateScenario = (id: string, body: ScenarioRequest, options?: RequestOptions) => request(`/api/v1/scenarios/${encodeURIComponent(id)}`, { method: 'PUT', headers: jsonHeaders, body: JSON.stringify(body) }, isScenarioResponse, options)
export const getScenario = (id: string, options?: RequestOptions) => request(`/api/v1/scenarios/${encodeURIComponent(id)}`, { method: 'GET' }, isScenarioResponse, options)
export const createAnalysis = (scenarioId: string, options?: RequestOptions) => request(`/api/v1/scenarios/${encodeURIComponent(scenarioId)}/analyses`, { method: 'POST' }, isAnalysisResponse, options)
export const getAnalysis = (id: string, options?: RequestOptions) => request(`/api/v1/analyses/${encodeURIComponent(id)}`, { method: 'GET' }, isAnalysisResponse, options)
export const getHealth = (options?: RequestOptions) => request('/api/v1/health', { method: 'GET' }, isHealthResponse, options)
