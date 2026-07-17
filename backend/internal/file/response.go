package file

import "time"

type UploadFileResponse struct {
	ID         uint64    `json:"id"`
	FileType   string    `json:"file_type"`
	FileName   string    `json:"file_name"`
	ObjectName string    `json:"object_name"`
	MimeType   string    `json:"mime_type"`
	Size       int64     `json:"size"`
	URL        string    `json:"url"`
	CreateTime time.Time `json:"create_time"`
}
