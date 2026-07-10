<script lang="ts" setup>
import { computed, reactive, ref, watch } from 'vue'
import { Image, MoreHorizontal, Paperclip, Phone, Search, Send, Smile } from 'lucide-vue-next'
import type { ConversationItemData } from '../../api/conversation'

interface LocalMessage {
  id: string
  sender: 'me' | 'other'
  content: string
  time: string
}

const props = defineProps<{
  conversation: ConversationItemData | null
}>()

const inputText = ref('')
const localMessages = reactive<Record<number, LocalMessage[]>>({})

const targetName = computed(() => {
  const user = props.conversation?.target_user
  if (!user) return ''
  return user.nickname || user.username || `用户 ${props.conversation?.target_id}`
})

const currentMessages = computed(() => {
  if (!props.conversation) return []
  return localMessages[props.conversation.id] || []
})

const canSend = computed(() => Boolean(props.conversation) && inputText.value.trim() !== '')

function formatTime(date: Date) {
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

function sendMessage() {
  if (!props.conversation || !canSend.value) return

  if (!localMessages[props.conversation.id]) {
    localMessages[props.conversation.id] = []
  }

  localMessages[props.conversation.id].push({
    id: `${props.conversation.id}-${Date.now()}`,
    sender: 'me',
    content: inputText.value.trim(),
    time: formatTime(new Date()),
  })
  inputText.value = ''
}

function handleKeydown(event: KeyboardEvent) {
  if ((event.ctrlKey || event.metaKey) && event.key === 'Enter') {
    event.preventDefault()
    sendMessage()
  }
}

watch(
  () => props.conversation?.id,
  (id) => {
    inputText.value = ''
    if (id && !localMessages[id]) {
      localMessages[id] = []
    }
  },
)
</script>

<template>
  <section class="chat-panel">
    <header class="chat-header">
      <div class="chat-title-block">
        <h1>{{ props.conversation ? targetName : '选择会话' }}</h1>
        <span class="chat-ribbon">
          <template v-if="props.conversation">@{{ props.conversation.target_user.username }} · LightChat 会话</template>
          <template v-else>选择一个会话开始聊天</template>
        </span>
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

    <main class="chat-body message-list-area">
      <div v-if="!props.conversation" class="chat-placeholder">
        选择一个会话开始聊天
      </div>

      <div v-else-if="currentMessages.length === 0" class="chat-empty-state">
        还没有消息，打个招呼吧
      </div>

      <div v-else class="mock-message-list">
        <article
          v-for="message in currentMessages"
          :key="message.id"
          :class="['mock-message', { mine: message.sender === 'me' }]"
        >
          <div class="message-bubble">{{ message.content }}</div>
          <time>{{ message.time }}</time>
        </article>
      </div>
    </main>

    <div class="chat-input-wrapper">
      <section class="chat-input">
        <div class="chat-toolbar" aria-label="输入工具栏">
          <button title="表情" type="button">
            <Smile :size="20" stroke-width="1.8" />
          </button>
          <button title="文件" type="button">
            <Paperclip :size="20" stroke-width="1.8" />
          </button>
          <button title="图片" type="button">
            <Image :size="20" stroke-width="1.8" />
          </button>
        </div>

        <textarea
          v-model="inputText"
          :disabled="!props.conversation"
          placeholder="输入消息，Ctrl + Enter 发送"
          @keydown="handleKeydown"
        ></textarea>

        <div class="chat-input-actions">
          <button :disabled="!canSend" type="button" @click="sendMessage">
            <Send :size="16" stroke-width="2" />
            <span>发送</span>
          </button>
        </div>
      </section>
    </div>
  </section>
</template>
