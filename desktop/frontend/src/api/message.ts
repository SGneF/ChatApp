import request from './request'

export type MessageType = 'text' | 'image' | 'file' | 'voice'
export type MessageStatus = 'normal' | 'sent' | 'read' | 'revoked'

export interface MessageResponse {
  id: number
  conversation_id: number
  sender_id: number
  receiver_id: number
  type: MessageType
  content: string
  status: MessageStatus
  create_time: string
  update_time: string
}

export interface MessageHistoryResponse {
  list: MessageResponse[]
  total: number
  page: number
  page_size: number
}

export interface MessageReadResponse {
  conversation_id: number
  reader_id: number
  target_id: number
  read_count: number
}

export interface MessageRevokeResponse {
  message_id: number
  sender_id: number
  receiver_id: number
  status: 'revoked'
}

export function getMessageHistory(conversationId: number, page = 1, pageSize = 30) {
  return request.get<MessageHistoryResponse, MessageHistoryResponse>('/message/history', {
    params: {
      conversation_id: conversationId,
      page,
      page_size: pageSize,
    },
  })
}

export function revokeMessageRequest(messageId: number) {
  return request.post<MessageRevokeResponse, MessageRevokeResponse>(`/message/${messageId}/revoke`)
}
