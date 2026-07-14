<script lang="ts" setup>
import { computed, nextTick, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { Image, MoreHorizontal, Paperclip, Phone, Search, Send, Smile } from 'lucide-vue-next'
import type { ConversationItemData } from '../../api/conversation'
import type { MessageResponse } from '../../api/message'
import { useMessageStore } from '../../stores/message'

const props = defineProps<{
  conversation: ConversationItemData | null
}>()

const messageStore = useMessageStore()
const { loadingHistory, messagesByConversation, socketError, socketStatus } = storeToRefs(messageStore)

const inputText = ref('')
const sending = ref(false)
const localError = ref('')
const listRef = ref<HTMLElement | null>(null)

const targetName = computed(() => {
  const user = props.conversation?.target_user
  if (!user) return ''
  return user.nickname || user.username || `用户 ${props.conversation?.target_id}`
})

const currentMessages = computed(() => {
  if (!props.conversation) return []
  return messagesByConversation.value[props.conversation.id] || []
})

const canSend = computed(() => {
  return Boolean(props.conversation) && inputText.value.trim() !== '' && !sending.value && messageStore.isSocketConnected
})

const socketLabel = computed(() => {
  switch (socketStatus.value) {
    case 'connected':
      return '在线'
    case 'connecting':
      return '连接中'
    case 'error':
      return '连接异常'
    case 'closed':
      return '离线'
    default:
      return '未连接'
  }
})

function isMine(message: MessageResponse) {
  return message.sender_id === messageStore.currentUserId
}

function formatTime(value: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

function displayContent(message: MessageResponse) {
  if (message.status === 'revoked') return '撤回了一条消息'
  if (message.type === 'image') return '[图片]'
  if (message.type === 'file') return '[文件]'
  if (message.type === 'voice') return '[语音]'
  return message.content
}

async function sendMessage() {
  if (!props.conversation || !canSend.value) return

  sending.value = true
  localError.value = ''
  try {
    messageStore.sendText(props.conversation.id, inputText.value.trim())
    inputText.value = ''
  } catch (err) {
    localError.value = err instanceof Error ? err.message : '消息发送失败'
  } finally {
    sending.value = false
  }
}

function handleKeydown(event: KeyboardEvent) {
  if ((event.ctrlKey || event.metaKey) && event.key === 'Enter') {
    event.preventDefault()
    void sendMessage()
  }
}

async function scrollToBottom() {
  await nextTick()
  if (listRef.value) {
    listRef.value.scrollTop = listRef.value.scrollHeight
  }
}

watch(
  () => props.conversation?.id,
  async (id) => {
    inputText.value = ''
    localError.value = ''
    if (id) {
      await messageStore.loadHistory(id).catch((err) => {
        localError.value = err instanceof Error ? err.message : '加载历史消息失败'
      })
      await scrollToBottom()
    }
  },
  { immediate: true },
)

watch(() => currentMessages.value.length, scrollToBottom)
</script>

<template>
  <section class="chat-panel">
    <header class="chat-header">
      <div class="chat-title-block">
        <h1>{{ props.conversation ? targetName : '选择会话' }}</h1>
        <span class="chat-ribbon">
          <template v-if="props.conversation">@{{ props.conversation.target_user.username }} · {{ socketLabel }}</template>
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

    <main ref="listRef" class="chat-body message-list-area">
      <div v-if="!props.conversation" class="chat-placeholder">
        选择一个会话开始聊天
      </div>

      <div v-else-if="loadingHistory" class="chat-empty-state">
        正在加载消息...
      </div>

      <div v-else-if="currentMessages.length === 0" class="chat-empty-state">
        还没有消息，打个招呼吧
      </div>

      <div v-else class="mock-message-list">
        <article
          v-for="message in currentMessages"
          :key="message.id"
          :class="['mock-message', { mine: isMine(message) }]"
        >
          <div class="message-bubble">{{ displayContent(message) }}</div>
          <time>{{ formatTime(message.create_time) }}</time>
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
          :disabled="!props.conversation || !messageStore.isSocketConnected"
          :placeholder="messageStore.isSocketConnected ? '输入消息，Ctrl + Enter 发送' : 'WebSocket 未连接，正在尝试重连'"
          @keydown="handleKeydown"
        ></textarea>

        <div class="chat-input-actions">
          <p v-if="localError || socketError" class="chat-send-error">{{ localError || socketError }}</p>
          <button :disabled="!canSend" type="button" @click="sendMessage">
            <Send :size="16" stroke-width="2" />
            <span>{{ sending ? '发送中' : '发送' }}</span>
          </button>
        </div>
      </section>
    </div>
  </section>
</template>
