import React, { Component, ErrorInfo, ReactNode } from 'react'

interface Props {
  children: ReactNode
  fallback?: ReactNode
}

interface State {
  hasError: boolean
  error: Error | null
  isChunkError: boolean
}

// Check if error is a chunk/module loading failure (happens after deployments)
function isChunkLoadError(error: Error): boolean {
  const message = error.message || ''
  const name = error.name || ''
  return (
    message.includes('Failed to fetch dynamically imported module') ||
    message.includes('Loading chunk') ||
    message.includes('Loading CSS chunk') ||
    name === 'ChunkLoadError' ||
    message.includes('Importing a module script failed')
  )
}

class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false, error: null, isChunkError: false }
  }

  static getDerivedStateFromError(error: Error): State {
    const isChunkError = isChunkLoadError(error)
    return { hasError: true, error, isChunkError }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    // Auto-reload on chunk errors (stale deployment)
    if (isChunkLoadError(error)) {
      // Only auto-reload once per session to avoid infinite loops
      const lastReload = sessionStorage.getItem('chunk-error-reload')
      const now = Date.now()

      if (!lastReload || now - parseInt(lastReload) > 10000) {
        sessionStorage.setItem('chunk-error-reload', now.toString())
        window.location.reload()
        return
      }
    }

    // Log error to monitoring service in production
    if (import.meta.env.PROD) {
      // TODO: Send to error monitoring service (e.g., Sentry)
      // Example: Sentry.captureException(error, { extra: errorInfo })
    }
  }

  handleRetry = () => {
    this.setState({ hasError: false, error: null, isChunkError: false })
  }

  handleHardReload = () => {
    // Force reload from server, bypassing cache
    window.location.reload()
  }

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback
      }

      // Special UI for chunk loading errors (deployment mismatch)
      if (this.state.isChunkError) {
        return (
          <div className="min-h-screen bg-surface-primary flex items-center justify-center p-4">
            <div className="max-w-md w-full bg-foreground/5 backdrop-blur-xl rounded-2xl border border-foreground/10 p-8 text-center">
              <div className="w-16 h-16 mx-auto mb-6 rounded-full bg-blue-500/20 flex items-center justify-center">
                <svg
                  className="w-8 h-8 text-blue-400"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                  />
                </svg>
              </div>
              <h2 className="text-xl font-bold text-foreground mb-2">App Updated</h2>
              <p className="text-foreground/60 mb-6">
                A new version of the application is available. Please refresh the page to get the latest updates.
              </p>
              <button
                onClick={this.handleHardReload}
                className="px-6 py-2.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors font-medium"
              >
                Refresh Now
              </button>
            </div>
          </div>
        )
      }

      return (
        <div className="min-h-screen bg-surface-primary flex items-center justify-center p-4">
          <div className="max-w-md w-full bg-foreground/5 backdrop-blur-xl rounded-2xl border border-foreground/10 p-8 text-center">
            <div className="w-16 h-16 mx-auto mb-6 rounded-full bg-red-500/20 flex items-center justify-center">
              <svg
                className="w-8 h-8 text-red-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
                />
              </svg>
            </div>
            <h2 className="text-xl font-bold text-foreground mb-2">Something went wrong</h2>
            <p className="text-foreground/60 mb-6">
              An unexpected error occurred. Please try again or contact support if the problem persists.
            </p>
            <div className="flex gap-3 justify-center">
              <button
                onClick={this.handleRetry}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
              >
                Try Again
              </button>
              <button
                onClick={() => window.location.href = '/'}
                className="px-4 py-2 bg-foreground/10 hover:bg-foreground/20 text-foreground rounded-lg transition-colors"
              >
                Go Home
              </button>
            </div>
            {import.meta.env.DEV && this.state.error && (
              <details className="mt-6 text-left">
                <summary className="text-sm text-foreground/50 cursor-pointer hover:text-foreground/70">
                  Error details (dev only)
                </summary>
                <pre className="mt-2 p-3 bg-black/30 rounded text-xs text-red-400 overflow-auto max-h-40">
                  {this.state.error.message}
                  {'\n\n'}
                  {this.state.error.stack}
                </pre>
              </details>
            )}
          </div>
        </div>
      )
    }

    return this.props.children
  }
}

export default ErrorBoundary
