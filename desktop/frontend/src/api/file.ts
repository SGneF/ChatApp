import request from './request'
import type { MessageType } from './message'

export type UploadFileType = Extract<MessageType, 'image' | 'file' | 'voice'>

export interface UploadFileResponse {
  id: number
  file_type: UploadFileType
  file_name: string
  object_name: string
  mime_type: string
  size: number
  url: string
  create_time: string
}

export interface FileMessagePayload {
  id: number
  file_type: UploadFileType
  file_name: string
  object_name: string
  mime_type: string
  size: number
  url: string
}

export function uploadFile(file: File, type: UploadFileType) {
  const formData = new FormData()
  formData.append('file', file)
  formData.append('type', type)

  return request.post<UploadFileResponse, UploadFileResponse>('/file/upload', formData, {
    timeout: 120000,
  })
}

export function getFileURL(objectName: string) {
  return request.get<{ url: string }, { url: string }>('/file/url', {
    params: {
      object_name: objectName,
    },
  })
}

export function createFileMessageContent(upload: UploadFileResponse) {
  const payload: FileMessagePayload = {
    id: upload.id,
    file_type: upload.file_type,
    file_name: upload.file_name,
    object_name: upload.object_name,
    mime_type: upload.mime_type,
    size: upload.size,
    url: upload.url,
  }

  return JSON.stringify(payload)
}

export function parseFileMessageContent(content: string): FileMessagePayload | null {
  if (!content) return null

  try {
    const parsed = JSON.parse(content) as Partial<FileMessagePayload>
    if (!parsed || typeof parsed !== 'object') return null
    if (!parsed.object_name && !parsed.url) return null

    return {
      id: Number(parsed.id || 0),
      file_type: (parsed.file_type || 'file') as UploadFileType,
      file_name: parsed.file_name || '未命名文件',
      object_name: parsed.object_name || '',
      mime_type: parsed.mime_type || 'application/octet-stream',
      size: Number(parsed.size || 0),
      url: parsed.url || '',
    }
  } catch {
    return null
  }
}

export function formatFileSize(size: number) {
  if (!Number.isFinite(size) || size <= 0) return '未知大小'
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  if (size < 1024 * 1024 * 1024) return `${(size / 1024 / 1024).toFixed(1)} MB`
  return `${(size / 1024 / 1024 / 1024).toFixed(1)} GB`
}
