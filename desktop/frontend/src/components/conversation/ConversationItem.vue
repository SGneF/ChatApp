<script lang="ts" setup>
import { MoreVertical, Pin, PinOff, Trash2 } from 'lucide-vue-next'
import type { ConversationItemData } from '../../api/conversation'

const props = defineProps<{
  conversation: ConversationItemData
  active: boolean
}>()

const emit = defineEmits<{
  (event: 'select', value: ConversationItemData): void
  (event: 'toggle-top', value: ConversationItemData): void
  (event: 'delete', value: ConversationItemData): void
}>()

function displayName(conversation: ConversationItemData) {
  return conversation.target_user.nickname || conversation.target_user.username || `用户 ${conversation.target_id}`
}

function initials(value: string) {
  return (value || 'LC').slice(0, 2).toUpperCase()
}

function previewText(value: string) {
  return value || '还没有消息'
}

function formatTime(value: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''

  const now = new Date()
  const isToday = date.toDateString() === now.toDateString()
  if (isToday) {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }
  return date.toLocaleDateString([], { month: '2-digit', day: '2-digit' })
}
</script>

<template>
  <article :class="['conversation-item', { active: props.active }]" @click="emit('select', props.conversation)">
    <span class="conversation-avatar">
      <img v-if="props.conversation.target_user.avatar" :alt="displayName(props.conversation)" :src="props.conversation.target_user.avatar" />
      <span v-else>{{ initials(displayName(props.conversation)) }}</span>
    </span>

    <span class="conversation-main">
      <span class="conversation-title-row">
        <strong>
          <Pin v-if="props.conversation.is_top" :size="12" stroke-width="2.4" />
          <span>{{ displayName(props.conversation) }}</span>
        </strong>
        <time>{{ formatTime(props.conversation.update_time) }}</time>
      </span>
      <span class="conversation-preview-row">
        <small>{{ previewText(props.conversation.last_message) }}</small>
        <em v-if="props.conversation.unread_count > 0">{{ props.conversation.unread_count }}</em>
      </span>
    </span>

    <span class="conversation-actions" @click.stop>
      <button :title="props.conversation.is_top ? '取消置顶' : '置顶'" type="button" @click="emit('toggle-top', props.conversation)">
        <PinOff v-if="props.conversation.is_top" :size="14" stroke-width="2" />
        <Pin v-else :size="14" stroke-width="2" />
      </button>
      <button title="删除会话" type="button" @click="emit('delete', props.conversation)">
        <Trash2 :size="14" stroke-width="2" />
      </button>
      <MoreVertical :size="14" stroke-width="2" />
    </span>
  </article>
</template>
