<script lang="ts" setup>
import { computed, reactive, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { Login, Register } from '../../wailsjs/go/main/App'
import { saveToken } from '../services/session'

const router = useRouter()
const loading = ref(false)
const errorMessage = ref('')

const form = reactive({
  username: '',
  nickname: '',
  password: '',
  confirmPassword: '',
})

const canRegister = computed(() => {
  return (
    form.username.trim() !== '' &&
    form.password !== '' &&
    form.confirmPassword !== '' &&
    !loading.value
  )
})

function getErrorMessage(err: unknown) {
  if (err instanceof Error) return err.message
  if (typeof err === 'string') return err
  return '注册失败'
}

async function submitRegister() {
  if (!canRegister.value) return

  if (form.password !== form.confirmPassword) {
    errorMessage.value = '两次输入的密码不一致'
    return
  }

  loading.value = true
  errorMessage.value = ''

  try {
    await Register({
      username: form.username.trim(),
      nickname: form.nickname.trim(),
      password: form.password,
    })

    const result = await Login({
      username: form.username.trim(),
      password: form.password,
    })

    saveToken(result.token)
    await router.push({ name: 'home' })
  } catch (err) {
    errorMessage.value = getErrorMessage(err)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <main class="auth-page">
    <section class="auth-card" aria-label="注册">
      <header class="auth-header">
        <h1>LightChat</h1>
        <p>桌面即时通讯系统</p>
      </header>

      <form class="auth-form" @submit.prevent="submitRegister">
        <label>
          <span>用户名</span>
          <input v-model="form.username" autocomplete="username" autofocus type="text" />
        </label>

        <label>
          <span>昵称</span>
          <input v-model="form.nickname" autocomplete="nickname" type="text" />
        </label>

        <label>
          <span>密码</span>
          <input v-model="form.password" autocomplete="new-password" type="password" />
        </label>

        <label>
          <span>确认密码</span>
          <input v-model="form.confirmPassword" autocomplete="new-password" type="password" />
        </label>

        <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>

        <button class="auth-submit" :disabled="!canRegister" type="submit">
          {{ loading ? '注册中' : '注册' }}
        </button>
      </form>

      <p class="auth-switch">
        已有账号？<RouterLink to="/login">去登录</RouterLink>
      </p>
    </section>
  </main>
</template>
