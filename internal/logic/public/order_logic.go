package public

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/EbadiDev/anahix-server/internal/logic/audit"
	"github.com/EbadiDev/anahix-server/internal/logic/loyalty"
	"github.com/EbadiDev/anahix-server/internal/logic/notification"
	"github.com/EbadiDev/anahix-server/internal/logic/provider/fulfillment"
	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/exchangerate"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderLogic {
	return &OrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// generateSecureToken creates a 32-byte hex random string for order access
func generateSecureToken() string {
	bytes := make([]byte, 32)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// generateOrderNumber creates a human-readable order code
func generateOrderNumber() string {
	now := time.Now()
	randomSuffix := make([]byte, 2)
	_, _ = rand.Read(randomSuffix)
	return fmt.Sprintf("ANX-%s-%X", now.Format("0601021504"), randomSuffix)
}

func (l *OrderLogic) CreateOrder(req *types.CreateOrderRequest) (*types.CreateOrderResponse, error) {
	// 1. Resolve requested items
	type itemEntry struct {
		Product    model.Product
		Quantity   int
		PriceToman int64
		PlanID     string
		Plan       *types.ProductPlan
	}
	var entries []itemEntry

	resolvePlan := func(prod *model.Product, planID string) (int64, *types.ProductPlan) {
		if planID != "" && prod.PlansJSON != "" {
			var plans []types.ProductPlan
			if err := json.Unmarshal([]byte(prod.PlansJSON), &plans); err == nil {
				for _, pl := range plans {
					if pl.ID == planID {
						price := prod.PriceToman
						if pl.PriceToman > 0 {
							price = pl.PriceToman
						}
						return price, &pl
					}
				}
			}
		}
		return prod.PriceToman, nil
	}

	validateStock := func(prod *model.Product, qty int) error {
		if prod.StockMode == "out_of_stock" {
			return xerr.NewErrMsg(fmt.Sprintf("محصول «%s» در حال حاضر ناموجود است و امکان خرید وجود ندارد", prod.TitleFA))
		}
		if prod.StockMode == "always_in_stock" {
			return nil
		}
		if prod.StockMode == "manual" {
			if prod.ManualStock <= 0 {
				return xerr.NewErrMsg(fmt.Sprintf("محصول «%s» در حال حاضر ناموجود است (موجودی صفر)", prod.TitleFA))
			}
			if prod.ManualStock < qty {
				return xerr.NewErrMsg(fmt.Sprintf("موجودی محصول «%s» کافی نمی‌باشد (موجودی فعلی: %d عدد)", prod.TitleFA, prod.ManualStock))
			}
			return nil
		}
		// Automatic / Instant stock verification
		if prod.DeliveryStrategy == "third_party_instant" || prod.ExternalProvider != "" {
			if prod.ManualStock <= 0 {
				return xerr.NewErrMsg(fmt.Sprintf("محصول «%s» در حال حاضر ناموجود است (موجودی صفر)", prod.TitleFA))
			}
			if prod.ManualStock < qty {
				return xerr.NewErrMsg(fmt.Sprintf("موجودی محصول «%s» کافی نمی‌باشد (موجودی فعلی: %d عدد)", prod.TitleFA, prod.ManualStock))
			}
			return nil
		}
		if prod.DeliveryType == model.DeliveryInstant {
			var count int64
			l.svcCtx.DB.WithContext(l.ctx).Model(&model.InventoryItem{}).
				Where("product_id = ? AND status = ?", prod.ID, model.InventoryAvailable).
				Count(&count)
			if count <= 0 {
				return xerr.NewErrMsg(fmt.Sprintf("محصول «%s» در حال حاضر ناموجود است (موجودی صفر)", prod.TitleFA))
			}
			if count < int64(qty) {
				return xerr.NewErrMsg(fmt.Sprintf("موجودی کلیدهای محصول «%s» کافی نمی‌باشد (موجودی فعلی: %d عدد)", prod.TitleFA, count))
			}
		}
		return nil
	}

	if len(req.Items) > 0 {
		for _, it := range req.Items {
			pID, err := uuid.Parse(it.ProductID)
			if err != nil {
				continue
			}
			qty := it.Quantity
			if qty <= 0 {
				qty = 1
			}
			var prod model.Product
			if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ? AND is_active = ?", pID, true).First(&prod).Error; err == nil {
				if err := validateStock(&prod, qty); err != nil {
					return nil, err
				}
				planPrice, planObj := resolvePlan(&prod, it.PlanID)
				entries = append(entries, itemEntry{
					Product:    prod,
					Quantity:   qty,
					PriceToman: planPrice,
					PlanID:     it.PlanID,
					Plan:       planObj,
				})
			}
		}
	} else if req.ProductID != "" {
		pID, err := uuid.Parse(req.ProductID)
		if err == nil {
			var prod model.Product
			if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ? AND is_active = ?", pID, true).First(&prod).Error; err == nil {
				if err := validateStock(&prod, 1); err != nil {
					return nil, err
				}
				planPrice, planObj := resolvePlan(&prod, req.PlanID)
				entries = append(entries, itemEntry{
					Product:    prod,
					Quantity:   1,
					PriceToman: planPrice,
					PlanID:     req.PlanID,
					Plan:       planObj,
				})
			}
		}
	}

	if len(entries) == 0 {
		return nil, xerr.NewErrMsg("هیچ محصول معتبری در سبد خرید یافت نشد")
	}

	// 2. Compute Subtotal
	var subtotalToman int64
	for _, entry := range entries {
		subtotalToman += entry.PriceToman * int64(entry.Quantity)
	}

	// 3. Process Coupon if provided
	var couponDiscountToman int64
	var appliedCoupon *model.Coupon
	if req.CouponCode != "" {
		code := strings.ToUpper(strings.TrimSpace(req.CouponCode))
		var c model.Coupon
		if err := l.svcCtx.DB.WithContext(l.ctx).Where("code = ? AND is_active = ?", code, true).First(&c).Error; err == nil {
			now := time.Now()
			valid := true
			if c.ExpiresAt != nil && now.After(*c.ExpiresAt) {
				valid = false
			}
			if c.MaxUsage > 0 && c.CurrentUsage >= c.MaxUsage {
				valid = false
			}
			if c.MinOrderAmountToman > 0 && subtotalToman < c.MinOrderAmountToman {
				valid = false
			}
			if valid {
				if c.DiscountType == model.DiscountTypePercentage {
					couponDiscountToman = (subtotalToman * c.DiscountValue) / 100
					if c.MaxDiscountToman > 0 && couponDiscountToman > c.MaxDiscountToman {
						couponDiscountToman = c.MaxDiscountToman
					}
				} else {
					couponDiscountToman = c.DiscountValue
				}
				if couponDiscountToman > subtotalToman {
					couponDiscountToman = subtotalToman
				}
				appliedCoupon = &c
			}
		}
	}

	// 4. Resolve Dynamic Customer Loyalty Tier & VIP Automatic Discount
	customerTierCode := "bronze"
	priorityScore := 0
	var loyaltyDiscountToman int64 = 0

	var loyaltySetting model.SystemSetting
	loyaltyActive := true
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("category = ? AND key = ?", "loyalty", "loyalty_enabled").First(&loyaltySetting).Error; err == nil {
		val := strings.ToLower(strings.TrimSpace(loyaltySetting.Value))
		if val == "false" || val == "disabled" || val == "coming_soon" {
			loyaltyActive = false
		}
	}

	if loyaltyActive && req.CustomerPhone != "" {
		var customer model.Customer
		if err := l.svcCtx.DB.WithContext(l.ctx).Where("phone = ?", req.CustomerPhone).First(&customer).Error; err == nil {
			if customer.Tier != "" {
				customerTierCode = strings.ToLower(customer.Tier)
			}
		}
	}

	if loyaltyActive {
		// Fetch dynamic tier configuration from database (zero hardcoding)
		var activeTier model.LoyaltyTier
		if err := l.svcCtx.DB.WithContext(l.ctx).Where("code = ? AND is_active = ?", customerTierCode, true).First(&activeTier).Error; err == nil {
			priorityScore = activeTier.QueuePriority
			if activeTier.DiscountPercent > 0 {
				loyaltyDiscountToman = int64(float64(subtotalToman) * (activeTier.DiscountPercent / 100.0))
			}
		}
	}

	totalDiscountToman := couponDiscountToman + loyaltyDiscountToman
	if totalDiscountToman > subtotalToman {
		totalDiscountToman = subtotalToman
	}

	totalAmountToman := subtotalToman - totalDiscountToman
	if totalAmountToman < 0 {
		totalAmountToman = 0
	}

	accessToken := generateSecureToken()
	orderNumber := generateOrderNumber()
	expiresAt := time.Now().Add(20 * time.Minute)
	primaryDeliveryType := entries[0].Product.DeliveryType

	var createdOrder model.Order

	txErr := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		orderID := uuid.New()
		now := time.Now()

		// If payment is via wallet balance, verify and debit customer
		isPaid := false
		var payingCustomer *model.Customer
		if req.PaymentMethod == string(model.PaymentMethodBalance) {
			var cust model.Customer
			var findErr error
			if req.AuthCustomerID != nil {
				findErr = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
					Where("id = ? AND is_active = ?", *req.AuthCustomerID, true).First(&cust).Error
			} else if req.CustomerPhone != "" {
				// Fallback when invoked directly via internal subsystems or unit tests
				findErr = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
					Where("phone = ? AND is_active = ?", req.CustomerPhone, true).First(&cust).Error
			} else {
				return xerr.NewErrMsg("جهت پرداخت از موجودی کیف پول، لطفاً ابتدا وارد حساب کاربری خود شوید")
			}
			if findErr != nil {
				return xerr.NewErrMsg("حساب کاربری یافت نشد یا غیرفعال است")
			}
			if cust.BalanceToman < totalAmountToman {
				return xerr.NewErrMsg(fmt.Sprintf("موجودی کیف پول شما (%d تومان) برای پرداخت این سفارش (%d تومان) کافی نیست", cust.BalanceToman, totalAmountToman))
			}
			cust.BalanceToman -= totalAmountToman
			if err := tx.Save(&cust).Error; err != nil {
				return err
			}
			payingCustomer = &cust
			isPaid = true
		}

		// Prepare Order Items and reserve inventory where applicable
		var orderItems []model.OrderItem
		for _, entry := range entries {
			prod := entry.Product
			effectiveStrategy := prod.DeliveryStrategy
			effectiveProvider := prod.ExternalProvider
			externalProdID := prod.ExternalProductID
			externalProdIDStr := prod.ExternalProductIDStr
			costUSD := prod.BaseCostUSD

			// Plan-level granular override (e.g. Gemini 6m instant vs 18m manual or different bot)
			if entry.Plan != nil {
				if entry.Plan.DeliveryStrategy != "" {
					effectiveStrategy = entry.Plan.DeliveryStrategy
				}
				if entry.Plan.ExternalProvider != "" {
					effectiveProvider = entry.Plan.ExternalProvider
				}
				if entry.Plan.ExternalProductIDStr != "" {
					externalProdIDStr = entry.Plan.ExternalProductIDStr
				}
				if entry.Plan.ExternalProductID > 0 {
					externalProdID = entry.Plan.ExternalProductID
				}
				if entry.Plan.BaseCostUSD > 0 {
					costUSD = entry.Plan.BaseCostUSD
				}
			}

			if effectiveStrategy == "" || effectiveStrategy == "auto" {
				if externalProdID > 0 || externalProdIDStr != "" {
					effectiveStrategy = "third_party_instant"
				} else if prod.DeliveryType == model.DeliveryInstant {
					effectiveStrategy = "local_inventory"
				} else {
					effectiveStrategy = "manual_operator"
				}
			}

			for q := 0; q < entry.Quantity; q++ {
				var reservedItemID *uuid.UUID
				deliveredPayload := ""
				deliveredPayloadType := prod.DeliveredPayloadType
				if deliveredPayloadType == "" {
					deliveredPayloadType = "license_code"
				}
				externalOrderID := ""
				instructions := prod.Instructions

				if effectiveStrategy == "local_inventory" {
					var item model.InventoryItem
					err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
						Where("product_id = ? AND status = ?", prod.ID, model.InventoryAvailable).
						First(&item).Error
					if err != nil {
						return xerr.NewErrCode(xerr.OutOfStock)
					}

					if isPaid {
						item.Status = model.InventorySold
						item.SoldAt = &now
						item.DeliveredAt = &now
					} else {
						reservedUntil := expiresAt
						item.Status = model.InventoryReserved
						item.ReservedUntil = &reservedUntil
					}
					item.ReservedOrderID = &orderID
					if err := tx.Save(&item).Error; err != nil {
						return err
					}
					reservedItemID = &item.ID

					if isPaid {
						dec, decErr := l.svcCtx.Vault.Decrypt(item.PayloadEncrypted)
						if decErr == nil {
							deliveredPayload = dec
						}
					}
				} else if effectiveStrategy == "third_party_instant" && isPaid && (externalProdID > 0 || externalProdIDStr != "") {
					providerKey := effectiveProvider
					if providerKey == "" || providerKey == "third_party_api" {
						providerKey = fulfillment.ProviderInsightX
					}

					var settings []model.SystemSetting
					_ = tx.Where("category IN (?, ?, ?)", "bot_providers", "instant_delivery", "insightx").Find(&settings).Error

					configMap := make(map[string]string)
					for _, s := range settings {
						configMap[s.Key] = s.Value
					}

					apiKey := configMap["provider_"+providerKey+"_api_key"]
					baseURL := configMap["provider_"+providerKey+"_base_url"]
					if apiKey == "" && providerKey == fulfillment.ProviderInsightX {
						apiKey = configMap["provider_api_key"]
						if apiKey == "" {
							apiKey = configMap["insightx_api_key"]
						}
					}

					if apiKey != "" {
						prodIDStr := externalProdIDStr
						if prodIDStr == "" && externalProdID > 0 {
							prodIDStr = strconv.FormatInt(externalProdID, 10)
						}

						botReq := fulfillment.ProviderOrderRequest{
							ProductID:        prodIDStr,
							NumericID:        externalProdID,
							Quantity:         1,
							IdempotencyKey:   fmt.Sprintf("%s_%d", orderID.String(), q),
							CustomerEmail:    req.CustomerAccountEmail,
							CustomerPassword: req.CustomerAccountPassword,
							TelegramUsername: req.CustomerNotes,
						}

						botCfg := fulfillment.ProviderConfig{
							Key:     providerKey,
							APIKey:  apiKey,
							BaseURL: baseURL,
						}

						orderRes, oErr := l.svcCtx.FulfillmentManager.CreateOrder(providerKey, botCfg, botReq)
						if oErr == nil && orderRes != nil && orderRes.Success {
							deliveredPayload = orderRes.DeliveredPayload
							if orderRes.PayloadType != "" {
								deliveredPayloadType = orderRes.PayloadType
							}
							externalOrderID = orderRes.OrderID
							if orderRes.Instructions != "" {
								instructions = orderRes.Instructions
							} else if deliveredPayload != "" {
								if deliveredPayloadType == "activation_link" {
									instructions = "لینک فعال‌سازی اختصاصی شما: " + deliveredPayload
								} else {
									instructions = "کد لایسنس اختصاصی شما: " + deliveredPayload
								}
							}
						}
					}
				}

				// Resolve effective rate for cost tracking
				ratePriority := "manual"
				var vnSetting model.SystemSetting
				if err := tx.Where("key IN (?, ?)", "vn_rate_priority", "fivesim_rate_priority").First(&vnSetting).Error; err == nil {
					ratePriority = vnSetting.Value
				}
				usdRate := int64(95000)
				var usdRateSetting model.SystemSetting
				if err := tx.Where("key IN (?, ?)", "vn_usd_rate", "fivesim_usd_rate").First(&usdRateSetting).Error; err == nil {
					if val, err := strconv.ParseInt(usdRateSetting.Value, 10, 64); err == nil && val > 0 {
						usdRate = val
					}
				}
				effectiveRate, _, _ := exchangerate.GetDefaultClient().ResolveRate(ratePriority, usdRate)

				costToman := int64(costUSD * float64(effectiveRate))

				customerInput := ""
				if req.CustomerNotes != "" {
					customerInput = req.CustomerNotes
				}

				orderItems = append(orderItems, model.OrderItem{
					ID:                   uuid.New(),
					OrderID:              orderID,
					ProductID:            prod.ID,
					InventoryItemID:      reservedItemID,
					ProductTitle:         prod.TitleFA,
					ProductSlug:          prod.Slug,
					DeliveryType:         prod.DeliveryType,
					UnitPriceToman:       prod.PriceToman,
					Quantity:             1,
					SubtotalToman:        prod.PriceToman,
					CostUSD:              costUSD,
					CostToman:            costToman,
					DeliveredPayload:     deliveredPayload,
					DeliveredPayloadType: deliveredPayloadType,
					ExternalOrderID:      externalOrderID,
					ExternalProvider:     effectiveProvider,
					CustomerInputData:    customerInput,
					Instructions:         instructions,
				})
			}
		}

		// Encrypt credentials if on-account upgrade
		var encryptedCreds string
		if req.CustomerAccountEmail != "" {
			plainCreds := fmt.Sprintf("Email: %s | Pass: %s | Notes: %s",
				req.CustomerAccountEmail, req.CustomerAccountPassword, req.CustomerNotes)
			enc, encErr := l.svcCtx.Vault.Encrypt(plainCreds)
			if encErr == nil {
				encryptedCreds = enc
			}
		}

		initialStatus := model.OrderPendingPayment
		var paidAtTime *time.Time
		var completedAtTime *time.Time

		if isPaid {
			paidAtTime = &now

			// Decrement manual stock for products with manual stock mode
			for _, entry := range entries {
				if entry.Product.StockMode == "manual" {
					var currentProd model.Product
					if err := tx.Where("id = ?", entry.Product.ID).First(&currentProd).Error; err == nil {
						newStock := currentProd.ManualStock - entry.Quantity
						if newStock < 0 {
							newStock = 0
						}
						_ = tx.Model(&currentProd).Update("manual_stock", newStock).Error
					}
				}
			}

			// Check if all items are instantly fulfilled
			allInstantDelivered := true
			for _, oi := range orderItems {
				if oi.DeliveredPayload == "" {
					allInstantDelivered = false
					break
				}
			}
			if allInstantDelivered {
				initialStatus = model.OrderCompleted
				completedAtTime = &now
			} else {
				// Graceful fail-safe fallback to Operator Queue (Mode C)
				initialStatus = model.OrderInQueue
			}
		}

		isCorp := false
		if req.IsCorporate != nil {
			isCorp = *req.IsCorporate
		} else if payingCustomer != nil && payingCustomer.IsCorporate {
			isCorp = true
		}

		compName := req.CompanyName
		natID := req.NationalID
		econCode := req.EconomicCode
		regNo := req.RegistrationNo
		postCode := req.PostalCode
		buyerAddr := req.BuyerAddress
		buyerPhone := req.BuyerPhone

		if payingCustomer != nil && payingCustomer.IsCorporate {
			if compName == "" {
				compName = payingCustomer.CompanyName
			}
			if natID == "" {
				natID = payingCustomer.NationalID
			}
			if econCode == "" {
				econCode = payingCustomer.EconomicCode
			}
			if regNo == "" {
				regNo = payingCustomer.RegistrationNo
			}
			if postCode == "" {
				postCode = payingCustomer.PostalCode
			}
			if buyerAddr == "" {
				buyerAddr = payingCustomer.CompanyAddress
			}
			if buyerPhone == "" {
				buyerPhone = payingCustomer.CompanyPhone
			}
		}

		order := model.Order{
			ID:                           orderID,
			OrderNumber:                  orderNumber,
			CustomerPhone:                req.CustomerPhone,
			CustomerEmail:                req.CustomerEmail,
			AccessToken:                  accessToken,
			Status:                       initialStatus,
			DeliveryType:                 primaryDeliveryType,
			SubtotalAmountToman:          subtotalToman,
			DiscountAmountToman:          totalDiscountToman,
			LoyaltyDiscountToman:         loyaltyDiscountToman,
			CustomerTier:                 customerTierCode,
			PriorityScore:                priorityScore,
			CouponCode:                   req.CouponCode,
			ReferralCode:                 req.ReferralCode,
			TotalAmountToman:             totalAmountToman,
			PaymentMethod:                model.PaymentMethod(req.PaymentMethod),
			PaidAt:                       paidAtTime,
			CompletedAt:                  completedAtTime,
			ExpiresAt:                    expiresAt,
			CustomerCredentialsEncrypted: encryptedCreds,
			TwoFactorStatus:              model.TwoFactorIdle,
			IsCorporate:                  isCorp,
			CompanyName:                  compName,
			NationalID:                   natID,
			EconomicCode:                 econCode,
			RegistrationNo:               regNo,
			PostalCode:                   postCode,
			BuyerAddress:                 buyerAddr,
			BuyerPhone:                   buyerPhone,
			Items:                        orderItems,
		}

		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		// Record Coupon Usage if coupon applied and order paid
		if appliedCoupon != nil && isPaid {
			tx.Model(appliedCoupon).Update("current_usage", appliedCoupon.CurrentUsage+1)
			_ = tx.Create(&model.CouponUsage{
				ID:            uuid.New(),
				CouponID:      appliedCoupon.ID,
				CustomerPhone: req.CustomerPhone,
				OrderID:       orderID,
				DiscountToman: couponDiscountToman,
				CreatedAt:     now,
			})
		}

		// Referral Tracking & Commission (Dynamic percentage from system settings)
		if req.ReferralCode != "" && isPaid && totalAmountToman > 0 {
			var refSetting model.SystemSetting
			isRefActive := true
			if err := tx.Where("category = ? AND key = ?", "loyalty", "referral_enabled").First(&refSetting).Error; err == nil {
				val := strings.ToLower(strings.TrimSpace(refSetting.Value))
				if val == "false" || val == "disabled" || val == "coming_soon" {
					isRefActive = false
				}
			}

			if isRefActive {
				var referrer model.Customer
				if err := tx.Where("referral_code = ?", req.ReferralCode).First(&referrer).Error; err == nil {
					if referrer.Phone != req.CustomerPhone {
						commissionPct := 10.0
						var commSetting model.SystemSetting
						if err := tx.Where("category = ? AND key = ?", "loyalty", "referral_commission_rate").First(&commSetting).Error; err == nil {
							if parsedRate, err := strconv.ParseFloat(strings.TrimSpace(commSetting.Value), 64); err == nil && parsedRate > 0 {
								commissionPct = parsedRate
							}
						}

						commissionRate := commissionPct / 100.0
						commissionToman := int64(float64(totalAmountToman) * commissionRate)

					var customerID *uuid.UUID
					if payingCustomer != nil {
						customerID = &payingCustomer.ID
					}

					if customerID != nil {
						// Record referral link if not exists
						var existingRef model.Referral
						if err := tx.Where("referred_customer_id = ?", *customerID).First(&existingRef).Error; err != nil {
							_ = tx.Create(&model.Referral{
								ID:                 uuid.New(),
								ReferrerCustomerID: referrer.ID,
								ReferredCustomerID: *customerID,
								ReferralCode:       req.ReferralCode,
								CreatedAt:          now,
							})
						}

						// Award commission to referrer
						if commissionToman > 0 {
							_ = tx.Create(&model.ReferralCommission{
								ID:                 uuid.New(),
								ReferrerCustomerID: referrer.ID,
								ReferredCustomerID: *customerID,
								OrderID:            orderID,
								OrderAmountToman:   totalAmountToman,
								CommissionToman:    commissionToman,
								CommissionRate:     commissionRate,
								Status:             model.CommissionConfirmed,
								CreatedAt:          now,
							})

							_ = tx.Model(&referrer).Update("referral_balance_toman", referrer.ReferralBalanceToman+commissionToman)
						}
					}
				}
			}
		}
	}

		// Financial transaction log
		txnStatus := model.TxStatusPending
		if isPaid {
			txnStatus = model.TxStatusSuccess
		}
		_ = audit.RecordTransaction(l.ctx, tx, &model.Transaction{
			OrderID:         &order.ID,
			OrderType:       model.TxTypeStoreOrder,
			CustomerPhone:   order.CustomerPhone,
			CustomerEmail:   order.CustomerEmail,
			AmountToman:     order.TotalAmountToman,
			PaymentPlatform: model.PaymentPlatform(order.PaymentMethod),
			PaymentRefID:    order.PaymentRefID,
			Status:          txnStatus,
			Description:     fmt.Sprintf("سفارش فروشگاهی #%s (%d آیتم)", order.OrderNumber, len(orderItems)),
		})

		createdOrder = order
		return nil
	})

	if txErr != nil {
		return nil, txErr
	}

	// Recalculate customer loyalty tier if paid
	if (createdOrder.Status == model.OrderPaid || createdOrder.Status == model.OrderCompleted) && createdOrder.CustomerPhone != "" {
		go func(phone string) {
			_ = loyalty.RecalculateByPhone(context.Background(), l.svcCtx.DB, phone)
		}(createdOrder.CustomerPhone)
	}

	// Dispatch Pattern SMS upon instant delivery & completion
	if createdOrder.Status == model.OrderCompleted && createdOrder.CustomerPhone != "" {
		payload := ""
		if len(createdOrder.Items) > 0 {
			payload = createdOrder.Items[0].DeliveredPayload
			if payload == "" {
				payload = createdOrder.Items[0].ProductTitle
			}
		}
		notifier := notification.NewSMSPatternNotifier(l.svcCtx)
		notifier.SendOrderCompletedAsync(createdOrder.CustomerPhone, createdOrder.OrderNumber, payload)
	}

	paymentURL := fmt.Sprintf("/checkout/pay/%s?token=%s", createdOrder.ID.String(), accessToken)

	return &types.CreateOrderResponse{
		OrderID:             createdOrder.ID.String(),
		OrderNumber:         createdOrder.OrderNumber,
		AccessToken:         createdOrder.AccessToken,
		SubtotalAmountToman: createdOrder.SubtotalAmountToman,
		DiscountAmountToman: createdOrder.DiscountAmountToman,
		TotalAmountToman:    createdOrder.TotalAmountToman,
		PaymentMethod:       string(createdOrder.PaymentMethod),
		PaymentURL:          paymentURL,
		ExpiresAt:           createdOrder.ExpiresAt,
	}, nil
}

func (l *OrderLogic) GetOrderDetail(orderIDOrNo string, token string) (*types.OrderDetailResponse, error) {
	var order model.Order
	query := l.svcCtx.DB.WithContext(l.ctx).Preload("Items")

	if orderUUID, err := uuid.Parse(orderIDOrNo); err == nil {
		query = query.Where("id = ?", orderUUID)
	} else {
		query = query.Where("order_number = ?", orderIDOrNo)
	}

	if err := query.First(&order).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.OrderNotFound)
	}

	tokenMatches := token != "" && order.AccessToken == token

	maskedPhone := order.CustomerPhone
	if !tokenMatches && len(maskedPhone) >= 7 {
		maskedPhone = maskedPhone[:4] + "***" + maskedPhone[len(maskedPhone)-4:]
	}

	resp := &types.OrderDetailResponse{
		OrderID:              order.ID.String(),
		OrderNumber:          order.OrderNumber,
		Status:               string(order.Status),
		DeliveryType:         string(order.DeliveryType),
		SubtotalAmountToman:  order.SubtotalAmountToman,
		DiscountAmountToman:  order.DiscountAmountToman,
		CouponCode:           order.CouponCode,
		TotalAmountToman:     order.TotalAmountToman,
		CustomerPhone:        maskedPhone,
		PaymentMethod:        string(order.PaymentMethod),
		CreatedAt:            order.CreatedAt,
		PaidAt:               order.PaidAt,
		CompletedAt:          order.CompletedAt,
		TwoFactorPrompt:      order.TwoFactorPrompt,
		TwoFactorStatus:      string(order.TwoFactorStatus),
		TwoFactorRequestedAt: order.TwoFactorRequestedAt,
	}

	// If the order is paid or completed AND token matches, decrypt any fulfilled items
	isDelivered := order.Status == model.OrderPaid || order.Status == model.OrderCompleted
	if isDelivered && tokenMatches {
		var deliveredDTOs []types.DeliveredItemDTO
		for _, item := range order.Items {
			payloadText := ""
			if item.InventoryItemID != nil {
				var invItem model.InventoryItem
				if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", *item.InventoryItemID).First(&invItem).Error; err == nil {
					decrypted, decErr := l.svcCtx.Vault.Decrypt(invItem.PayloadEncrypted)
					if decErr == nil {
						payloadText = decrypted
					}
				}
			} else if item.DeliveredPayload != "" {
				payloadText = item.DeliveredPayload
			}

			deliveredDTOs = append(deliveredDTOs, types.DeliveredItemDTO{
				ProductTitle: item.ProductTitle,
				ProductSlug:  item.ProductSlug,
				DeliveryType: string(item.DeliveryType),
				Payload:      payloadText,
				Instructions: item.Instructions,
			})
		}
		resp.DeliveredItems = deliveredDTOs
	}

	return resp, nil
}

func (l *OrderLogic) SubmitTwoFactor(orderIDOrNo string, token string, code string) error {
	var order model.Order
	query := l.svcCtx.DB.WithContext(l.ctx)

	if orderUUID, err := uuid.Parse(orderIDOrNo); err == nil {
		query = query.Where("id = ?", orderUUID)
	} else {
		query = query.Where("order_number = ?", orderIDOrNo)
	}

	if token != "" {
		query = query.Where("access_token = ?", token)
	}

	if err := query.First(&order).Error; err != nil {
		return xerr.NewErrCode(xerr.OrderNotFound)
	}

	if order.TwoFactorStatus != model.TwoFactorRequested {
		return xerr.NewErrCodeMsg(xerr.OrderInvalidStatus, "در حال حاضر نیازی به کد تایید دومرحله‌ای نیست")
	}

	order.TwoFactorCode = code
	order.TwoFactorStatus = model.TwoFactorSubmitted

	if err := l.svcCtx.DB.WithContext(l.ctx).Save(&order).Error; err != nil {
		return xerr.NewErrCode(xerr.DatabaseError)
	}

	return nil
}
