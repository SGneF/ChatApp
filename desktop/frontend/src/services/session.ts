export const TOKEN_KEY = 'lightchat.token'

const INSTANCE_NAME_PREFIX = 'lightchat.instance:'
let memoryToken = ''

function createInstanceId() {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) {
    return crypto.randomUUID()
  }
  return `${Date.now()}-${Math.random().toString(16).slice(2)}`
}

function getInstanceId() {
  if (typeof window === 'undefined') return 'default'

  if (!window.name.startsWith(INSTANCE_NAME_PREFIX)) {
    window.name = `${INSTANCE_NAME_PREFIX}${createInstanceId()}`
  }

  return window.name.slice(INSTANCE_NAME_PREFIX.length)
}

function getInstanceTokenKey() {
  return `${TOKEN_KEY}.${getInstanceId()}`
}

export function getToken() {
  if (memoryToken) return memoryToken

  memoryToken = sessionStorage.getItem(getInstanceTokenKey()) || ''
  return memoryToken
}

export function saveToken(token: string) {
  memoryToken = token
  sessionStorage.setItem(getInstanceTokenKey(), token)

  sessionStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(TOKEN_KEY)
}

export function clearToken() {
  memoryToken = ''
  sessionStorage.removeItem(getInstanceTokenKey())

  sessionStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(TOKEN_KEY)
}
