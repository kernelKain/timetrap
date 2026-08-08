import { useEffect, useState } from 'react'
import {
  LoaderCircle,
  ShieldCheck,
  TriangleAlert,
  Wifi,
} from 'lucide-react'

type HealthResponse = {
  status: string
  service: string
  version: string
  database: string
  timestamp: string
}

type ApiState =
  | { kind: 'checking' }
  | { kind: 'healthy'; health: HealthResponse }
  | { kind: 'unavailable'; message: string }

const apiBaseUrl = (
  import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
).replace(/\/$/, '')

const githubUrl =
  import.meta.env.VITE_GITHUB_URL ||
  'https://github.com/kernelKain/timetrap'

function App() {
  const [apiState, setApiState] = useState<ApiState>({ kind: 'checking' })

  useEffect(() => {
    const controller = new AbortController()

    async function checkApi() {
      try {
        const response = await fetch(`${apiBaseUrl}/api/v1/health`, {
          headers: {
            Accept: 'application/json',
          },
          signal: controller.signal,
        })

        if (!response.ok) {
          throw new Error(`The API returned status ${response.status}.`)
        }

        const health = (await response.json()) as HealthResponse

        if (health.status !== 'ok') {
          throw new Error('The API reported an unhealthy status.')
        }

        setApiState({ kind: 'healthy', health })
      } catch (error) {
        if (error instanceof DOMException && error.name === 'AbortError') {
          return
        }

        setApiState({
          kind: 'unavailable',
          message:
            error instanceof Error
              ? error.message
              : 'The API could not be reached.',
        })
      }
    }

    void checkApi()

    return () => {
      controller.abort()
    }
  }, [])

  return (
    <main className="min-h-screen bg-zinc-950 px-5 py-10 text-zinc-100 sm:px-8">
      <div className="mx-auto flex min-h-[calc(100vh-5rem)] max-w-5xl flex-col">
        <nav className="flex items-center justify-between">
          <div className="flex items-center gap-2 text-sm font-semibold">
            <ShieldCheck className="size-5 text-emerald-400" />
            <span>TimeTrap</span>
          </div>

          <a
            className="inline-flex items-center gap-2 rounded-full border border-zinc-800 px-4 py-2 text-sm text-zinc-300 transition hover:border-zinc-600 hover:text-white"
            href={githubUrl}
            target="_blank"
            rel="noreferrer"
          >
            GitHub
          </a>
        </nav>

        <section className="flex flex-1 items-center py-16">
          <div className="grid w-full gap-12 lg:grid-cols-[1.4fr_0.8fr] lg:items-center">
            <div>
              <p className="mb-4 text-sm font-semibold uppercase tracking-[0.2em] text-emerald-400">
                Design-time authorization verifier
              </p>

              <h1 className="max-w-3xl text-4xl font-bold tracking-tight sm:text-6xl">
                Find the minute trust outlives truth.
              </h1>

              <p className="mt-6 max-w-2xl text-lg leading-8 text-zinc-400">
                TimeTrap will help developers model time-dependent authorization
                and reveal when stale sessions, tokens, or cached entitlements
                remain valid longer than their source permissions.
              </p>

              <p className="mt-5 max-w-2xl text-sm leading-6 text-zinc-500">
                Phase 1 proves the connection between this React interface and
                the Go API. Scenario modeling and analysis arrive in later
                phases.
              </p>
            </div>

            <section
              aria-live="polite"
              className="rounded-3xl border border-zinc-800 bg-zinc-900/70 p-6 shadow-2xl shadow-black/30"
            >
              <p className="text-sm font-medium text-zinc-400">System status</p>

              {apiState.kind === 'checking' && (
                <div className="mt-5 flex items-start gap-3">
                  <LoaderCircle className="mt-0.5 size-5 animate-spin text-amber-400" />
                  <div>
                    <p className="font-semibold">Checking API…</p>
                    <p className="mt-1 text-sm text-zinc-500">
                      Contacting {apiBaseUrl}
                    </p>
                  </div>
                </div>
              )}

              {apiState.kind === 'healthy' && (
                <div className="mt-5">
                  <div className="flex items-start gap-3">
                    <Wifi className="mt-0.5 size-5 text-emerald-400" />
                    <div>
                      <p className="font-semibold text-emerald-300">
                        API healthy
                      </p>
                      <p className="mt-1 text-sm text-zinc-500">
                        The frontend successfully reached the Go service.
                      </p>
                    </div>
                  </div>

                  <dl className="mt-6 grid gap-3 text-sm">
                    <div className="flex justify-between gap-4 border-t border-zinc-800 pt-3">
                      <dt className="text-zinc-500">Service</dt>
                      <dd>{apiState.health.service}</dd>
                    </div>
                    <div className="flex justify-between gap-4 border-t border-zinc-800 pt-3">
                      <dt className="text-zinc-500">Version</dt>
                      <dd>{apiState.health.version}</dd>
                    </div>
                    <div className="flex justify-between gap-4 border-t border-zinc-800 pt-3">
                      <dt className="text-zinc-500">Database</dt>
                      <dd>{apiState.health.database}</dd>
                    </div>
                  </dl>
                </div>
              )}

              {apiState.kind === 'unavailable' && (
                <div className="mt-5 flex items-start gap-3">
                  <TriangleAlert className="mt-0.5 size-5 text-red-400" />
                  <div>
                    <p className="font-semibold text-red-300">
                      API unavailable
                    </p>
                    <p className="mt-1 text-sm text-zinc-500">
                      {apiState.message} Check that the Go API is running.
                    </p>
                  </div>
                </div>
              )}
            </section>
          </div>
        </section>

        <footer className="border-t border-zinc-900 pt-6 text-sm text-zinc-600">
          TimeTrap verifies bounded models; it does not scan live systems or
          guarantee complete application security.
        </footer>
      </div>
    </main>
  )
}

export default App