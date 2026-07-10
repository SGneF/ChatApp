<script lang="ts" setup>
import { computed } from 'vue'
import type { Conversation, NavKey } from '../types/chat'
import Sidebar from './Sidebar.vue'
import ConversationList from './ConversationList.vue'
import ChatPanel from './ChatPanel.vue'
import FriendView from './FriendView.vue'

const props = defineProps<{
  userName: string
  userAvatar?: string
  activeNav: NavKey
  conversations: Conversation[]
  activeConversationId: string
}>()

const emit = defineEmits<{
  (event: 'update:activeNav', value: NavKey): void
  (event: 'select-conversation', value: string): void
  (event: 'send-message', value: string): void
  (event: 'logout'): void
}>()

const activeConversation = computed(() => {
  return props.conversations.find((item) => item.id === props.activeConversationId) || null
})
</script>

<template>
  <main class="main-layout">
    <Sidebar
      :active-nav="activeNav"
      :user-avatar="userAvatar"
      :user-name="userName"
      @logout="emit('logout')"
      @update:active-nav="emit('update:activeNav', $event)"
    />

    <FriendView v-if="activeNav === 'contacts'" />

    <template v-else>
      <ConversationList
        :active-conversation-id="activeConversationId"
        :conversations="conversations"
        @select="emit('select-conversation', $event)"
      />

      <ChatPanel
        :conversation="activeConversation"
        @send="emit('send-message', $event)"
      />
    </template>
  </main>
</template>
