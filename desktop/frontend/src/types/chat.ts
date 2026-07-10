export type NavKey = 'messages' | 'contacts' | 'ai'

export interface ChatMessage {
  id: string
  sender: 'me' | 'other'
  content: string
  time: string
}

export interface Conversation {
  id: string
  name: string
  avatar: string
  lastMessage: string
  lastTime: string
  unread: number
  messages: ChatMessage[]
}

export function createMockConversations(): Conversation[] {
  return [
    {
      id: 'team',
      name: '产品研发群',
      avatar: '研',
      lastMessage: '今天先把桌面端聊天页定稿。',
      lastTime: '10:28',
      unread: 3,
      messages: [
        { id: 'team-1', sender: 'other', content: '后端接口现在可以先用 mock 数据展示。', time: '10:15' },
        { id: 'team-2', sender: 'me', content: '收到，我先把三栏布局和切换逻辑做好。', time: '10:18' },
        { id: 'team-3', sender: 'other', content: '今天先把桌面端聊天页定稿。', time: '10:28' },
      ],
    },
    {
      id: 'design',
      name: '设计协作',
      avatar: '设',
      lastMessage: '整体保持微信 PC 版那种克制感。',
      lastTime: '09:42',
      unread: 1,
      messages: [
        { id: 'design-1', sender: 'other', content: '会话列表不要太花哨，注意信息密度。', time: '09:30' },
        { id: 'design-2', sender: 'me', content: '我会控制颜色和阴影，避免后台管理系统风格。', time: '09:35' },
        { id: 'design-3', sender: 'other', content: '整体保持微信 PC 版那种克制感。', time: '09:42' },
      ],
    },
    {
      id: 'ai',
      name: 'AI 助手',
      avatar: 'AI',
      lastMessage: '可以帮你整理聊天摘要。',
      lastTime: '昨天',
      unread: 0,
      messages: [
        { id: 'ai-1', sender: 'other', content: '你好，我可以帮你整理聊天摘要和待办。', time: '昨天' },
        { id: 'ai-2', sender: 'me', content: '后面接入真实 AI 接口时再扩展。', time: '昨天' },
      ],
    },
    {
      id: 'ops',
      name: '运维通知',
      avatar: '运',
      lastMessage: '数据库备份任务已完成。',
      lastTime: '周一',
      unread: 0,
      messages: [
        { id: 'ops-1', sender: 'other', content: '数据库备份任务已完成。', time: '周一' },
      ],
    },
  ]
}


