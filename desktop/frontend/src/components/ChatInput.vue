<script lang="ts" setup>
import { computed, ref } from 'vue'
import { Image, Paperclip, Send, Smile } from 'lucide-vue-next'

const props = defineProps<{
  disabled?: boolean
}>()

const emit = defineEmits<{
  (event: 'send', value: string): void
}>()

const text = ref('')
const canSend = computed(() => text.value.trim() !== '' && !props.disabled)

function sendMessage() {
  if (!canSend.value) return

  emit('send', text.value.trim())
  text.value = ''
}

function handleKeydown(event: KeyboardEvent) {
  if ((event.ctrlKey || event.metaKey) && event.key === 'Enter') {
    event.preventDefault()
    sendMessage()
  }
}
</script>

<template>
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
      v-model="text"
      :disabled="props.disabled"
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
</template>
