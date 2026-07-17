import { defineStore } from 'pinia'
import {
  getMessageHistory,
  type MessageReadResponse,
  type MessageResponse,
  type MessageRevokeResponse,
  type MessageType,
} from '../api/message'
import { getToken } from '../services/session'
import {
  chatSocket,
  type ChatSocketEvent,
  type ChatSocketOfflineSyncData,
  type ChatSocketStatus,
} from '../services/chatSocket'
import { useConversationStore } from './conversation'

const REVOKED_PREVIEW = '撤回了一条消息'

function sortMessages(list: MessageResponse[]) {
  return [...list].sort((a, b) => {
    const diff = new Date(a.create_time).getTime() - new Date(b.create_time).getTime()
    if (diff !== 0) return diff
    return a.id - b.id
  })
}

function isReadableSentStatus(status: string) {
  return status === 'normal' || status === 'sent'
}

function buildMessagePreview(message: MessageResponse) {
  if (message.status === 'revoked') return REVOKED_PREVIEW

  switch (message.type) {
    case 'image':
      return '[图片]'
    case 'file':
      return '[文件]'
    case 'voice':
      return '[语音]'
    default:
      return message.content
  }
}

export const useMessageStore = defineStore('message', {
  state: () => ({
    currentUserId: 0,
    activeReadingConversationId: 0,
    messagesByConversation: {} as Record<number, MessageResponse[]>,
    loadingHistory: false,
    socketStatus: 'idle' as ChatSocketStatus,
    socketError: '',
    unsubscribeMessage: null as null | (() => void),
    unsubscribeStatus: null as null | (() => void),
  }),

  getters: {
    isSocketConnected: (state) => state.socketStatus === 'connected',
  },

  actions: {
    setCurrentUserId(userId: number) {
      this.currentUserId = userId
    },

    setActiveReadingConversation(conversationId: number | null | undefined) {
      this.activeReadingConversationId = conversationId || 0
    },

    clearActiveReadingConversation(conversationId?: number | null) {
      if (!conversationId || this.activeReadingConversationId === conversationId) {
        this.activeReadingConversationId = 0
      }
    },

    isActivelyReadingConversation(conversationId: number) {
      return this.activeReadingConversationId === conversationId
    },

    connect() {
      const token = getToken()
      if (!token) return

      if (!this.unsubscribeMessage) {
        this.unsubscribeMessage = chatSocket.onMessage((event) => this.handleSocketEvent(event))
      }
      if (!this.unsubscribeStatus) {
        this.unsubscribeStatus = chatSocket.onStatus((status) => {
          this.socketStatus = status
        })
      }

      chatSocket.connect(token)
    },

    disconnect() {
      this.unsubscribeMessage?.()
      this.unsubscribeStatus?.()
      this.unsubscribeMessage = null
      this.unsubscribeStatus = null
      chatSocket.disconnect()
      this.socketStatus = 'closed'
      this.socketError = ''
      this.activeReadingConversationId = 0
      this.messagesByConversation = {}
    },

    async loadHistory(conversationId: number) {
      this.loadingHistory = true
      try {
        const history = await getMessageHistory(conversationId, 1, 50)
        this.messagesByConversation[conversationId] = sortMessages(history.list)
      } finally {
        this.loadingHistory = false
      }
    },

    sendMessage(conversationId: number, type: MessageType, content: string) {
      this.socketError = ''
      chatSocket.sendChatMessage(conversationId, content.trim(), type)
    },

    sendText(conversationId: number, content: string) {
      this.sendMessage(conversationId, 'text', content)
    },

    markConversationRead(conversationId: number) {
      if (!conversationId) return

      this.socketError = ''
      if (this.isSocketConnected) {
        chatSocket.sendMessageRead(conversationId)
        return
      }

      const conversationStore = useConversationStore()
      void conversationStore.markRead(conversationId).catch(() => undefined)
    },

    revokeMessage(messageId: number) {
      if (!messageId) return
      this.socketError = ''
      chatSocket.sendMessageRevoke(messageId)
    },

    handleSocketEvent(event: ChatSocketEvent) {
      if (event.type === 'connected') {
        this.socketError = ''
        if (event.data.user_id) {
          this.currentUserId = event.data.user_id
        }
        return
      }

      if (event.type === 'chat_error') {
        this.socketError = event.data.message || '消息操作失败'
        return
      }

      if (event.type === 'offline_sync') {
        void this.handleOfflineSync(event.data)
        return
      }

      if (event.type === 'message_read' || event.type === 'message_read_ack') {
        this.applyReadSync(event.data)
        return
      }

      if (event.type === 'message_revoke' || event.type === 'message_revoke_ack') {
        this.applyRevokeSync(event.data)
        return
      }

      if (event.type === 'chat_ack' || event.type === 'chat_message') {
        this.appendMessage(event.data)
        this.updateConversationByMessage(event.data)
      }
    },

    async handleOfflineSync(data: ChatSocketOfflineSyncData) {
      const conversationStore = useConversationStore()
      await conversationStore.fetchConversationList().catch(() => undefined)

      const offlineMessages = data.conversations.flatMap((item) => item.messages || [])
      offlineMessages.forEach((message) => this.appendMessage(message))

      const readingConversationId = this.activeReadingConversationId
      if (!readingConversationId) return

      const currentHasOfflineMessages = data.conversations.some(
        (item) => item.conversation_id === readingConversationId,
      )
      if (!currentHasOfflineMessages) return

      await this.loadHistory(readingConversationId).catch(() => undefined)
      this.markConversationRead(readingConversationId)
    },

    appendMessage(message: MessageResponse) {
      const conversationId = this.resolveConversationId(message)
      const list = this.messagesByConversation[conversationId] || []
      if (list.some((item) => item.id === message.id)) return

      this.messagesByConversation[conversationId] = sortMessages([...list, message])
    },

    resolveConversationId(message: MessageResponse) {
      const conversationStore = useConversationStore()
      if (conversationStore.conversationList.some((item) => item.id === message.conversation_id)) {
        return message.conversation_id
      }

      const otherUserId = message.sender_id === this.currentUserId ? message.receiver_id : message.sender_id
      const conversation = conversationStore.conversationList.find((item) => item.target_id === otherUserId)
      return conversation?.id || message.conversation_id
    },

    resolveConversationIdByUsers(userA: number, userB: number) {
      const conversationStore = useConversationStore()
      const otherUserId = userA === this.currentUserId ? userB : userA
      return conversationStore.conversationList.find((item) => item.target_id === otherUserId)?.id || 0
    },

    updateConversationByMessage(message: MessageResponse) {
      const conversationStore = useConversationStore()
      const conversationId = this.resolveConversationId(message)
      const conversation = conversationStore.conversationList.find((item) => item.id === conversationId)
      const now = message.create_time || new Date().toISOString()

      if (!conversation) {
        void conversationStore.fetchConversationList()
        return
      }

      const isIncoming = message.sender_id !== this.currentUserId
      const isReading = this.activeReadingConversationId === conversation.id
      let nextUnreadCount = conversation.unread_count

      if (isIncoming) {
        nextUnreadCount = isReading ? 0 : conversation.unread_count + 1
      }

      const nextConversation = {
        ...conversation,
        last_message_id: message.id,
        last_message: buildMessagePreview(message),
        unread_count: nextUnreadCount,
        update_time: now,
      }

      conversationStore.upsertConversation(nextConversation)

      if (conversationStore.currentConversation?.id === conversation.id) {
        conversationStore.currentConversation = {
          ...conversationStore.currentConversation,
          last_message_id: message.id,
          last_message: buildMessagePreview(message),
          unread_count: nextUnreadCount,
          update_time: now,
        }
      }

      if (isIncoming && isReading) {
        this.markConversationRead(conversation.id)
      }
    },

    applyReadSync(data: MessageReadResponse) {
      const conversationStore = useConversationStore()
      const readerIsMe = data.reader_id === this.currentUserId
      const conversationId = readerIsMe
        ? data.conversation_id
        : this.resolveConversationIdByUsers(data.reader_id, data.target_id)

      if (!conversationId) {
        void conversationStore.fetchConversationList()
        return
      }

      const list = this.messagesByConversation[conversationId]
      if (list) {
        this.messagesByConversation[conversationId] = list.map((message) => {
          const shouldMarkRead = message.sender_id === data.target_id
            && message.receiver_id === data.reader_id
            && isReadableSentStatus(message.status)

          return shouldMarkRead ? { ...message, status: 'read' } : message
        })
      }

      if (readerIsMe) {
        const conversation = conversationStore.conversationList.find((item) => item.id === conversationId)
        if (conversation) {
          conversationStore.upsertConversation({ ...conversation, unread_count: 0 })
        }
        if (conversationStore.currentConversation?.id === conversationId) {
          conversationStore.currentConversation = { ...conversationStore.currentConversation, unread_count: 0 }
        }
      }
    },

    applyRevokeSync(data: MessageRevokeResponse) {
      const conversationStore = useConversationStore()
      let targetConversationId = this.resolveConversationIdByUsers(data.sender_id, data.receiver_id)

      Object.entries(this.messagesByConversation).forEach(([conversationId, list]) => {
        if (!list.some((message) => message.id === data.message_id)) return

        targetConversationId = Number(conversationId)
        this.messagesByConversation[targetConversationId] = list.map((message) => {
          if (message.id !== data.message_id) return message
          return {
            ...message,
            content: '',
            status: 'revoked',
            update_time: new Date().toISOString(),
          }
        })
      })

      const updateConversationPreview = (conversationId: number) => {
        const conversation = conversationStore.conversationList.find((item) => item.id === conversationId)
        if (!conversation || conversation.last_message_id !== data.message_id) return

        conversationStore.upsertConversation({
          ...conversation,
          last_message: REVOKED_PREVIEW,
          update_time: new Date().toISOString(),
        })
      }

      if (targetConversationId) {
        updateConversationPreview(targetConversationId)
        if (conversationStore.currentConversation?.id === targetConversationId) {
          const current = conversationStore.currentConversation
          conversationStore.currentConversation = {
            ...current,
            last_message: current.last_message_id === data.message_id ? REVOKED_PREVIEW : current.last_message,
          }
        }
      } else {
        void conversationStore.fetchConversationList()
      }
    },
  },
})
