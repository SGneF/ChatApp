<script lang="ts" setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useRouter } from 'vue-router'
import { GetUserInfo } from '../../wailsjs/go/main/App'
import ChatPanel from '../components/chat/ChatPanel.vue'
import ConversationList from '../components/conversation/ConversationList.vue'
import FriendView from '../components/FriendView.vue'
import ProfileModal from '../components/ProfileModal.vue'
import Sidebar from '../components/Sidebar.vue'
import { clearToken, getToken } from '../services/session'
import { useConversationStore } from '../stores/conversation'
import { useMessageStore } from '../stores/message'
import type { ConversationItemData } from '../api/conversation'
import type { NavKey } from '../types/chat'

interface UserResponse {
  id: number
  username: string
  nickname: string
  avatar: string
  signature: string
  create_time: string
  update_time: string
}

const router = useRouter()
const conversationStore = useConversationStore()
const messageStore = useMessageStore()
const { conversationList, currentConversation, loading } = storeToRefs(conversationStore)

const user = ref<UserResponse | null>(null)
const activeNav = ref<NavKey>('messages')
const pageError = ref('')
const profileVisible = ref(false)

const userName = computed(() => {
  if (!user.value) return 'LightChat'
  return user.value.nickname || user.value.username
})

const userAvatar = computed(() => user.value?.avatar || '')

function getErrorMessage(err: unknown) {
  if (err instanceof Error) return err.message
  if (typeof err === 'string') return err
  return '操作失败'
}

async function loadCurrentUser() {
  const token = getToken()
  if (!token) {
    await router.replace({ name: 'login' })
    return
  }

  try {
    user.value = await GetUserInfo(token)
    messageStore.setCurrentUserId(user.value.id)
    messageStore.connect()
  } catch {
    clearToken()
    await router.replace({ name: 'login' })
    return
  }

  try {
    await conversationStore.fetchConversationList()
  } catch (err) {
    pageError.value = getErrorMessage(err)
  }
}

async function selectConversation(conversation: ConversationItemData) {
  pageError.value = ''
  try {
    await conversationStore.selectConversation(conversation)
  } catch (err) {
    pageError.value = getErrorMessage(err)
  }
}

async function removeConversation(conversation: ConversationItemData) {
  if (!window.confirm(`确认删除与「${conversation.target_user.nickname || conversation.target_user.username}」的会话吗？`)) return

  pageError.value = ''
  try {
    await conversationStore.removeConversation(conversation.id)
  } catch (err) {
    pageError.value = getErrorMessage(err)
  }
}

async function toggleConversationTop(conversation: ConversationItemData) {
  pageError.value = ''
  try {
    await conversationStore.toggleTop(conversation)
  } catch (err) {
    pageError.value = getErrorMessage(err)
  }
}

function showMessages() {
  pageError.value = ''
  activeNav.value = 'messages'
}

function updateActiveNav(value: NavKey) {
  activeNav.value = value
  pageError.value = ''

  if (value === 'messages') {
    void conversationStore.fetchConversationList()
  }
}

function openProfile() {
  if (!user.value) return
  pageError.value = ''
  profileVisible.value = true
}

function handleProfileSaved(nextUser: UserResponse) {
  user.value = nextUser
  messageStore.setCurrentUserId(nextUser.id)
  pageError.value = ''
}

async function logout() {
  messageStore.disconnect()
  clearToken()
  await router.replace({ name: 'login' })
}

onMounted(loadCurrentUser)
onUnmounted(() => {
  messageStore.disconnect()
})
</script>

<template>
  <main class="main-layout">
    <Sidebar
      :active-nav="activeNav"
      :user-avatar="userAvatar"
      :user-name="userName"
      @logout="logout"
      @open-profile="openProfile"
      @update:active-nav="updateActiveNav"
    />

    <FriendView v-if="activeNav === 'contacts'" @conversation-opened="showMessages" />

    <template v-else>
      <ConversationList
        :active-conversation-id="currentConversation?.id"
        :conversations="conversationList"
        :loading="loading"
        @delete="removeConversation"
        @select="selectConversation"
        @toggle-top="toggleConversationTop"
      />

      <ChatPanel
        :conversation="currentConversation"
        :current-user-avatar="userAvatar"
        :current-user-name="userName"
      />
    </template>

    <ProfileModal
      v-if="profileVisible && user"
      :user="user"
      @close="profileVisible = false"
      @saved="handleProfileSaved"
    />

    <div v-if="pageError" class="home-toast error">{{ pageError }}</div>
  </main>
</template>
