package file

import "time"

const (
	FileTypeImage = "image"
	FileTypeFile  = "file"
	FileTypeVoice = "voice"
)

type FileRecord struct {
	ID uint64 `gorm:"primaryKey;autoIncrement" json:"id"`

	UploaderID uint64 `gorm:"not null;index" json:"uploader_id"`

	BucketName string `gorm:"type:varchar(100);not null" json:"bucket_name"`
	ObjectName string `gorm:"type:varchar(500);not null;uniqueIndex" json:"object_name"`

	OriginalName string `gorm:"type:varchar(255);not null" json:"original_name"`
	FileType     string `gorm:"type:varchar(20);not null;index" json:"file_type"`
	MimeType     string `gorm:"type:varchar(100)" json:"mime_type"`
	Size         int64  `gorm:"not null" json:"size"`

	CreateTime time.Time `gorm:"column:create_time;autoCreateTime" json:"create_time"`
}

func (FileRecord) TableName() string {
	return "file_records"
}
