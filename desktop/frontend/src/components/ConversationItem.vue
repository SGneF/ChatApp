<script lang="ts" setup>
import type { Conversation } from '../types/chat'

const props = defineProps<{
  conversation: Conversation
  active: boolean
}>()
</script>

<template>
  <button :class="['conversation-item', { active: props.active }]" type="button">
    <span class="conversation-avatar">{{ props.conversation.avatar }}</span>

    <span class="conversation-main">
      <span class="conversation-title-row">
        <strong>{{ props.conversation.name }}</strong>
        <time>{{ props.conversation.lastTime }}</time>
      </span>
      <span class="conversation-preview-row">
        <small>{{ props.conversation.lastMessage }}</small>
        <em v-if="props.conversation.unread > 0">{{ props.conversation.unread }}</em>
      </span>
    </span>
  </button>
</template>
