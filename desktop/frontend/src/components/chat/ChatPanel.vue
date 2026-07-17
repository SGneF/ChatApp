<script lang="ts" setup>
import { computed, nextTick, onUnmounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { Image, Mic, MoreHorizontal, Paperclip, Phone, RotateCcw, Search, Send, Smile } from 'lucide-vue-next'
import type { ConversationItemData } from '../../api/conversation'
import type { MessageResponse } from '../../api/message'
import {
  createFileMessageContent,
  formatFileSize,
  parseFileMessageContent,
  uploadFile,
  type FileMessagePayload,
  type UploadFileType,
} from '../../api/file'
import { useMessageStore } from '../../stores/message'

const props = defineProps<{
  conversation: ConversationItemData | null
  currentUserAvatar?: string
  currentUserName?: string
}>()

const messageStore = useMessageStore()
const { loadingHistory, messagesByConversation, socketError, socketStatus } = storeToRefs(messageStore)

const inputText = ref('')
const sending = ref(false)
const localError = ref('')
const uploadingType = ref<UploadFileType | ''>('')
const listRef = ref<HTMLElement | null>(null)
const imageInputRef = ref<HTMLInputElement | null>(null)
const fileInputRef = ref<HTMLInputElement | null>(null)
const voiceInputRef = ref<HTMLInputElement | null>(null)

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

const canUpload = computed(() => {
  return Boolean(props.conversation) && !uploadingType.value && messageStore.isSocketConnected
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

function initials(name: string) {
  return (name || '用户').trim().slice(0, 2).toUpperCase()
}

function isMine(message: MessageResponse) {
  return message.sender_id === messageStore.currentUserId
}

function avatarFor(message: MessageResponse) {
  if (isMine(message)) return props.currentUserAvatar || ''
  return props.conversation?.target_user.avatar || ''
}

function avatarNameFor(message: MessageResponse) {
  if (isMine(message)) return props.currentUserName || '我'
  return targetName.value || '对方'
}

function isReadableIncoming(message: MessageResponse) {
  return !isMine(message) && (message.status === 'normal' || message.status === 'sent')
}

function formatTime(value: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

function displayContent(message: MessageResponse) {
  if (message.type === 'image') return '[图片]'
  if (message.type === 'file') return '[文件]'
  if (message.type === 'voice') return '[语音]'
  return message.content
}

function filePayload(message: MessageResponse): FileMessagePayload | null {
  if (message.type === 'text') return null
  return parseFileMessageContent(message.content)
}

function fileName(message: MessageResponse) {
  return filePayload(message)?.file_name || displayContent(message)
}

function fileURL(message: MessageResponse) {
  return filePayload(message)?.url || ''
}

function fileSizeLabel(message: MessageResponse) {
  const payload = filePayload(message)
  return payload ? formatFileSize(payload.size) : ''
}

function revokeNotice(message: MessageResponse) {
  return isMine(message) ? '你撤回了一条消息' : '对方撤回了一条消息'
}

function deliveryLabel(message: MessageResponse) {
  if (!isMine(message) || message.status === 'revoked') return ''
  if (message.status === 'read') return '已读'
  return '已送达'
}

function canRevoke(message: MessageResponse) {
  if (!isMine(message) || message.status === 'revoked' || !messageStore.isSocketConnected) return false
  const createdAt = new Date(message.create_time).getTime()
  if (Number.isNaN(createdAt)) return false
  return Date.now() - createdAt <= 2 * 60 * 1000
}

function syncReadIfNeeded() {
  if (!props.conversation || !messageStore.isSocketConnected) return
  if (!messageStore.isActivelyReadingConversation(props.conversation.id)) return
  if (!currentMessages.value.some(isReadableIncoming)) return
  messageStore.markConversationRead(props.conversation.id)
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

function openFilePicker(type: UploadFileType) {
  if (!canUpload.value) return

  if (type === 'image') imageInputRef.value?.click()
  if (type === 'file') fileInputRef.value?.click()
  if (type === 'voice') voiceInputRef.value?.click()
}

async function handleFileSelected(event: Event, type: UploadFileType) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''

  if (!file || !props.conversation) return
  if (!messageStore.isSocketConnected) {
    localError.value = 'WebSocket 未连接，暂时不能发送文件'
    return
  }

  uploadingType.value = type
  localError.value = ''
  try {
    const uploaded = await uploadFile(file, type)
    messageStore.sendMessage(props.conversation.id, type, createFileMessageContent(uploaded))
  } catch (err) {
    localError.value = err instanceof Error ? err.message : '文件上传失败'
  } finally {
    uploadingType.value = ''
  }
}

function revokeMessage(message: MessageResponse) {
  if (!canRevoke(message)) return

  localError.value = ''
  try {
    messageStore.revokeMessage(message.id)
  } catch (err) {
    localError.value = err instanceof Error ? err.message : '撤回消息失败'
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
  async (id, oldId) => {
    if (oldId) {
      messageStore.clearActiveReadingConversation(oldId)
    }

    inputText.value = ''
    localError.value = ''

    if (!id) {
      messageStore.clearActiveReadingConversation()
      return
    }

    messageStore.setActiveReadingConversation(id)
    await messageStore.loadHistory(id).catch((err) => {
      localError.value = err instanceof Error ? err.message : '加载历史消息失败'
    })
    syncReadIfNeeded()
    await scrollToBottom()
  },
  { immediate: true },
)

watch(() => currentMessages.value.length, async () => {
  syncReadIfNeeded()
  await scrollToBottom()
})

onUnmounted(() => {
  messageStore.clearActiveReadingConversation(props.conversation?.id)
})
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
        <template v-for="message in currentMessages" :key="message.id">
          <article v-if="message.status === 'revoked'" class="message-system">
            <span>{{ revokeNotice(message) }}</span>
          </article>

          <article v-else :class="['mock-message', { mine: isMine(message) }]">
            <div class="message-avatar" :title="avatarNameFor(message)">
              <img v-if="avatarFor(message)" :alt="avatarNameFor(message)" :src="avatarFor(message)" />
              <span v-else>{{ initials(avatarNameFor(message)) }}</span>
            </div>

            <div class="message-content">
              <div v-if="message.type === 'image'" class="message-bubble media image-media">
                <a v-if="fileURL(message)" :href="fileURL(message)" rel="noreferrer" target="_blank">
                  <img class="message-image" :alt="fileName(message)" :src="fileURL(message)" />
                </a>
                <span v-else>{{ displayContent(message) }}</span>
              </div>

              <div v-else-if="message.type === 'voice'" class="message-bubble media voice-media">
                <div class="message-voice-card">
                  <Mic :size="18" stroke-width="2" />
                  <div class="message-file-info">
                    <strong>{{ fileName(message) }}</strong>
                    <span>{{ fileSizeLabel(message) }}</span>
                  </div>
                </div>
                <audio v-if="fileURL(message)" :src="fileURL(message)" controls></audio>
              </div>

              <div v-else-if="message.type === 'file'" class="message-bubble media file-media">
                <a v-if="fileURL(message)" class="message-file-card" :href="fileURL(message)" rel="noreferrer" target="_blank">
                  <span class="message-file-icon"><Paperclip :size="18" stroke-width="2" /></span>
                  <span class="message-file-info">
                    <strong>{{ fileName(message) }}</strong>
                    <small>{{ fileSizeLabel(message) }}</small>
                  </span>
                </a>
                <div v-else class="message-file-card disabled">
                  <span class="message-file-icon"><Paperclip :size="18" stroke-width="2" /></span>
                  <span class="message-file-info">
                    <strong>{{ fileName(message) }}</strong>
                    <small>{{ fileSizeLabel(message) }}</small>
                  </span>
                </div>
              </div>

              <div v-else class="message-bubble">{{ displayContent(message) }}</div>

              <div class="message-meta">
                <button
                  v-if="canRevoke(message)"
                  class="message-revoke"
                  title="撤回消息"
                  type="button"
                  @click="revokeMessage(message)"
                >
                  <RotateCcw :size="13" stroke-width="2" />
                  <span>撤回</span>
                </button>
                <time>{{ formatTime(message.create_time) }}</time>
                <span v-if="deliveryLabel(message)" class="message-status">{{ deliveryLabel(message) }}</span>
              </div>
            </div>
          </article>
        </template>
      </div>
    </main>

    <div class="chat-input-wrapper">
      <section class="chat-input">
        <div class="chat-toolbar" aria-label="输入工具栏">
          <button title="表情" type="button">
            <Smile :size="20" stroke-width="1.8" />
          </button>
          <button :disabled="!canUpload" title="发送文件" type="button" @click="openFilePicker('file')">
            <Paperclip :size="20" stroke-width="1.8" />
          </button>
          <button :disabled="!canUpload" title="发送图片" type="button" @click="openFilePicker('image')">
            <Image :size="20" stroke-width="1.8" />
          </button>
          <button :disabled="!canUpload" title="发送语音文件" type="button" @click="openFilePicker('voice')">
            <Mic :size="20" stroke-width="1.8" />
          </button>
          <span v-if="uploadingType" class="chat-uploading">上传中...</span>

          <input
            ref="fileInputRef"
            class="message-upload-input"
            type="file"
            @change="handleFileSelected($event, 'file')"
          />
          <input
            ref="imageInputRef"
            accept="image/*,.jpg,.jpeg,.png,.gif,.webp"
            class="message-upload-input"
            type="file"
            @change="handleFileSelected($event, 'image')"
          />
          <input
            ref="voiceInputRef"
            accept="audio/*,.mp3,.wav,.ogg,.m4a,.webm"
            class="message-upload-input"
            type="file"
            @change="handleFileSelected($event, 'voice')"
          />
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