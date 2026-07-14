import type { MessageResponse, MessageType } from '../api/message'

export type ChatSocketStatus = 'idle' | 'connecting' | 'connected' | 'closed' | 'error'

export interface ChatSocketConnectedEvent {
  type: 'connected'
  data: {
    user_id: number
    message: string
  }
}

export interface ChatSocketMessageEvent {
  type: 'chat_ack' | 'chat_message'
  data: MessageResponse
}

export interface ChatSocketErrorEvent {
  type: 'chat_error'
  data: {
    message: string
  }
}

export type ChatSocketEvent = ChatSocketConnectedEvent | ChatSocketMessageEvent | ChatSocketErrorEvent

type MessageListener = (event: ChatSocketEvent) => void
type StatusListener = (status: ChatSocketStatus) => void

function resolveWsURL(token: string) {
  const configured = import.meta.env.VITE_WS_BASE_URL as string | undefined
  if (configured) {
    const url = new URL(configured)
    url.searchParams.set('token', token)
    return url.toString()
  }

  const apiBase = (import.meta.env.VITE_API_BASE_URL as string | undefined) || 'http://127.0.0.1:8080/api'
  const url = new URL(apiBase)
  url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
  url.pathname = '/ws'
  url.search = ''
  url.searchParams.set('token', token)
  return url.toString()
}

class ChatSocketClient {
  private socket: WebSocket | null = null
  private token = ''
  private status: ChatSocketStatus = 'idle'
  private reconnectTimer: number | null = null
  private reconnectAttempts = 0
  private manualClose = false
  private messageListeners = new Set<MessageListener>()
  private statusListeners = new Set<StatusListener>()

  get currentStatus() {
    return this.status
  }

  connect(token: string) {
    if (!token) return
    if (this.socket && this.token === token && (this.status === 'connected' || this.status === 'connecting')) return

    this.disconnect()
    this.token = token
    this.manualClose = false
    this.setStatus('connecting')

    const socket = new WebSocket(resolveWsURL(token))
    this.socket = socket

    socket.onopen = () => {
      this.reconnectAttempts = 0
      this.setStatus('connected')
    }

    socket.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data) as ChatSocketEvent
        this.messageListeners.forEach((listener) => listener(message))
      } catch {
        this.setStatus('error')
      }
    }

    socket.onerror = () => {
      this.setStatus('error')
    }

    socket.onclose = () => {
      if (this.socket === socket) {
        this.socket = null
      }
      this.setStatus('closed')
      if (!this.manualClose && this.token) {
        this.scheduleReconnect()
      }
    }
  }

  disconnect() {
    this.manualClose = true
    if (this.reconnectTimer !== null) {
      window.clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    if (this.socket) {
      this.socket.close()
      this.socket = null
    }
    this.setStatus('closed')
  }

  sendChatMessage(conversationId: number, content: string, type: MessageType = 'text') {
    const payload = {
      type: 'chat_message',
      data: {
        conversation_id: conversationId,
        type,
        content,
      },
    }

    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      throw new Error('WebSocket 未连接')
    }

    this.socket.send(JSON.stringify(payload))
  }

  onMessage(listener: MessageListener) {
    this.messageListeners.add(listener)
    return () => this.messageListeners.delete(listener)
  }

  onStatus(listener: StatusListener) {
    this.statusListeners.add(listener)
    listener(this.status)
    return () => this.statusListeners.delete(listener)
  }

  private scheduleReconnect() {
    if (this.reconnectTimer !== null) return
    const delay = Math.min(1000 * 2 ** this.reconnectAttempts, 10000)
    this.reconnectAttempts += 1
    this.reconnectTimer = window.setTimeout(() => {
      this.reconnectTimer = null
      if (!this.manualClose && this.token) {
        this.connect(this.token)
      }
    }, delay)
  }

  private setStatus(status: ChatSocketStatus) {
    this.status = status
    this.statusListeners.forEach((listener) => listener(status))
  }
}

export const chatSocket = new ChatSocketClient()
