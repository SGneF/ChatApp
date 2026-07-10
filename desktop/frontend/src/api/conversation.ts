import request from './request'

export interface ConversationTargetUser {
  id: number
  username: string
  nickname: string
  avatar: string
  signature: string
}

export interface ConversationItemData {
  id: number
  user_id: number
  target_id: number
  type: 'single' | 'group'
  target_user: ConversationTargetUser
  last_message_id: number
  last_message: string
  unread_count: number
  is_top: boolean
  create_time: string
  update_time: string
}

export function createSingleConversation(targetId: number) {
  return request.post<ConversationItemData, ConversationItemData>('/conversation/single', {
    target_id: targetId,
  })
}

export function getConversationList() {
  return request.get<ConversationItemData[], ConversationItemData[]>('/conversation/list')
}

export function getConversationDetail(conversationId: number) {
  return request.get<ConversationItemData, ConversationItemData>(`/conversation/${conversationId}`)
}

export function deleteConversation(conversationId: number) {
  return request.delete<void, void>(`/conversation/${conversationId}`)
}

export function markConversationRead(conversationId: number) {
  return request.post<void, void>(`/conversation/${conversationId}/read`)
}

export function setConversationTop(conversationId: number, isTop: boolean) {
  return request.post<void, void>(`/conversation/${conversationId}/top`, {
    is_top: isTop,
  })
}
