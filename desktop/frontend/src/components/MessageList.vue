<script lang="ts" setup>
import { nextTick, onMounted, ref, watch } from 'vue'
import type { ChatMessage } from '../types/chat'
import MessageBubble from './MessageBubble.vue'

const props = defineProps<{
  messages: ChatMessage[]
}>()

const listRef = ref<HTMLElement | null>(null)

async function scrollToBottom() {
  await nextTick()
  if (!listRef.value) return
  listRef.value.scrollTop = listRef.value.scrollHeight
}

watch(() => props.messages.length, scrollToBottom)
onMounted(scrollToBottom)
</script>

<template>
  <section ref="listRef" class="chat-body message-list-area" aria-label="消息列表">
    <MessageBubble
      v-for="message in props.messages"
      :key="message.id"
      :message="message"
    />
  </section>
</template>
