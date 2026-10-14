package audit

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GenerateTransactionNumber creates a clean, collision-free transaction ID
func GenerateTransactionNumber(prefix string) string {
	if prefix == "" {
		prefix = "TXN"
	}
	buf := make([]byte, 3)
	_, _ = rand.Read(buf)
	return fmt.Sprintf("%s-%s-%s", prefix, time.Now().Format("0601021504"), hex.EncodeToString(buf))
}

// RecordTransaction writes a financial or product transaction record
func RecordTransaction(ctx context.Context, db *gorm.DB, tx *model.Transaction) error {
	if tx == nil {
		return nil
	}
	if tx.TransactionNumber == "" {
		tx.TransactionNumber = GenerateTransactionNumber("TXN")
	}
	if tx.CreatedAt.IsZero() {
		tx.CreatedAt = time.Now()
	}
	if tx.UpdatedAt.IsZero() {
		tx.UpdatedAt = time.Now()
	}

	err := db.WithContext(ctx).Create(tx).Error
	if err != nil {
		log.Printf("⚠️ [AUDIT] Failed to record transaction %s: %v", tx.TransactionNumber, err)
		return err
	}
	return nil
}

// RecordAdminAction persists an administrator audit log
func RecordAdminAction(ctx context.Context, db *gorm.DB, entry *model.AdminActionLog) error {
	if entry == nil || db == nil {
		return nil
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}

	err := db.WithContext(ctx).Create(entry).Error
	if err != nil {
		log.Printf("⚠️ [AUDIT] Failed to record admin action (%s by %s): %v", entry.Action, entry.AdminUsername, err)
		return err
	}
	return nil
}

// LogAdminActionAsync provides a fire-and-forget helper to log admin events without blocking request threads
func LogAdminActionAsync(db *gorm.DB, adminID *uuid.UUID, username, action, resource, resourceID, description string, details any, ip, userAgent string) {
	if db == nil {
		return
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("⚠️ [AUDIT PANIC] %v", r)
			}
		}()

		detailsJSON := "{}"
		if details != nil {
			if b, err := json.Marshal(details); err == nil {
				detailsJSON = string(b)
			}
		}

		entry := &model.AdminActionLog{
			AdminID:       adminID,
			AdminUsername: username,
			Action:        action,
			Resource:      resource,
			ResourceID:    resourceID,
			Description:   description,
			Details:       detailsJSON,
			IPAddress:     ip,
			UserAgent:     userAgent,
			CreatedAt:     time.Now(),
		}
		_ = RecordAdminAction(context.Background(), db, entry)
	}()
}
