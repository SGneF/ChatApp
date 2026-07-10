import { defineStore } from 'pinia'
import {
  createSingleConversation,
  deleteConversation,
  getConversationDetail,
  getConversationList,
  markConversationRead,
  setConversationTop,
  type ConversationItemData,
} from '../api/conversation'

function sortConversations(list: ConversationItemData[]) {
  return [...list].sort((a, b) => {
    if (a.is_top !== b.is_top) return a.is_top ? -1 : 1
    return new Date(b.update_time).getTime() - new Date(a.update_time).getTime()
  })
}

export const useConversationStore = defineStore('conversation', {
  state: () => ({
    conversationList: [] as ConversationItemData[],
    currentConversation: null as ConversationItemData | null,
    loading: false,
  }),

  actions: {
    upsertConversation(conversation: ConversationItemData) {
      const index = this.conversationList.findIndex((item) => item.id === conversation.id)
      if (index >= 0) {
        this.conversationList.splice(index, 1, conversation)
      } else {
        this.conversationList.unshift(conversation)
      }
      this.conversationList = sortConversations(this.conversationList)
    },

    async fetchConversationList() {
      this.loading = true
      try {
        this.conversationList = sortConversations(await getConversationList())
        if (this.currentConversation) {
          const latest = this.conversationList.find((item) => item.id === this.currentConversation?.id)
          this.currentConversation = latest || null
        }
      } finally {
        this.loading = false
      }
    },

    async createOrOpenSingleConversation(targetId: number) {
      this.loading = true
      try {
        const conversation = await createSingleConversation(targetId)
        this.upsertConversation(conversation)
        this.currentConversation = conversation
        await this.markRead(conversation.id).catch(() => undefined)
        return conversation
      } finally {
        this.loading = false
      }
    },

    async selectConversation(conversation: ConversationItemData) {
      this.currentConversation = conversation
      try {
        const detail = await getConversationDetail(conversation.id)
        this.upsertConversation(detail)
        this.currentConversation = detail
      } catch {
        this.currentConversation = conversation
      }
      await this.markRead(conversation.id).catch(() => undefined)
    },

    async removeConversation(conversationId: number) {
      await deleteConversation(conversationId)
      this.conversationList = this.conversationList.filter((item) => item.id !== conversationId)
      if (this.currentConversation?.id === conversationId) {
        this.currentConversation = null
      }
    },

    async markRead(conversationId: number) {
      await markConversationRead(conversationId)
      const conversation = this.conversationList.find((item) => item.id === conversationId)
      if (conversation) conversation.unread_count = 0
      if (this.currentConversation?.id === conversationId) {
        this.currentConversation = { ...this.currentConversation, unread_count: 0 }
      }
    },

    async toggleTop(conversation: ConversationItemData) {
      const nextTop = !conversation.is_top
      await setConversationTop(conversation.id, nextTop)
      this.upsertConversation({ ...conversation, is_top: nextTop, update_time: new Date().toISOString() })
      if (this.currentConversation?.id === conversation.id) {
        this.currentConversation = { ...this.currentConversation, is_top: nextTop }
      }
    },
  },
})


