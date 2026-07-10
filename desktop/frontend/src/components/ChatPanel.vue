<script lang="ts" setup>
import { MoreHorizontal, Phone, Search } from 'lucide-vue-next'
import type { Conversation } from '../types/chat'
import ChatInput from './ChatInput.vue'
import MessageList from './MessageList.vue'

const props = defineProps<{
  conversation: Conversation | null
}>()

const emit = defineEmits<{
  (event: 'send', value: string): void
}>()
</script>

<template>
  <section class="chat-panel">
    <header class="chat-header">
      <div class="chat-title-block">
        <h1>{{ props.conversation?.name || '选择会话' }}</h1>
        <span class="chat-ribbon">清新会话</span>
      </div>
      <div class="chat-header-actions">
        <button title="搜索消息" type="button">
          <Search :size="19" stroke-width="2" />
        </button>
        <button title="语音通话" type="button">
          <Phone :size="19" stroke-width="2" />
        </button>
        <button title="更多" type="button">
          <MoreHorizontal :size="21" stroke-width="2" />
        </button>
      </div>
    </header>

    <MessageList :messages="props.conversation?.messages || []" />

    <div class="chat-input-wrapper">
      <ChatInput
        :disabled="!props.conversation"
        @send="emit('send', $event)"
      />
    </div>
  </section>
</template>

