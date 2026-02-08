import { useEffect, useRef } from 'react'
import { EditorState } from '@codemirror/state'
import { EditorView, keymap, lineNumbers, highlightActiveLine } from '@codemirror/view'
import { defaultKeymap, history, historyKeymap } from '@codemirror/commands'
import { oneDark } from '@codemirror/theme-one-dark'
import { json } from '@codemirror/lang-json'
import { yaml } from '@codemirror/lang-yaml'
import { xml } from '@codemirror/lang-xml'

interface Props {
  value: string
  onChange: (value: string) => void
  language?: string
  readOnly?: boolean
  className?: string
  autoHeight?: boolean
}

function getLangExtension(lang?: string) {
  switch (lang) {
    case 'json':
      return json()
    case 'yaml':
    case 'yml':
      return yaml()
    case 'xml':
    case 'html':
      return xml()
    default:
      return []
  }
}

function detectLang(filePath: string): string {
  const ext = filePath.split('.').pop()?.toLowerCase() ?? ''
  if (['json'].includes(ext)) return 'json'
  if (['yaml', 'yml'].includes(ext)) return 'yaml'
  if (['xml', 'html', 'htm'].includes(ext)) return 'xml'
  return ''
}

export function CodeEditor({ value, onChange, language, readOnly = false, className, autoHeight = false }: Props) {
  const containerRef = useRef<HTMLDivElement>(null)
  const viewRef = useRef<EditorView | null>(null)
  const onChangeRef = useRef(onChange)
  onChangeRef.current = onChange

  useEffect(() => {
    if (!containerRef.current) return

    const lang = language ?? ''
    const langExt = getLangExtension(lang)

    const state = EditorState.create({
      doc: value,
      extensions: [
        lineNumbers(),
        highlightActiveLine(),
        history(),
        keymap.of([...defaultKeymap, ...historyKeymap]),
        oneDark,
        langExt,
        EditorView.updateListener.of((update) => {
          if (update.docChanged) {
            onChangeRef.current(update.state.doc.toString())
          }
        }),
        ...(readOnly ? [EditorState.readOnly.of(true)] : []),
        EditorView.theme({
          '&': { fontSize: '13px', ...(autoHeight ? {} : { height: '400px' }) },
          '.cm-scroller': { overflow: autoHeight ? 'visible' : 'auto' },
        }),
      ],
    })

    const view = new EditorView({ state, parent: containerRef.current })
    viewRef.current = view

    return () => view.destroy()
  }, [language, readOnly, autoHeight])

  useEffect(() => {
    const view = viewRef.current
    if (!view) return
    const currentValue = view.state.doc.toString()
    if (currentValue !== value) {
      view.dispatch({
        changes: { from: 0, to: currentValue.length, insert: value },
      })
    }
  }, [value])

  return <div ref={containerRef} className={className ?? 'rounded-md border overflow-hidden'} />
}

export { detectLang }
