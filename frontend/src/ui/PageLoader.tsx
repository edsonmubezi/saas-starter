// Loading spinner shown during lazy component loading
export default function PageLoader() {
  return (
    <div className="flex items-center justify-center min-h-[400px]">
      <div className="flex flex-col items-center gap-3">
        <div className="w-10 h-10 border-3 border-blue-500/30 border-t-blue-500 rounded-full animate-spin" />
        <span className="text-sm text-foreground/60">Loading...</span>
      </div>
    </div>
  )
}
