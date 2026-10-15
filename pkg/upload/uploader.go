package upload

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/EbadiDev/anahix-server/internal/types"
)

const (
	MaxFileSize     = 10 * 1024 * 1024 // 10MB
	UploadDirectory = "uploads/tickets"
)

var allowedExtensions = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".webp": true,
	".gif":  true,
	".pdf":  true,
	".txt":  true,
	".log":  true,
	".zip":  true,
}

var imageExtensions = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".webp": true,
	".gif":  true,
}

// SaveTicketAttachment validates and stores an uploaded ticket file
func SaveTicketAttachment(fileHeader *multipart.FileHeader) (*types.UploadAttachmentResponse, error) {
	if fileHeader == nil {
		return nil, errors.New("فایلی ارسال نشده است")
	}

	if fileHeader.Size > MaxFileSize {
		return nil, fmt.Errorf("حجم فایل نباید بیشتر از ۱۰ مگابایت باشد (حجم فایل شما: %.1f مگابایت)", float64(fileHeader.Size)/(1024*1024))
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !allowedExtensions[ext] {
		return nil, errors.New("فرمت فایل مجاز نیست (فرمت‌های مجاز: عکس، PDF، TXT، LOG، ZIP)")
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(UploadDirectory, 0755); err != nil {
		return nil, fmt.Errorf("خطا در ایجاد پوشه ذخیره‌سازی: %w", err)
	}

	// Generate safe random unique filename
	randomBytes := make([]byte, 8)
	_, _ = rand.Read(randomBytes)
	uniqueFileName := fmt.Sprintf("%d_%s%s", time.Now().Unix(), hex.EncodeToString(randomBytes), ext)
	dstPath := filepath.Join(UploadDirectory, uniqueFileName)

	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("خطا در باز کردن فایل ارسالی: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		return nil, fmt.Errorf("خطا در ایجاد فایل روی دیسک: %w", err)
	}
	defer dst.Close()

	written, err := io.Copy(dst, src)
	if err != nil {
		return nil, fmt.Errorf("خطا در ذخیره‌سازی فایل: %w", err)
	}

	mimeType := fileHeader.Header.Get("Content-Type")
	if mimeType == "" {
		if imageExtensions[ext] {
			mimeType = "image/" + strings.TrimPrefix(ext, ".")
		} else {
			mimeType = "application/octet-stream"
		}
	}

	return &types.UploadAttachmentResponse{
		ID:       uniqueFileName,
		FileName: fileHeader.Filename,
		FileURL:  "/uploads/tickets/" + uniqueFileName,
		FileSize: written,
		MimeType: mimeType,
		IsImage:  imageExtensions[ext],
	}, nil
}
