export function detectLang(filePath: string): string {
  const ext = filePath.split('.').pop()?.toLowerCase() ?? ''
  if (['json'].includes(ext)) return 'json'
  if (['yaml', 'yml'].includes(ext)) return 'yaml'
  if (['xml', 'html', 'htm'].includes(ext)) return 'xml'
  return ''
}
