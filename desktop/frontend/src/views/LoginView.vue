<script lang="ts" setup>
import { computed, reactive, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { Login } from '../../wailsjs/go/main/App'
import { saveToken } from '../services/session'

const router = useRouter()
const loading = ref(false)
const errorMessage = ref('')

const form = reactive({
  username: '',
  password: '',
})

const canLogin = computed(() => {
  return form.username.trim() !== '' && form.password !== '' && !loading.value
})

function getErrorMessage(err: unknown) {
  if (err instanceof Error) return err.message
  if (typeof err === 'string') return err
  return '登录失败'
}

async function submitLogin() {
  if (!canLogin.value) return

  loading.value = true
  errorMessage.value = ''

  try {
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
    <section class="auth-card" aria-label="登录">
      <header class="auth-header">
        <h1>LightChat</h1>
        <p>桌面即时通讯系统</p>
      </header>

      <form class="auth-form" @submit.prevent="submitLogin">
        <label>
          <span>用户名</span>
          <input v-model="form.username" autocomplete="username" autofocus type="text" />
        </label>

        <label>
          <span>密码</span>
          <input v-model="form.password" autocomplete="current-password" type="password" />
        </label>

        <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>

        <button class="auth-submit" :disabled="!canLogin" type="submit">
          {{ loading ? '登录中' : '登录' }}
        </button>
      </form>

      <p class="auth-switch">
        还没有账号？<RouterLink to="/register">去注册</RouterLink>
      </p>
    </section>
  </main>
</template>
