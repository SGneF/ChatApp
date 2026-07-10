<script lang="ts" setup>
import { Plus, Search, Send, UserPlus, X } from 'lucide-vue-next'
import { computed, reactive, ref } from 'vue'
import { ApplyFriend, SearchUsers } from '../../../wailsjs/go/main/App'
import type { ConversationItemData } from '../../api/conversation'
import { getToken } from '../../services/session'
import ConversationItem from './ConversationItem.vue'

interface SearchUserItem {
  id: number
  username: string
  nickname: string
  avatar: string
  signature: string
}

const props = defineProps<{
  conversations: ConversationItemData[]
  activeConversationId?: number
  loading?: boolean
}>()

const emit = defineEmits<{
  (event: 'select', value: ConversationItemData): void
  (event: 'toggle-top', value: ConversationItemData): void
  (event: 'delete', value: ConversationItemData): void
}>()

const keyword = ref('')
const addFriendOpen = ref(false)
const searching = ref(false)
const searched = ref(false)
const applyingUserId = ref<number | null>(null)
const applyError = ref('')
const applySuccess = ref('')
const searchUserKeyword = ref('')
const searchResults = ref<SearchUserItem[]>([])

const applyForm = reactive({
  remark: '',
})

const filteredConversations = computed(() => {
  const value = keyword.value.trim().toLowerCase()
  if (!value) return props.conversations

  return props.conversations.filter((item) => {
    const name = item.target_user.nickname || item.target_user.username || ''
    return name.toLowerCase().includes(value) || item.last_message.toLowerCase().includes(value)
  })
})

const canSearchUser = computed(() => searchUserKeyword.value.trim() !== '' && !searching.value)

function getErrorMessage(err: unknown) {
  if (err instanceof Error) return err.message
  if (typeof err === 'string') return err
  return '操作失败，请稍后重试'
}

function displayUserName(user: SearchUserItem) {
  return user.nickname || user.username || `用户 ${user.id}`
}

function initials(name: string) {
  return (name || 'LC').slice(0, 2).toUpperCase()
}

function resetAddFriendState() {
  searchUserKeyword.value = ''
  searchResults.value = []
  searched.value = false
  applyingUserId.value = null
  applyError.value = ''
  applySuccess.value = ''
  applyForm.remark = ''
}

function openAddFriend() {
  resetAddFriendState()
  addFriendOpen.value = true
}

function closeAddFriend() {
  addFriendOpen.value = false
}

async function searchUsers() {
  if (!canSearchUser.value) return

  const token = getToken()
  if (!token) {
    applyError.value = '登录已过期，请重新登录'
    return
  }

  searching.value = true
  searched.value = true
  applyError.value = ''
  applySuccess.value = ''

  try {
    searchResults.value = (await SearchUsers(token, searchUserKeyword.value.trim())) as SearchUserItem[]
  } catch (err) {
    searchResults.value = []
    applyError.value = getErrorMessage(err)
  } finally {
    searching.value = false
  }
}

async function submitApply(user: SearchUserItem) {
  const token = getToken()
  if (!token) {
    applyError.value = '登录已过期，请重新登录'
    return
  }

  applyingUserId.value = user.id
  applyError.value = ''
  applySuccess.value = ''

  try {
    await ApplyFriend(token, {
      to_user_id: user.id,
      remark: applyForm.remark.trim(),
    })
    applySuccess.value = `已向 ${displayUserName(user)} 发送好友申请`
  } catch (err) {
    applyError.value = getErrorMessage(err)
  } finally {
    applyingUserId.value = null
  }
}
</script>

<template>
  <section class="conversation-list-panel">
    <div class="conversation-search-bar">
      <label class="conversation-search">
        <Search :size="16" stroke-width="2" />
        <input v-model="keyword" autocomplete="off" placeholder="搜索会话" type="text" />
      </label>
      <button class="conversation-add" title="添加好友" type="button" @click="openAddFriend">
        <Plus :size="18" stroke-width="2" />
      </button>
    </div>

    <div class="conversation-list" aria-label="会话列表">
      <div v-if="props.loading" class="conversation-state">会话加载中...</div>
      <div v-else-if="filteredConversations.length === 0" class="conversation-empty">
        暂无会话，去找好友聊天吧
      </div>
      <template v-else>
        <ConversationItem
          v-for="conversation in filteredConversations"
          :key="conversation.id"
          :active="conversation.id === props.activeConversationId"
          :conversation="conversation"
          @delete="emit('delete', $event)"
          @select="emit('select', $event)"
          @toggle-top="emit('toggle-top', $event)"
        />
      </template>
    </div>

    <div v-if="addFriendOpen" class="modal-backdrop" @click.self="closeAddFriend">
      <section class="add-friend-modal" role="dialog" aria-label="添加好友">
        <header>
          <div>
            <h2>添加好友</h2>
            <p>搜索用户名、昵称或用户 ID 后发送申请</p>
          </div>
          <button title="关闭" type="button" @click="closeAddFriend">
            <X :size="18" stroke-width="2" />
          </button>
        </header>

        <form class="add-friend-form" @submit.prevent="searchUsers">
          <label>
            <span>查找用户</span>
            <input v-model="searchUserKeyword" autocomplete="off" placeholder="输入用户名、昵称或 ID" type="text" />
          </label>

          <button :disabled="!canSearchUser" type="submit">
            <Search :size="16" stroke-width="2" />
            <span>{{ searching ? '搜索中' : '搜索用户' }}</span>
          </button>

          <label>
            <span>申请备注</span>
            <textarea v-model="applyForm.remark" maxlength="255" placeholder="介绍一下你是谁"></textarea>
          </label>

          <p v-if="applyError" class="modal-message error">{{ applyError }}</p>
          <p v-if="applySuccess" class="modal-message success">{{ applySuccess }}</p>
        </form>

        <div v-if="searchResults.length" class="add-friend-results">
          <article v-for="user in searchResults" :key="user.id" class="add-friend-result">
            <div class="add-friend-avatar">
              <img v-if="user.avatar" :alt="displayUserName(user)" :src="user.avatar" />
              <span v-else>{{ initials(displayUserName(user)) }}</span>
            </div>
            <div class="add-friend-user">
              <strong>{{ displayUserName(user) }}</strong>
              <small>@{{ user.username }} · ID {{ user.id }}</small>
              <p v-if="user.signature">{{ user.signature }}</p>
            </div>
            <button :disabled="applyingUserId !== null" type="button" @click="submitApply(user)">
              <Send v-if="applyingUserId !== user.id" :size="15" stroke-width="2" />
              <UserPlus v-else :size="15" stroke-width="2" />
              <span>{{ applyingUserId === user.id ? '发送中' : '申请' }}</span>
            </button>
          </article>
        </div>

        <div v-else-if="searched && !searching && !applyError" class="add-friend-empty">
          没有找到匹配的用户
        </div>
      </section>
    </div>
  </section>
</template>


