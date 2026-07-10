<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useConversationStore } from '../stores/conversation'
import {
  Check,
  ChevronDown,
  MessageCircle,
  RefreshCw,
  Search,
  Trash2,
  Users,
  X,
} from 'lucide-vue-next'
import {
  AcceptFriendRequest,
  DeleteFriend,
  ListFriendRequests,
  ListFriends,
  RejectFriendRequest,
} from '../../wailsjs/go/main/App'
import { getToken } from '../services/session'

interface FriendRequestItem {
  id: number
  from_user_id: number
  from_username: string
  from_nickname: string
  from_avatar: string
  remark: string
  status: string
  create_time: string
}

interface FriendItem {
  id: number
  username: string
  nickname: string
  avatar: string
  signature: string
  remark: string
}

interface GroupItem {
  id: string
  name: string
  avatar: string
  description: string
  meta: string
}

type DetailTarget =
  | { type: 'group'; id: string }
  | { type: 'request'; id: number }
  | { type: 'friend'; id: number }

const emit = defineEmits<{
  (event: 'conversation-opened'): void
}>()

const conversationStore = useConversationStore()

const groups: GroupItem[] = [
  { id: 'team', name: '产品研发群', avatar: '群', description: '桌面端功能协作与需求同步', meta: '群聊' },
  { id: 'ai', name: 'AI 助手', avatar: 'AI', description: '摘要整理、待办提取与智能问答', meta: '智能助手' },
]

const friends = ref<FriendItem[]>([])
const requests = ref<FriendRequestItem[]>([])
const keyword = ref('')
const errorMessage = ref('')
const successMessage = ref('')
const loadingFriends = ref(false)
const loadingRequests = ref(false)
const handlingRequestId = ref<number | null>(null)
const deletingFriendId = ref<number | null>(null)
const openingConversationFriendId = ref<number | null>(null)
const selected = ref<DetailTarget>({ type: 'group', id: groups[0].id })

const collapsed = reactive({
  groups: false,
  requests: false,
  friends: false,
})

const filteredGroups = computed(() => {
  const value = keyword.value.trim().toLowerCase()
  if (!value) return groups
  return groups.filter((item) => item.name.toLowerCase().includes(value) || item.description.toLowerCase().includes(value))
})

const filteredRequests = computed(() => {
  const value = keyword.value.trim().toLowerCase()
  if (!value) return requests.value
  return requests.value.filter((item) => {
    return displayRequestName(item).toLowerCase().includes(value) || item.from_username.toLowerCase().includes(value)
  })
})

const filteredFriends = computed(() => {
  const value = keyword.value.trim().toLowerCase()
  if (!value) return friends.value
  return friends.value.filter((item) => {
    return displayFriendName(item).toLowerCase().includes(value) || item.username.toLowerCase().includes(value)
  })
})

const selectedGroup = computed(() => {
  if (selected.value.type !== 'group') return null
  return groups.find((item) => item.id === selected.value.id) || null
})

const selectedRequest = computed(() => {
  if (selected.value.type !== 'request') return null
  return requests.value.find((item) => item.id === selected.value.id) || null
})

const selectedFriend = computed(() => {
  if (selected.value.type !== 'friend') return null
  return friends.value.find((item) => item.id === selected.value.id) || null
})

function getErrorMessage(err: unknown) {
  if (err instanceof Error) return err.message
  if (typeof err === 'string') return err
  return '操作失败'
}

function requireToken() {
  const token = getToken()
  if (!token) {
    errorMessage.value = '登录已过期，请重新登录'
    return ''
  }
  return token
}

function displayFriendName(friend: FriendItem) {
  return friend.nickname || friend.username || `用户 ${friend.id}`
}

function displayRequestName(request: FriendRequestItem) {
  return request.from_nickname || request.from_username || `用户 ${request.from_user_id}`
}

function initials(value: string) {
  const text = value.trim()
  return (text || '用户').slice(0, 2).toUpperCase()
}

function formatTime(value: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString([], {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function clearNotice() {
  errorMessage.value = ''
  successMessage.value = ''
}

function selectGroup(group: GroupItem) {
  selected.value = { type: 'group', id: group.id }
}

function selectRequest(request: FriendRequestItem) {
  selected.value = { type: 'request', id: request.id }
}

function selectFriend(friend: FriendItem) {
  selected.value = { type: 'friend', id: friend.id }
}

function isSelected(type: DetailTarget['type'], id: string | number) {
  return selected.value.type === type && selected.value.id === id
}

function toggleSection(key: keyof typeof collapsed) {
  collapsed[key] = !collapsed[key]
}

function ensureSelection() {
  if (selected.value.type === 'friend' && friends.value.some((item) => item.id === selected.value.id)) return
  if (selected.value.type === 'request' && requests.value.some((item) => item.id === selected.value.id)) return
  if (selected.value.type === 'group' && groups.some((item) => item.id === selected.value.id)) return

  if (groups[0]) {
    selected.value = { type: 'group', id: groups[0].id }
  }
}

async function loadFriends() {
  const token = requireToken()
  if (!token) return

  loadingFriends.value = true
  try {
    friends.value = (await ListFriends(token)) as FriendItem[]
    ensureSelection()
  } catch (err) {
    errorMessage.value = getErrorMessage(err)
  } finally {
    loadingFriends.value = false
  }
}

async function loadRequests() {
  const token = requireToken()
  if (!token) return

  loadingRequests.value = true
  try {
    requests.value = (await ListFriendRequests(token)) as FriendRequestItem[]
    ensureSelection()
  } catch (err) {
    errorMessage.value = getErrorMessage(err)
  } finally {
    loadingRequests.value = false
  }
}

async function refreshAll() {
  clearNotice()
  await Promise.allSettled([loadFriends(), loadRequests()])
  ensureSelection()
}

async function acceptRequest(request: FriendRequestItem) {
  const token = requireToken()
  if (!token) return

  clearNotice()
  handlingRequestId.value = request.id
  try {
    await AcceptFriendRequest(token, { request_id: request.id })
    successMessage.value = '已同意好友申请'
    await refreshAll()
  } catch (err) {
    errorMessage.value = getErrorMessage(err)
  } finally {
    handlingRequestId.value = null
  }
}

async function rejectRequest(request: FriendRequestItem) {
  const token = requireToken()
  if (!token) return

  clearNotice()
  handlingRequestId.value = request.id
  try {
    await RejectFriendRequest(token, { request_id: request.id })
    successMessage.value = '已拒绝好友申请'
    await loadRequests()
    ensureSelection()
  } catch (err) {
    errorMessage.value = getErrorMessage(err)
  } finally {
    handlingRequestId.value = null
  }
}

async function removeFriend(friend: FriendItem) {
  const token = requireToken()
  if (!token) return
  if (!window.confirm(`确认删除好友「${displayFriendName(friend)}」吗？`)) return

  clearNotice()
  deletingFriendId.value = friend.id
  try {
    await DeleteFriend(token, friend.id)
    successMessage.value = '好友已删除'
    await loadFriends()
    ensureSelection()
  } catch (err) {
    errorMessage.value = getErrorMessage(err)
  } finally {
    deletingFriendId.value = null
  }
}

async function startChat(friend?: FriendItem) {
  if (!friend) {
    successMessage.value = '请先在好友列表选择一个好友，群聊稍后接入'
    errorMessage.value = ''
    return
  }

  clearNotice()
  openingConversationFriendId.value = friend.id
  try {
    await conversationStore.createOrOpenSingleConversation(friend.id)
    emit('conversation-opened')
  } catch (err) {
    errorMessage.value = getErrorMessage(err)
  } finally {
    openingConversationFriendId.value = null
  }
}

onMounted(refreshAll)
</script>

<template>
  <section class="friend-view contact-layout">
    <aside class="contact-list-pane">
      <div class="contact-search-row">
        <label class="contact-search">
          <Search :size="16" stroke-width="2" />
          <input v-model="keyword" autocomplete="off" placeholder="搜索" type="text" />
        </label>
        <button class="contact-refresh" :disabled="loadingFriends || loadingRequests" title="刷新" type="button" @click="refreshAll">
          <RefreshCw :size="16" stroke-width="2" />
        </button>
      </div>

      <div v-if="errorMessage" class="friend-alert error compact">{{ errorMessage }}</div>
      <div v-if="successMessage" class="friend-alert success compact">{{ successMessage }}</div>

      <div class="contact-sections">
        <section class="contact-section">
          <button class="contact-section-title" type="button" @click="toggleSection('groups')">
            <span>群聊</span>
            <small>{{ filteredGroups.length }}</small>
            <ChevronDown :class="['section-chevron', { collapsed: collapsed.groups }]" :size="16" stroke-width="2" />
          </button>
          <div v-show="!collapsed.groups" class="contact-section-body">
            <button
              v-for="group in filteredGroups"
              :key="group.id"
              :class="['contact-row', { active: isSelected('group', group.id) }]"
              type="button"
              @click="selectGroup(group)"
            >
              <span class="contact-avatar group">{{ group.avatar }}</span>
              <span class="contact-copy">
                <strong>{{ group.name }}</strong>
                <small>{{ group.meta }}</small>
              </span>
            </button>
          </div>
        </section>

        <section class="contact-section">
          <button class="contact-section-title" type="button" @click="toggleSection('requests')">
            <span>好友申请</span>
            <small>{{ filteredRequests.length }}</small>
            <ChevronDown :class="['section-chevron', { collapsed: collapsed.requests }]" :size="16" stroke-width="2" />
          </button>
          <div v-show="!collapsed.requests" class="contact-section-body">
            <div v-if="filteredRequests.length === 0" class="contact-empty">暂无申请</div>
            <button
              v-for="request in filteredRequests"
              :key="request.id"
              :class="['contact-row', { active: isSelected('request', request.id) }]"
              type="button"
              @click="selectRequest(request)"
            >
              <span class="contact-avatar request">
                <img v-if="request.from_avatar" :alt="displayRequestName(request)" :src="request.from_avatar" />
                <span v-else>{{ initials(displayRequestName(request)) }}</span>
              </span>
              <span class="contact-copy">
                <strong>{{ displayRequestName(request) }}</strong>
                <small>{{ request.remark || '请求添加你为好友' }}</small>
              </span>
            </button>
          </div>
        </section>

        <section class="contact-section">
          <button class="contact-section-title" type="button" @click="toggleSection('friends')">
            <span>好友列表</span>
            <small>{{ filteredFriends.length }}</small>
            <ChevronDown :class="['section-chevron', { collapsed: collapsed.friends }]" :size="16" stroke-width="2" />
          </button>
          <div v-show="!collapsed.friends" class="contact-section-body">
            <div v-if="filteredFriends.length === 0" class="contact-empty">暂无好友</div>
            <button
              v-for="friend in filteredFriends"
              :key="friend.id"
              :class="['contact-row', { active: isSelected('friend', friend.id) }]"
              type="button"
              @click="selectFriend(friend)"
            >
              <span class="contact-avatar">
                <img v-if="friend.avatar" :alt="displayFriendName(friend)" :src="friend.avatar" />
                <span v-else>{{ initials(displayFriendName(friend)) }}</span>
              </span>
              <span class="contact-copy">
                <strong>{{ displayFriendName(friend) }}</strong>
                <small>@{{ friend.username }}</small>
              </span>
            </button>
          </div>
        </section>
      </div>
    </aside>

    <main class="contact-detail-pane">
      <template v-if="selectedGroup">
        <div class="contact-detail-card">
          <div class="detail-avatar group">{{ selectedGroup.avatar }}</div>
          <h1>{{ selectedGroup.name }}</h1>
          <p>{{ selectedGroup.description }}</p>
          <small>{{ selectedGroup.meta }}</small>
          <div class="detail-actions">
            <button class="primary" type="button" @click="startChat()">
              <MessageCircle :size="16" stroke-width="2" />
              <span>发送消息</span>
            </button>
          </div>
        </div>
      </template>

      <template v-else-if="selectedRequest">
        <div class="contact-detail-card">
          <div class="detail-avatar request">
            <img v-if="selectedRequest.from_avatar" :alt="displayRequestName(selectedRequest)" :src="selectedRequest.from_avatar" />
            <span v-else>{{ initials(displayRequestName(selectedRequest)) }}</span>
          </div>
          <h1>{{ displayRequestName(selectedRequest) }}</h1>
          <p>@{{ selectedRequest.from_username }} · ID {{ selectedRequest.from_user_id }}</p>
          <small>{{ selectedRequest.remark || '请求添加你为好友' }}</small>
          <time>{{ formatTime(selectedRequest.create_time) }}</time>
          <div class="detail-actions">
            <button class="primary" :disabled="handlingRequestId === selectedRequest.id" type="button" @click="acceptRequest(selectedRequest)">
              <Check :size="16" stroke-width="2" />
              <span>同意</span>
            </button>
            <button class="ghost" :disabled="handlingRequestId === selectedRequest.id" type="button" @click="rejectRequest(selectedRequest)">
              <X :size="16" stroke-width="2" />
              <span>拒绝</span>
            </button>
          </div>
        </div>
      </template>

      <template v-else-if="selectedFriend">
        <div class="contact-detail-card">
          <div class="detail-avatar">
            <img v-if="selectedFriend.avatar" :alt="displayFriendName(selectedFriend)" :src="selectedFriend.avatar" />
            <span v-else>{{ initials(displayFriendName(selectedFriend)) }}</span>
          </div>
          <h1>{{ displayFriendName(selectedFriend) }}</h1>
          <p>@{{ selectedFriend.username }} · ID {{ selectedFriend.id }}</p>
          <small>{{ selectedFriend.signature || '这个人还没有填写签名' }}</small>
          <div class="detail-actions">
            <button class="primary" :disabled="openingConversationFriendId === selectedFriend.id" type="button" @click="startChat(selectedFriend)">
              <MessageCircle :size="16" stroke-width="2" />
              <span>{{ openingConversationFriendId === selectedFriend.id ? '打开中' : '发送消息' }}</span>
            </button>
            <button class="danger" :disabled="deletingFriendId === selectedFriend.id" type="button" @click="removeFriend(selectedFriend)">
              <Trash2 :size="16" stroke-width="2" />
              <span>删除好友</span>
            </button>
          </div>
        </div>
      </template>

      <div v-else class="contact-detail-empty">
        <Users :size="34" stroke-width="1.8" />
        <p>请选择一个联系人</p>
      </div>
    </main>
  </section>
</template>




