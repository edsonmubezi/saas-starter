// src/ui/confirm/ConfirmProvider.tsx
import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  PropsWithChildren,
} from "react"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "./alert-dialog"

/** ----------------------------- Types & Context ----------------------------- */

export type ConfirmVariant = "default" | "destructive" | "success" | "warning"

export type ConfirmOptions = {
  title?: string
  description?: string
  confirmText?: string
  cancelText?: string
  /**
   * Visual variant for the confirm dialog.
   * - default: Blue theme for general confirmations
   * - destructive: Red theme for delete/danger actions
   * - success: Green theme for positive confirmations
   * - warning: Yellow/Orange theme for warnings
   */
  variant?: ConfirmVariant
  /** Optional custom icon to override the default variant icon */
  icon?: React.ReactNode
}

export type ConfirmFn = (opts?: ConfirmOptions) => Promise<boolean>

const ConfirmContext = createContext<ConfirmFn | null>(null)

/** ---------------------------- Confirm Provider ----------------------------- */

type ProviderProps = PropsWithChildren<{
  /**
   * Defaults applied to every confirm call. Per-call options override these.
   */
  defaultOptions?: ConfirmOptions
}>

export function ConfirmProvider({ children, defaultOptions }: ProviderProps) {
  const [open, setOpen] = useState(false)
  const [opts, setOpts] = useState<ConfirmOptions>({})
  const resolverRef = useRef<((v: boolean) => void) | null>(null)

  /**
   * Open a confirmation dialog and resolve with true/false.
   */
  const confirm = useCallback<ConfirmFn>((options) => {
    return new Promise<boolean>((resolve) => {
      resolverRef.current = resolve
      setOpts(options ?? {})
      setOpen(true)
    })
  }, [])

  /**
   * Resolve the promise and close the dialog.
   */
  const resolveAndClose = useCallback((value: boolean) => {
    // Resolve first to let awaiting code continue immediately.
    resolverRef.current?.(value)
    resolverRef.current = null
    setOpen(false)
  }, [])

  /**
   * Handle user-initiated close (ESC / click outside).
   * Treat it as "Cancel".
   */
  const onOpenChange = useCallback(
    (next: boolean) => {
      if (!next) resolveAndClose(false)
    },
    [resolveAndClose],
  )

  /**
   * Safety: if the provider unmounts while a dialog is open,
   * resolve the pending promise as "false" to avoid hanging awaits.
   */
  useEffect(() => {
    return () => {
      if (resolverRef.current) {
        resolverRef.current(false)
        resolverRef.current = null
      }
    }
  }, [])

  const ctxValue = useMemo(() => confirm, [confirm])

  const merged: Required<Omit<ConfirmOptions, 'icon'>> & { icon?: React.ReactNode } = {
    title: "Are you sure?",
    description: "This action cannot be undone.",
    confirmText: "Confirm",
    cancelText: "Cancel",
    variant: "default",
    ...defaultOptions,
    ...opts,
  }

  // Map variant to action button variant
  const actionVariant = merged.variant === "destructive" ? "destructive"
    : merged.variant === "success" ? "success"
    : "default"

  return (
    <ConfirmContext.Provider value={ctxValue}>
      {children}

      <AlertDialog open={open} onOpenChange={onOpenChange}>
        <AlertDialogContent variant={merged.variant}>
          <AlertDialogHeader variant={merged.variant} icon={merged.icon}>
            <AlertDialogTitle>{merged.title}</AlertDialogTitle>
            {merged.description && (
              <AlertDialogDescription>{merged.description}</AlertDialogDescription>
            )}
          </AlertDialogHeader>

          <AlertDialogFooter>
            <AlertDialogCancel onClick={() => resolveAndClose(false)}>
              {merged.cancelText}
            </AlertDialogCancel>

            <AlertDialogAction
              onClick={() => resolveAndClose(true)}
              variant={actionVariant}
            >
              {merged.confirmText}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </ConfirmContext.Provider>
  )
}

/** ------------------------------- Hook ------------------------------------- */

export function useConfirm(): ConfirmFn {
  const ctx = useContext(ConfirmContext)
  if (!ctx) {
    throw new Error("useConfirm must be used within a <ConfirmProvider>")
  }
  return ctx
}

/** --------------------------- Convenience API ------------------------------ */

/**
 * Helper for delete confirmations. Use:
 *   const confirm = useConfirm();
 *   if (await confirmDelete(confirm, "deduction adjustment")) { ... }
 */
export async function confirmDelete(confirm: ConfirmFn, label?: string) {
  return confirm({
    title: `Delete ${label ?? "item"}?`,
    description: "This action cannot be undone. The item will be permanently removed.",
    confirmText: "Delete",
    cancelText: "Cancel",
    variant: "destructive",
  })
}

/**
 * Helper for warning confirmations. Use:
 *   const confirm = useConfirm();
 *   if (await confirmWarning(confirm, "This will reset all settings", "Reset")) { ... }
 */
export async function confirmWarning(confirm: ConfirmFn, description: string, action?: string) {
  return confirm({
    title: "Warning",
    description,
    confirmText: action ?? "Continue",
    cancelText: "Cancel",
    variant: "warning",
  })
}

/**
 * Helper for success/positive confirmations. Use:
 *   const confirm = useConfirm();
 *   if (await confirmAction(confirm, "Approve this request?", "Approve")) { ... }
 */
export async function confirmAction(confirm: ConfirmFn, description: string, action?: string) {
  return confirm({
    title: "Confirm Action",
    description,
    confirmText: action ?? "Confirm",
    cancelText: "Cancel",
    variant: "success",
  })
}
