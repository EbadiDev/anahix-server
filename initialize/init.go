package initialize

import (
	"log"

	"github.com/EbadiDev/anahix-server/initialize/migrate"
	"gorm.io/gorm"
)

// StartInitDatabase runs all ArchNet-style migrations and initial database seeders
func StartInitDatabase(db *gorm.DB) {
	log.Println("⚡ [InitDB] Running ArchNet-style database initialization and migrations...")
	if err := migrate.InitBasicProducts(db); err != nil {
		log.Printf("⚠️ [InitDB] Product initialization notice: %v", err)
	}
}
