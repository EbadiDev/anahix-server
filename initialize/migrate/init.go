package migrate

import (
	"log"

	"gorm.io/gorm"
)

// InitBasicProducts executes database/00001_init_products.up.sql
func InitBasicProducts(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := ExecuteSQLFile(tx, "database/00001_init_products.up.sql"); err != nil {
			log.Printf("❌ [Migrate] Failed to execute 00001_init_products.up.sql: %v", err)
			return err
		}
		log.Println("🌱 [Migrate] Successfully executed 00001_init_products.up.sql (ArchNet standard)")
		return nil
	})
}
