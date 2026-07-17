<script lang="ts" setup>
import { computed, reactive, ref, watch } from 'vue'
import { Save, X } from 'lucide-vue-next'
import { UpdateProfile } from '../../wailsjs/go/main/App'
import { getToken } from '../services/session'

interface UserProfile {
  id: number
  username: string
  nickname: string
  avatar: string
  signature: string
  create_time: string
  update_time: string
}

const props = defineProps<{
  user: UserProfile
}>()

const emit = defineEmits<{
  (event: 'close'): void
  (event: 'saved', user: UserProfile): void
}>()

const form = reactive({
  nickname: '',
  avatar: '',
  signature: '',
})
const saving = ref(false)
const error = ref('')

const previewName = computed(() => form.nickname.trim() || props.user.nickname || props.user.username)
const previewAvatar = computed(() => form.avatar.trim())

function initials(name: string) {
  return (name || '用户').trim().slice(0, 2).toUpperCase()
}

function resetForm() {
  form.nickname = props.user.nickname || props.user.username || ''
  form.avatar = props.user.avatar || ''
  form.signature = props.user.signature || ''
  error.value = ''
}

function getErrorMessage(err: unknown) {
  if (err instanceof Error) return err.message
  if (typeof err === 'string') return err
  return '保存失败'
}

async function saveProfile() {
  const nickname = form.nickname.trim()
  if (!nickname) {
    error.value = '昵称不能为空'
    return
  }

  const token = getToken()
  if (!token) {
    error.value = '登录状态已失效，请重新登录'
    return
  }

  saving.value = true
  error.value = ''
  try {
    const updated = await UpdateProfile(token, {
      nickname,
      avatar: form.avatar.trim(),
      signature: form.signature.trim(),
    })
    emit('saved', updated as UserProfile)
    emit('close')
  } catch (err) {
    error.value = getErrorMessage(err)
  } finally {
    saving.value = false
  }
}

watch(() => props.user, resetForm, { immediate: true })
</script>

<template>
  <div class="modal-backdrop profile-backdrop" @click.self="emit('close')">
    <section class="profile-modal" aria-labelledby="profile-modal-title" role="dialog" aria-modal="true">
      <header>
        <div>
          <h2 id="profile-modal-title">个人资料</h2>
          <p>编辑头像、昵称和个性签名</p>
        </div>
        <button title="关闭" type="button" @click="emit('close')">
          <X :size="18" stroke-width="2" />
        </button>
      </header>

      <form class="profile-form" @submit.prevent="saveProfile">
        <div class="profile-preview">
          <div class="profile-avatar-preview">
            <img v-if="previewAvatar" :alt="previewName" :src="previewAvatar" />
            <span v-else>{{ initials(previewName) }}</span>
          </div>
          <div class="profile-preview-copy">
            <strong>{{ previewName }}</strong>
            <small>@{{ props.user.username }}</small>
          </div>
        </div>

        <label>
          <span>昵称</span>
          <input v-model="form.nickname" maxlength="32" placeholder="请输入昵称" type="text" />
        </label>

        <label>
          <span>头像地址</span>
          <input v-model="form.avatar" maxlength="255" placeholder="请输入图片 URL" type="text" />
        </label>

        <label>
          <span>个性签名</span>
          <textarea v-model="form.signature" maxlength="255" placeholder="写一句个性签名"></textarea>
        </label>

        <p v-if="error" class="modal-message error">{{ error }}</p>

        <div class="profile-actions">
          <button class="ghost" :disabled="saving" type="button" @click="resetForm">重置</button>
          <button class="primary" :disabled="saving" type="submit">
            <Save :size="16" stroke-width="2" />
            <span>{{ saving ? '保存中' : '保存资料' }}</span>
          </button>
        </div>
      </form>
    </section>
  </div>
</template>
