import { request, API_BASE } from './common'
import { API_PREFIX } from './apiPrefix'
import { loadTokens } from '../state/tokenStore'

const BASE = `${API_PREFIX.ORG}/chat`

export interface ChatThread {
  id: string
  title: string
  last_message: string
  message_count: number
  created_at: string
  updated_at: string
}

export interface ChatMessage {
  id: string
  thread_id: string
  role: 'user' | 'assistant'
  content: string
  created_at: string
}

export interface ChatAPIKeyConfig {
  api_key: string
  model: string
}

// Thread CRUD
export async function listThreads(): Promise<ChatThread[]> {
  const res = await request<any>(`${BASE}/threads`)
  return res?.data ?? []
}

export async function getMessages(threadId: string): Promise<ChatMessage[]> {
  const res = await request<any>(`${BASE}/threads/${encodeURIComponent(threadId)}/messages`)
  return res?.data ?? []
}

export async function updateThread(threadId: string, title: string) {
  return request(`${BASE}/threads/${encodeURIComponent(threadId)}`, {
    method: 'PUT',
    body: JSON.stringify({ title }),
  })
}

export async function deleteThread(threadId: string) {
  return request(`${BASE}/threads/${encodeURIComponent(threadId)}`, {
    method: 'DELETE',
  })
}

// API Key management
export async function getAPIKeyConfig(): Promise<ChatAPIKeyConfig> {
  const res = await request<any>(`${BASE}/api-key`)
  return res?.data ?? { api_key: '', model: 'gpt-4o-mini' }
}

export async function updateAPIKey(apiKey: string, model: string) {
  return request(`${BASE}/api-key`, {
    method: 'PUT',
    body: JSON.stringify({ api_key: apiKey, model }),
  })
}

// SSE Streaming helpers
function parseSSEStream(
  res: Response,
  onChunk: (text: string) => void,
  onThreadId: ((id: string) => void) | null,
  onDone: () => void,
  onError: (err: string) => void,
) {
  const reader = res.body?.getReader()
  if (!reader) { onError('No response body'); return }
  const decoder = new TextDecoder()
  let buffer = ''

  function pump(): Promise<void> {
    return reader!.read().then(({ done, value }) => {
      if (done) { onDone(); return }

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() ?? ''

      for (const line of lines) {
        if (!line.startsWith('data: ')) continue
        const data = line.slice(6)

        if (data === '[DONE]') { onDone(); return }
        if (data.startsWith('[ERROR]')) { onError(data.slice(8)); return }
        if (data.startsWith('[THREAD_ID]') && onThreadId) {
          onThreadId(data.slice(12).trim())
          continue
        }
        // Unescape JSON-encoded chunk
        try {
          onChunk(JSON.parse(`"${data}"`))
        } catch {
          onChunk(data)
        }
      }
      return pump()
    })
  }

  pump().catch((err) => {
    if (err?.name !== 'AbortError') onError(err?.message ?? 'Stream failed')
  })
}

function makeStreamRequest(
  url: string,
  body: object,
  onChunk: (text: string) => void,
  onThreadId: ((id: string) => void) | null,
  onDone: () => void,
  onError: (err: string) => void,
): AbortController {
  const controller = new AbortController()
  const { access } = loadTokens()

  fetch(`${API_BASE}${url}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${access}`,
    },
    body: JSON.stringify(body),
    signal: controller.signal,
  })
    .then(async (res) => {
      if (!res.ok) {
        const json = await res.json().catch(() => null)
        onError(json?.message ?? `Request failed: ${res.status}`)
        return
      }
      parseSSEStream(res, onChunk, onThreadId, onDone, onError)
    })
    .catch((err) => {
      if (err?.name !== 'AbortError') onError(err?.message ?? 'Connection failed')
    })

  return controller
}

// Create a new thread + stream first response
export function streamNewThread(
  message: string,
  title: string,
  onChunk: (text: string) => void,
  onThreadId: (id: string) => void,
  onDone: () => void,
  onError: (err: string) => void,
): AbortController {
  return makeStreamRequest(
    `${BASE}/threads/new/stream`,
    { message, title },
    onChunk,
    onThreadId,
    onDone,
    onError,
  )
}

// Send message to existing thread + stream response
export function streamMessage(
  threadId: string,
  message: string,
  onChunk: (text: string) => void,
  onDone: () => void,
  onError: (err: string) => void,
): AbortController {
  return makeStreamRequest(
    `${BASE}/threads/${encodeURIComponent(threadId)}/stream`,
    { message },
    onChunk,
    null,
    onDone,
    onError,
  )
}
