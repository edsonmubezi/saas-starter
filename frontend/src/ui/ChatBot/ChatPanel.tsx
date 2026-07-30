import { useState, useRef, useEffect, useCallback } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Send, Plus, MessageSquare, Trash2, Loader2 } from 'lucide-react'
import clsx from 'clsx'
import toast from 'react-hot-toast'
import ChatMessage from './ChatMessage'
import {
  listThreads,
  getMessages,
  deleteThread,
  streamNewThread,
  streamMessage,
  type ChatThread,
  type ChatMessage as ChatMessageType,
} from '../../utils/chat'

export default function ChatPanel() {
  const qc = useQueryClient()
  const [activeThreadId, setActiveThreadId] = useState<string | null>(null)
  const [input, setInput] = useState('')
  const [isStreaming, setIsStreaming] = useState(false)
  const [streamingText, setStreamingText] = useState('')
  const [showThreads, setShowThreads] = useState(false)
  const [localMessages, setLocalMessages] = useState<ChatMessageType[]>([])
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLTextAreaElement>(null)
  const abortRef = useRef<AbortController | null>(null)

  // Fetch threads
  const { data: threads = [] } = useQuery({
    queryKey: ['chat-threads'],
    queryFn: listThreads,
    staleTime: 30000,
  })

  // Fetch messages for active thread
  const { data: fetchedMessages = [], isLoading: loadingMessages } = useQuery({
    queryKey: ['chat-messages', activeThreadId],
    queryFn: () => getMessages(activeThreadId!),
    enabled: !!activeThreadId,
    staleTime: 10000,
  })

  // Merge fetched + local messages
  const messages = activeThreadId ? fetchedMessages : localMessages

  // Auto-scroll on new messages
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages, streamingText])

  // Focus input on mount
  useEffect(() => {
    inputRef.current?.focus()
  }, [activeThreadId])

  const handleSend = useCallback(() => {
    const text = input.trim()
    if (!text || isStreaming) return
    setInput('')
    setIsStreaming(true)
    setStreamingText('')

    if (!activeThreadId) {
      // New thread
      const userMsg: ChatMessageType = {
        id: 'temp-user',
        thread_id: '',
        role: 'user',
        content: text,
        created_at: new Date().toISOString(),
      }
      setLocalMessages([userMsg])

      abortRef.current = streamNewThread(
        text,
        '',
        (chunk) => setStreamingText((prev) => prev + chunk),
        (threadId) => {
          setActiveThreadId(threadId)
          setLocalMessages([])
          qc.invalidateQueries({ queryKey: ['chat-threads'] })
        },
        () => {
          setIsStreaming(false)
          setStreamingText('')
          qc.invalidateQueries({ queryKey: ['chat-threads'] })
          // Messages will be fetched via the query now that activeThreadId is set
          setTimeout(() => {
            qc.invalidateQueries({ queryKey: ['chat-messages'] })
          }, 200)
        },
        (err) => {
          setIsStreaming(false)
          setStreamingText('')
          toast.error(err || 'Chat failed')
        },
      )
    } else {
      // Existing thread
      abortRef.current = streamMessage(
        activeThreadId,
        text,
        (chunk) => setStreamingText((prev) => prev + chunk),
        () => {
          setIsStreaming(false)
          setStreamingText('')
          qc.invalidateQueries({ queryKey: ['chat-messages', activeThreadId] })
          qc.invalidateQueries({ queryKey: ['chat-threads'] })
        },
        (err) => {
          setIsStreaming(false)
          setStreamingText('')
          toast.error(err || 'Chat failed')
        },
      )

      // Optimistically add user message
      qc.setQueryData(['chat-messages', activeThreadId], (old: ChatMessageType[] | undefined) => [
        ...(old ?? []),
        {
          id: 'temp-' + Date.now(),
          thread_id: activeThreadId,
          role: 'user' as const,
          content: text,
          created_at: new Date().toISOString(),
        },
      ])
    }
  }, [input, isStreaming, activeThreadId, qc])

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  const handleNewChat = () => {
    if (abortRef.current) abortRef.current.abort()
    setActiveThreadId(null)
    setLocalMessages([])
    setStreamingText('')
    setIsStreaming(false)
    setShowThreads(false)
    setInput('')
  }

  const handleDeleteThread = async (id: string) => {
    try {
      await deleteThread(id)
      qc.invalidateQueries({ queryKey: ['chat-threads'] })
      if (activeThreadId === id) handleNewChat()
    } catch {
      toast.error('Failed to delete')
    }
  }

  const handleSelectThread = (thread: ChatThread) => {
    if (abortRef.current) abortRef.current.abort()
    setActiveThreadId(thread.id)
    setStreamingText('')
    setIsStreaming(false)
    setLocalMessages([])
    setShowThreads(false)
  }

  return (
    <div className="flex flex-col flex-1 min-h-0">
      {/* Toolbar */}
      <div className="flex items-center gap-2 px-3 py-2 border-b border-foreground/5">
        <button
          onClick={() => setShowThreads((s) => !s)}
          className={clsx(
            'px-2.5 py-1.5 rounded-lg text-xs font-medium transition-all',
            showThreads
              ? 'bg-blue-500/15 text-blue-400 border border-blue-500/30'
              : 'text-foreground/50 hover:bg-foreground/5'
          )}
        >
          <MessageSquare size={14} className="inline mr-1" />
          History
          {threads.length > 0 && (
            <span className="ml-1 text-[10px] opacity-60">({threads.length})</span>
          )}
        </button>
        <button
          onClick={handleNewChat}
          className="px-2.5 py-1.5 rounded-lg text-xs font-medium text-foreground/50 hover:bg-foreground/5 transition-all"
        >
          <Plus size={14} className="inline mr-1" />
          New
        </button>
      </div>

      {/* Thread List Overlay */}
      {showThreads && (
        <div className="border-b border-foreground/5 max-h-48 overflow-y-auto custom-scrollbar bg-surface-primary/50">
          {threads.length === 0 ? (
            <p className="text-xs text-foreground/30 text-center py-4">No conversations yet</p>
          ) : (
            threads.map((t) => (
              <div
                key={t.id}
                className={clsx(
                  'flex items-center gap-2 px-3 py-2 text-xs cursor-pointer transition-all',
                  activeThreadId === t.id
                    ? 'bg-blue-500/10 text-blue-300'
                    : 'text-foreground/60 hover:bg-foreground/5'
                )}
                onClick={() => handleSelectThread(t)}
              >
                <MessageSquare size={12} className="flex-shrink-0 opacity-40" />
                <span className="flex-1 truncate">{t.title}</span>
                <button
                  onClick={(e) => {
                    e.stopPropagation()
                    handleDeleteThread(t.id)
                  }}
                  className="p-1 rounded hover:bg-red-500/20 text-foreground/30 hover:text-red-400 transition-all"
                >
                  <Trash2 size={12} />
                </button>
              </div>
            ))
          )}
        </div>
      )}

      {/* Messages Area */}
      <div className="flex-1 overflow-y-auto custom-scrollbar">
        {messages.length === 0 && !streamingText && !isStreaming ? (
          <div className="flex flex-col items-center justify-center h-full text-center px-6">
            <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-blue-500/20 to-purple-500/20 flex items-center justify-center mb-4">
              <MessageSquare size={28} className="text-blue-400" />
            </div>
            <h3 className="text-sm font-semibold text-foreground mb-1">AI Assistant</h3>
            <p className="text-xs text-foreground/40 mb-6 max-w-[260px]">
              Ask me anything about using the system. I can guide you through any process step by step.
            </p>
            <div className="space-y-2 w-full max-w-[280px]">
              {['How do I manage users?', 'How do I set up roles and permissions?', 'How do I configure alerts?'].map(
                (q) => (
                  <button
                    key={q}
                    onClick={() => { setInput(q); setTimeout(() => inputRef.current?.focus(), 50) }}
                    className="w-full text-left px-3 py-2 rounded-xl bg-foreground/5 border border-foreground/10 text-xs text-foreground/60 hover:bg-foreground/10 hover:text-foreground/80 transition-all"
                  >
                    {q}
                  </button>
                )
              )}
            </div>
          </div>
        ) : (
          <>
            {loadingMessages && (
              <div className="flex items-center justify-center py-8">
                <Loader2 size={20} className="animate-spin text-foreground/30" />
              </div>
            )}
            {messages.map((msg) => (
              <ChatMessage key={msg.id} role={msg.role} content={msg.content} />
            ))}
            {streamingText && (
              <ChatMessage role="assistant" content={streamingText} isStreaming />
            )}
            <div ref={messagesEndRef} />
          </>
        )}
      </div>

      {/* Input Area */}
      <div className="border-t border-foreground/5 px-3 py-3">
        <div className="flex items-end gap-2">
          <textarea
            ref={inputRef}
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Ask about any process..."
            rows={1}
            className="flex-1 resize-none rounded-xl bg-foreground/5 border border-foreground/10 px-3 py-2.5 text-sm text-foreground placeholder:text-foreground/30 focus:outline-none focus:border-blue-500/40 focus:bg-foreground/[0.07] transition-all"
            style={{ maxHeight: '100px' }}
            disabled={isStreaming}
          />
          <button
            onClick={handleSend}
            disabled={!input.trim() || isStreaming}
            className={clsx(
              'p-2.5 rounded-xl transition-all flex-shrink-0',
              input.trim() && !isStreaming
                ? 'bg-gradient-to-r from-blue-500 to-purple-600 text-white hover:from-blue-600 hover:to-purple-700 shadow-lg shadow-blue-500/20'
                : 'bg-foreground/5 text-foreground/20 cursor-not-allowed'
            )}
          >
            {isStreaming ? (
              <Loader2 size={18} className="animate-spin" />
            ) : (
              <Send size={18} />
            )}
          </button>
        </div>
      </div>
    </div>
  )
}
