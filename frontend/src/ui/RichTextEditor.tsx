import React, { useMemo } from 'react'
import ReactQuill from 'react-quill'
import 'react-quill/dist/quill.snow.css'
import './quill-dark.css'

type Props = {
  value: string
  onChange: (html: string) => void
  placeholder?: string
  className?: string
}

export default function RichTextEditor({ value, onChange, placeholder, className }: Props) {
  const modules = useMemo(
    () => ({
      toolbar: [
        [{ header: [1, 2, 3, false] }],
        [{ size: ['small', false, 'large', 'huge'] }],
        ['bold', 'italic', 'underline', 'strike'],
        [{ color: [] }, { background: [] }],
        [{ align: [] }],
        [{ list: 'ordered' }, { list: 'bullet' }],
        [{ indent: '-1' }, { indent: '+1' }],
        ['blockquote', 'code-block'],
        ['link', 'image'],
        ['clean'],
      ],
      clipboard: { matchVisual: false },
    }),
    [],
  )

  const formats = useMemo(
    () => [
      'header', 'size',
      'bold', 'italic', 'underline', 'strike',
      'color', 'background',
      'align',
      'list', 'bullet', 'indent',
      'blockquote', 'code-block',
      'link', 'image',
    ],
    [],
  )

  return (
    <div className={`quill-editor-wrapper ${className ?? ''}`}>
      <ReactQuill
        theme="snow"
        value={value}
        onChange={onChange}
        modules={modules}
        formats={formats}
        placeholder={placeholder}
      />
    </div>
  )
}
