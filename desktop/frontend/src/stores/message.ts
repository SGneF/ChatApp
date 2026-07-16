import { defineStore } from 'pinia'
import { getMessageHistory, type MessageResponse } from '../api/message'
import { getToken } from '../services/session'
import {
  chatSocket,
  type ChatSocketEvent,
  type ChatSocketOfflineSyncData,
  type ChatSocketStatus,
} from '../services/chatSocket'
import { useConversationStore } from './conversation'

function sortMessages(list: MessageResponse[]) {
  return [...list].sort((a, b) => {
    const diff = new Date(a.create_time).getTime() - new Date(b.create_time).getTime()
    if (diff !== 0) return diff
    return a.id - b.id
  })
}

export const useMessageStore = defineStore('message', {
  state: () => ({
    currentUserId: 0,
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

    sendText(conversationId: number, content: string) {
      this.socketError = ''
      chatSocket.sendChatMessage(conversationId, content.trim(), 'text')
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
        this.socketError = event.data.message || '消息发送失败'
        return
      }

      if (event.type === 'offline_sync') {
        void this.handleOfflineSync(event.data)
        return
      }

      this.appendMessage(event.data)
      this.updateConversationByMessage(event.data)
    },

    async handleOfflineSync(data: ChatSocketOfflineSyncData) {
      const conversationStore = useConversationStore()
      await conversationStore.fetchConversationList().catch(() => undefined)

      const offlineMessages = data.conversations.flatMap((item) => item.messages || [])
      offlineMessages.forEach((message) => this.appendMessage(message))

      const currentConversationId = conversationStore.currentConversation?.id
      if (!currentConversationId) return

      const currentHasOfflineMessages = data.conversations.some(
        (item) => item.conversation_id === currentConversationId,
      )
      if (!currentHasOfflineMessages) return

      await this.loadHistory(currentConversationId).catch(() => undefined)
      await conversationStore.markRead(currentConversationId).catch(() => undefined)
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
      const isCurrent = conversationStore.currentConversation?.id === conversation.id

      conversationStore.upsertConversation({
        ...conversation,
        last_message_id: message.id,
        last_message: message.status === 'revoked' ? '撤回了一条消息' : message.content,
        unread_count: isIncoming && !isCurrent ? conversation.unread_count + 1 : 0,
        update_time: now,
      })

      if (isCurrent) {
        conversationStore.currentConversation = {
          ...conversationStore.currentConversation!,
          last_message_id: message.id,
          last_message: message.status === 'revoked' ? '撤回了一条消息' : message.content,
          unread_count: 0,
          update_time: now,
        }
        void conversationStore.markRead(conversation.id).catch(() => undefined)
      }
    },
  },
})