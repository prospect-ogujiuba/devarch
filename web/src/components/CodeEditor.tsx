import { classNames } from '../lib/utils'

interface CodeEditorProps {
  value: string
  onChange: (value: string) => void
  readOnly?: boolean
  invalid?: boolean
  minRows?: number
}

export function CodeEditor({ value, onChange, readOnly = false, invalid = false, minRows = 18 }: CodeEditorProps) {
  return (
    <textarea
      className={classNames('code-editor', invalid && 'code-editor--invalid')}
      spellCheck={false}
      value={value}
      readOnly={readOnly}
      rows={minRows}
      onChange={(event) => onChange(event.target.value)}
    />
  )
}
