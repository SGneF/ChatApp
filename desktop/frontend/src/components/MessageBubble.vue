<script lang="ts" setup>
import type { ChatMessage } from '../types/chat'

const props = defineProps<{
  message: ChatMessage
}>()
</script>

<template>
  <article :class="['message-item', props.message.sender === 'me' ? 'mine' : 'other']">
    <div class="message-avatar">{{ props.message.sender === 'me' ? '我' : '对' }}</div>
    <div class="message-content">
      <span class="message-tag">{{ props.message.sender === 'me' ? '我' : '对方' }}</span>
      <div class="message-bubble">
        {{ props.message.content }}
      </div>
      <time>{{ props.message.time }}</time>
    </div>
  </article>
</template>
