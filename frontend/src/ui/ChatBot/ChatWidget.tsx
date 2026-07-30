import { useState, useCallback } from 'react'
import { MessageSquare, X, Minimize2 } from 'lucide-react'
import clsx from 'clsx'
import ChatPanel from './ChatPanel'

export default function ChatWidget() {
  const [isOpen, setIsOpen] = useState(false)
  const [isMinimized, setIsMinimized] = useState(false)

  const toggle = useCallback(() => {
    if (isMinimized) {
      setIsMinimized(false)
    } else {
      setIsOpen((o) => !o)
    }
  }, [isMinimized])

  return (
    <>
      {/* Chat Panel */}
      {isOpen && (
        <div
          className={clsx(
            'fixed z-[45] transition-all duration-300 ease-out',
            isMinimized
              ? 'bottom-20 right-6 w-80 h-12'
              : 'bottom-20 right-6 w-[420px] h-[600px] max-h-[80vh] max-sm:w-[calc(100vw-2rem)] max-sm:right-4 max-sm:bottom-16 max-sm:h-[70vh]',
            'rounded-2xl border border-foreground/10 bg-surface-secondary shadow-2xl shadow-black/20 overflow-hidden flex flex-col'
          )}
        >
          {/* Header */}
          <div className="flex items-center justify-between px-4 py-3 border-b border-foreground/10 bg-gradient-to-r from-blue-500/10 to-purple-500/10 flex-shrink-0">
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center">
                <MessageSquare size={16} className="text-white" />
              </div>
              <div>
                <h3 className="text-sm font-semibold text-foreground">AI Assistant</h3>
                <p className="text-[10px] text-foreground/50">AI-powered help</p>
              </div>
            </div>
            <div className="flex items-center gap-1">
              <button
                onClick={() => setIsMinimized((m) => !m)}
                className="p-1.5 rounded-lg hover:bg-foreground/10 text-foreground/50 hover:text-foreground transition-all"
                title="Minimize"
              >
                <Minimize2 size={14} />
              </button>
              <button
                onClick={() => setIsOpen(false)}
                className="p-1.5 rounded-lg hover:bg-foreground/10 text-foreground/50 hover:text-foreground transition-all"
                title="Close"
              >
                <X size={14} />
              </button>
            </div>
          </div>

          {/* Chat content */}
          {!isMinimized && <ChatPanel />}
        </div>
      )}

      {/* Floating Action Button */}
      <button
        onClick={toggle}
        className={clsx(
          'fixed bottom-6 right-6 z-[45] w-14 h-14 rounded-full shadow-lg transition-all duration-300',
          'bg-gradient-to-br from-blue-500 to-purple-600 hover:from-blue-600 hover:to-purple-700',
          'flex items-center justify-center text-white',
          'hover:scale-105 active:scale-95',
          isOpen && !isMinimized && 'opacity-0 pointer-events-none'
        )}
        title="AI Help Assistant"
      >
        <MessageSquare size={24} />
      </button>
    </>
  )
}
