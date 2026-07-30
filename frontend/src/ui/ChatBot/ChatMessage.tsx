import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import clsx from 'clsx'
import { Bot, User } from 'lucide-react'

interface Props {
  role: 'user' | 'assistant'
  content: string
  isStreaming?: boolean
}

export default function ChatMessage({ role, content, isStreaming }: Props) {
  const isUser = role === 'user'

  return (
    <div className={clsx('flex gap-3 px-4 py-3', isUser ? 'flex-row-reverse' : 'flex-row')}>
      <div
        className={clsx(
          'w-7 h-7 rounded-lg flex items-center justify-center flex-shrink-0 mt-0.5',
          isUser ? 'bg-blue-500/20 text-blue-400' : 'bg-purple-500/20 text-purple-400'
        )}
      >
        {isUser ? <User size={14} /> : <Bot size={14} />}
      </div>
      <div
        className={clsx(
          'max-w-[85%] rounded-2xl px-4 py-2.5 text-sm',
          isUser
            ? 'bg-blue-500/15 text-foreground border border-blue-500/20'
            : 'bg-foreground/5 text-foreground border border-foreground/10'
        )}
      >
        {isUser ? (
          <p className="whitespace-pre-wrap">{content}</p>
        ) : (
          <div className="prose prose-sm dark:prose-invert max-w-none prose-p:my-1 prose-li:my-0.5 prose-headings:my-2 prose-pre:bg-foreground/10 prose-pre:text-foreground prose-code:text-blue-300">
            <ReactMarkdown remarkPlugins={[remarkGfm]}>{content}</ReactMarkdown>
            {isStreaming && (
              <span className="inline-block w-1.5 h-4 bg-purple-400 animate-pulse ml-0.5 rounded-sm" />
            )}
          </div>
        )}
      </div>
    </div>
  )
}
