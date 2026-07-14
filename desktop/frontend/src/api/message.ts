import request from './request'

export type MessageType = 'text' | 'image' | 'file' | 'voice'
export type MessageStatus = 'normal' | 'revoked'

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

export function getMessageHistory(conversationId: number, page = 1, pageSize = 30) {
  return request.get<MessageHistoryResponse, MessageHistoryResponse>('/message/history', {
    params: {
      conversation_id: conversationId,
      page,
      page_size: pageSize,
    },
  })
}
