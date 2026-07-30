
import React from 'react'
import Modal from './Modal'

export default function ConfirmModal({
  open,
  title = 'Confirm',
  message,
  confirmText = 'Confirm',
  cancelText = 'Cancel',
  onCancel,
  onConfirm,
  danger = false,
}: {
  open: boolean
  title?: string
  message: React.ReactNode
  confirmText?: string
  cancelText?: string
  danger?: boolean
  onCancel: () => void
  onConfirm: () => Promise<void> | void
}) {
  return (
    <Modal open={open} onClose={onCancel} title={title}>
      <div className="grid gap-4">
        <div className="text-foreground/80">{message}</div>
        <div className="flex justify-end gap-2">
          <button className="btn" onClick={onCancel}>{cancelText}</button>
          <button
            className={`btn ${danger ? 'bg-red-600 hover:bg-red-700 text-white' : ''}`}
            onClick={() => void onConfirm()}
          >
            {confirmText}
          </button>
        </div>
      </div>
    </Modal>
  )
}
