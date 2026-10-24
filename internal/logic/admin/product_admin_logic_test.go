package admin

import (
	"context"
	"testing"

	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *svc.ServiceContext {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite in-memory db: %v", err)
	}

	if err := db.AutoMigrate(&model.Product{}, &model.InventoryItem{}); err != nil {
		t.Fatalf("failed to auto migrate: %v", err)
	}

	return &svc.ServiceContext{DB: db}
}

func TestProductAdminLogic_CRUD(t *testing.T) {
	svcCtx := setupTestDB(t)
	ctx := context.Background()
	l := NewProductAdminLogic(ctx, svcCtx)

	// 1. Create Product
	p := &model.Product{
		Title:        "ChatGPT Plus",
		TitleFA:      "اکانت چت‌جی‌پی‌تی پلاس",
		Slug:         "chatgpt-plus",
		PriceToman:   1950000,
		DeliveryType: model.DeliveryInstant,
		IsActive:     true,
	}

	created, err := l.CreateProduct(p)
	if err != nil {
		t.Fatalf("CreateProduct failed: %v", err)
	}
	if created.ID == uuid.Nil {
		t.Fatal("expected non-nil UUID for created product")
	}

	// 2. Get Product
	fetched, err := l.GetProduct(created.ID)
	if err != nil {
		t.Fatalf("GetProduct failed: %v", err)
	}
	if fetched.TitleFA != "اکانت چت‌جی‌پی‌تی پلاس" {
		t.Fatalf("expected title FA 'اکانت چت‌جی‌پی‌تی پلاس', got %s", fetched.TitleFA)
	}

	// 3. List Products
	list, err := l.ListProducts()
	if err != nil {
		t.Fatalf("ListProducts failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 product in list, got %d", len(list))
	}

	// 4. Update Product
	newTitleFA := "اکانت اختصاصی چت‌جی‌پی‌تی ۴"
	newPrice := int64(2100000)
	updated, err := l.UpdateProduct(created.ID, &types.AdminUpdateProductRequest{
		TitleFA:    &newTitleFA,
		PriceToman: &newPrice,
	})
	if err != nil {
		t.Fatalf("UpdateProduct failed: %v", err)
	}
	if updated.TitleFA != newTitleFA || updated.PriceToman != newPrice {
		t.Fatalf("updated product mismatch: %+v", updated)
	}

	// 5. Delete Product
	deleted, err := l.DeleteProduct(created.ID)
	if err != nil {
		t.Fatalf("DeleteProduct failed: %v", err)
	}
	if deleted.ID != created.ID {
		t.Fatalf("expected deleted ID %s, got %s", created.ID, deleted.ID)
	}

	// Verify not found after delete
	_, err = l.GetProduct(created.ID)
	if err == nil {
		t.Fatal("expected error getting deleted product, got nil")
	}
}
