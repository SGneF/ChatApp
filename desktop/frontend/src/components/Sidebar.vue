<script lang="ts" setup>
import { Bot, Contact, LogOut, MessageCircle, Settings } from 'lucide-vue-next'
import type { NavKey } from '../types/chat'

const props = defineProps<{
  userName: string
  userAvatar?: string
  activeNav: NavKey
}>()

const emit = defineEmits<{
  (event: 'update:activeNav', value: NavKey): void
  (event: 'logout'): void
  (event: 'open-profile'): void
}>()

const navItems: Array<{ key: NavKey; label: string; icon: typeof MessageCircle }> = [
  { key: 'messages', label: '聊天', icon: MessageCircle },
  { key: 'contacts', label: '通讯录', icon: Contact },
  { key: 'ai', label: 'AI 助手', icon: Bot },
]

function initials(name: string) {
  return (name || 'LC').slice(0, 2).toUpperCase()
}
</script>

<template>
  <aside class="app-sidebar">
    <button class="sidebar-avatar" :title="`查看个人资料：${props.userName}`" type="button" @click="emit('open-profile')">
      <img v-if="props.userAvatar" :alt="props.userName" :src="props.userAvatar" />
      <span v-else>{{ initials(props.userName) }}</span>
    </button>

    <nav class="sidebar-nav" aria-label="主导航">
      <button
        v-for="item in navItems"
        :key="item.key"
        :class="['sidebar-button', { active: props.activeNav === item.key }]"
        :title="item.label"
        type="button"
        @click="emit('update:activeNav', item.key)"
      >
        <component :is="item.icon" :size="22" stroke-width="2" />
        <span>{{ item.label }}</span>
      </button>
    </nav>

    <div class="sidebar-bottom">
      <button class="sidebar-button sidebar-settings" title="设置" type="button">
        <Settings :size="22" stroke-width="2" />
        <span>设置</span>
      </button>

      <button class="sidebar-button sidebar-logout" title="退出登录" type="button" @click="emit('logout')">
        <LogOut :size="22" stroke-width="2" />
        <span>退出</span>
      </button>
    </div>
  </aside>
</template>
