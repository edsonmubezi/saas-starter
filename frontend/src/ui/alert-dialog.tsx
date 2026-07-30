"use client"

import * as React from "react"
import * as AlertDialogPrimitive from "@radix-ui/react-alert-dialog"
import { AlertTriangle, CheckCircle2, Info, X, AlertCircle } from "lucide-react"
import clsx from "clsx"

/* Root + Trigger ----------------------------------------------------------- */
export const AlertDialog = AlertDialogPrimitive.Root
export const AlertDialogTrigger = AlertDialogPrimitive.Trigger

/* Overlay ------------------------------------------------------------------ */
export const AlertDialogOverlay = React.forwardRef<
  React.ElementRef<typeof AlertDialogPrimitive.Overlay>,
  React.ComponentPropsWithoutRef<typeof AlertDialogPrimitive.Overlay>
>(({ className, ...props }, ref) => (
  <AlertDialogPrimitive.Overlay
    ref={ref}
    className={clsx(
      "fixed inset-0 z-50 bg-black/70 backdrop-blur-sm",
      "data-[state=open]:animate-in data-[state=closed]:animate-out",
      "data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0",
      className
    )}
    {...props}
  />
))
AlertDialogOverlay.displayName = AlertDialogPrimitive.Overlay.displayName

/* Content (with Portal) ---------------------------------------------------- */
export const AlertDialogPortal = AlertDialogPrimitive.Portal

export const AlertDialogContent = React.forwardRef<
  React.ElementRef<typeof AlertDialogPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof AlertDialogPrimitive.Content> & {
    variant?: "default" | "destructive" | "success" | "warning"
  }
>(({ className, children, variant = "default", ...props }, ref) => {
  const variantStyles = {
    default: "from-blue-500/10 via-transparent to-purple-500/5",
    destructive: "from-red-500/10 via-transparent to-red-500/5",
    success: "from-emerald-500/10 via-transparent to-emerald-500/5",
    warning: "from-yellow-500/10 via-transparent to-orange-500/5",
  }

  return (
    <AlertDialogPortal>
      <AlertDialogOverlay />
      <AlertDialogPrimitive.Content
        ref={ref}
        className={clsx(
          "fixed left-1/2 top-1/2 z-50 w-[95vw] max-w-md -translate-x-1/2 -translate-y-1/2",
          "bg-gradient-to-br from-surface-elevated via-surface-secondary to-surface-elevated",
          "rounded-2xl border border-foreground/10 shadow-2xl shadow-black/30",
          "data-[state=open]:animate-in data-[state=closed]:animate-out",
          "data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0",
          "data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95",
          "data-[state=closed]:slide-out-to-left-1/2 data-[state=closed]:slide-out-to-top-[48%]",
          "data-[state=open]:slide-in-from-left-1/2 data-[state=open]:slide-in-from-top-[48%]",
          "duration-200",
          "overflow-hidden",
          className
        )}
        {...props}
      >
        {/* Gradient overlay for visual appeal */}
        <div className={clsx("absolute inset-0 bg-gradient-to-br opacity-50 pointer-events-none", variantStyles[variant])} />
        <div className="relative p-6">
          {children}
        </div>
      </AlertDialogPrimitive.Content>
    </AlertDialogPortal>
  )
})
AlertDialogContent.displayName = AlertDialogPrimitive.Content.displayName

/* Header layout helper ----------------------------------------------------- */
export const AlertDialogHeader = ({
  children,
  icon,
  variant = "default"
}: {
  children: React.ReactNode
  icon?: React.ReactNode
  variant?: "default" | "destructive" | "success" | "warning"
}) => {
  const iconColors = {
    default: "from-blue-500 to-blue-600 text-white shadow-blue-500/30",
    destructive: "from-red-500 to-red-600 text-white shadow-red-500/30",
    success: "from-emerald-500 to-emerald-600 text-white shadow-emerald-500/30",
    warning: "from-yellow-500 to-orange-500 text-white shadow-yellow-500/30",
  }

  const defaultIcons = {
    default: <Info className="w-5 h-5" />,
    destructive: <AlertTriangle className="w-5 h-5" />,
    success: <CheckCircle2 className="w-5 h-5" />,
    warning: <AlertCircle className="w-5 h-5" />,
  }

  return (
    <div className="flex items-start gap-4">
      <div className={clsx(
        "flex-shrink-0 w-11 h-11 rounded-xl bg-gradient-to-br flex items-center justify-center shadow-lg",
        iconColors[variant]
      )}>
        {icon || defaultIcons[variant]}
      </div>
      <div className="flex-1 space-y-1 pt-1">
        {children}
      </div>
    </div>
  )
}

export const AlertDialogFooter = ({ children }: { children: React.ReactNode }) => (
  <div className="mt-6 flex items-center justify-end gap-3">{children}</div>
)

/* Title/Description -------------------------------------------------------- */
export const AlertDialogTitle = React.forwardRef<
  React.ElementRef<typeof AlertDialogPrimitive.Title>,
  React.ComponentPropsWithoutRef<typeof AlertDialogPrimitive.Title>
>(({ className, ...props }, ref) => (
  <AlertDialogPrimitive.Title
    ref={ref}
    className={clsx("text-lg font-semibold text-foreground", className)}
    {...props}
  />
))
AlertDialogTitle.displayName = AlertDialogPrimitive.Title.displayName

export const AlertDialogDescription = React.forwardRef<
  React.ElementRef<typeof AlertDialogPrimitive.Description>,
  React.ComponentPropsWithoutRef<typeof AlertDialogPrimitive.Description>
>(({ className, ...props }, ref) => (
  <AlertDialogPrimitive.Description
    ref={ref}
    className={clsx("text-sm text-foreground/60 leading-relaxed", className)}
    {...props}
  />
))
AlertDialogDescription.displayName = AlertDialogPrimitive.Description.displayName

/* Buttons (Cancel/Action) -------------------------------------------------- */
export const AlertDialogCancel = React.forwardRef<
  React.ElementRef<typeof AlertDialogPrimitive.Cancel>,
  React.ComponentPropsWithoutRef<typeof AlertDialogPrimitive.Cancel>
>(({ className, ...props }, ref) => (
  <AlertDialogPrimitive.Cancel
    ref={ref}
    className={clsx(
      "px-5 py-2.5 rounded-xl text-sm font-medium transition-all duration-200",
      "bg-foreground/5 hover:bg-foreground/10 border border-foreground/10 hover:border-foreground/20",
      "text-foreground/70 hover:text-foreground",
      "focus:outline-none focus:ring-2 focus:ring-foreground/20",
      className
    )}
    {...props}
  />
))
AlertDialogCancel.displayName = AlertDialogPrimitive.Cancel.displayName

export const AlertDialogAction = React.forwardRef<
  React.ElementRef<typeof AlertDialogPrimitive.Action>,
  React.ComponentPropsWithoutRef<typeof AlertDialogPrimitive.Action> & {
    variant?: "default" | "destructive" | "success"
  }
>(({ className, variant = "default", ...props }, ref) => {
  const variantStyles = {
    default: "bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-500 hover:to-blue-600 shadow-blue-500/25 focus:ring-blue-500/50",
    destructive: "bg-gradient-to-r from-red-600 to-red-700 hover:from-red-500 hover:to-red-600 shadow-red-500/25 focus:ring-red-500/50",
    success: "bg-gradient-to-r from-emerald-600 to-emerald-700 hover:from-emerald-500 hover:to-emerald-600 shadow-emerald-500/25 focus:ring-emerald-500/50",
  }

  return (
    <AlertDialogPrimitive.Action
      ref={ref}
      className={clsx(
        "px-5 py-2.5 rounded-xl text-sm font-medium text-white transition-all duration-200",
        "shadow-lg",
        "focus:outline-none focus:ring-2",
        variantStyles[variant],
        className
      )}
      {...props}
    />
  )
})
AlertDialogAction.displayName = AlertDialogPrimitive.Action.displayName
