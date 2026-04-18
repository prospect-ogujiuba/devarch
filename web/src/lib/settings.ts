const STORAGE_KEYS = {
  apiBase: 'devarch-v2.api-base',
  logTail: 'devarch-v2.log-tail',
  selectedWorkspace: 'devarch-v2.selected-workspace',
} as const

const DEFAULTS = {
  apiBase: '/api',
  logTail: 200,
  selectedWorkspace: '',
} as const

function canUseStorage() {
  return typeof window !== 'undefined' && typeof window.localStorage !== 'undefined'
}

function readString(key: string, fallback: string) {
  if (!canUseStorage()) return fallback
  const value = window.localStorage.getItem(key)
  return value && value.trim() ? value : fallback
}

function writeString(key: string, value: string) {
  if (!canUseStorage()) return
  const trimmed = value.trim()
  if (!trimmed) {
    window.localStorage.removeItem(key)
    return
  }
  window.localStorage.setItem(key, trimmed)
}

function readNumber(key: string, fallback: number) {
  const value = readString(key, String(fallback))
  const parsed = Number.parseInt(value, 10)
  return Number.isFinite(parsed) && parsed > 0 ? parsed : fallback
}

export function getApiBase() {
  return readString(STORAGE_KEYS.apiBase, DEFAULTS.apiBase)
}

export function setApiBase(value: string) {
  writeString(STORAGE_KEYS.apiBase, value || DEFAULTS.apiBase)
}

export function getDefaultLogTail() {
  return readNumber(STORAGE_KEYS.logTail, DEFAULTS.logTail)
}

export function setDefaultLogTail(value: number) {
  if (!canUseStorage()) return
  const normalized = Number.isFinite(value) && value > 0 ? Math.floor(value) : DEFAULTS.logTail
  window.localStorage.setItem(STORAGE_KEYS.logTail, String(normalized))
}

export function getSelectedWorkspace() {
  return readString(STORAGE_KEYS.selectedWorkspace, DEFAULTS.selectedWorkspace)
}

export function setSelectedWorkspace(value: string) {
  writeString(STORAGE_KEYS.selectedWorkspace, value)
}

export function resetUiPreferences() {
  if (!canUseStorage()) return
  window.localStorage.removeItem(STORAGE_KEYS.apiBase)
  window.localStorage.removeItem(STORAGE_KEYS.logTail)
  window.localStorage.removeItem(STORAGE_KEYS.selectedWorkspace)
}

export const uiPreferenceDefaults = DEFAULTS
