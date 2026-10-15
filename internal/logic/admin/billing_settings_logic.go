package admin

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/EbadiDev/anahix-server/internal/logic/audit"
	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
)

type BillingSettingsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBillingSettingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BillingSettingsLogic {
	return &BillingSettingsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetBillingSettings loads current official corporate seller billing settings
func (l *BillingSettingsLogic) GetBillingSettings() (*types.BillingSettingsDTO, error) {
	dto := &types.BillingSettingsDTO{
		SellerCompanyName:    "شرکت نوآوران داده‌پرداز آناهیکس (با مسئولیت محدود)",
		SellerBrandFA:        "فروشگاه تخصصی هوش مصنوعی آناهیکس",
		SellerNationalID:     "14012345678",
		SellerEconomicCode:   "411567891234",
		SellerRegistrationNo: "589234",
		SellerPostalCode:     "1997834561",
		SellerProvince:       "تهران",
		SellerCity:           "تهران",
		SellerAddress:        "تهران، خیابان ولیعصر، تقاطع میرداماد، برج فناوری اطلاعات، طبقه ۵، واحد ۱۵",
		SellerPhone:          "021-88776655",
		SellerEmail:          "billing@anahix.com",
		SellerWebsite:        "https://anahix.com",
		VATPercent:           10,
		VATEnabled:           true,
		InvoiceNotes:         "این صورتحساب مطابق با ماده ۱۹ قانون مالیات بر ارزش افزوده تنظیم گردیده و مشمول خرید رسمی شرکتی و سازمانی می‌باشد.",
		TaxMemoryID:          "ANX-TAX-9824",
		StampSignatureURL:    "",
	}

	var settings []model.SystemSetting
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("category = ?", "billing").Find(&settings).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	for _, s := range settings {
		val := strings.TrimSpace(s.Value)
		switch s.Key {
		case "billing_seller_company_name":
			if val != "" {
				dto.SellerCompanyName = val
			}
		case "billing_seller_brand_fa":
			if val != "" {
				dto.SellerBrandFA = val
			}
		case "billing_seller_national_id":
			if val != "" {
				dto.SellerNationalID = val
			}
		case "billing_seller_economic_code":
			if val != "" {
				dto.SellerEconomicCode = val
			}
		case "billing_seller_registration_no":
			if val != "" {
				dto.SellerRegistrationNo = val
			}
		case "billing_seller_postal_code":
			if val != "" {
				dto.SellerPostalCode = val
			}
		case "billing_seller_province":
			if val != "" {
				dto.SellerProvince = val
			}
		case "billing_seller_city":
			if val != "" {
				dto.SellerCity = val
			}
		case "billing_seller_address":
			if val != "" {
				dto.SellerAddress = val
			}
		case "billing_seller_phone":
			if val != "" {
				dto.SellerPhone = val
			}
		case "billing_seller_email":
			if val != "" {
				dto.SellerEmail = val
			}
		case "billing_seller_website":
			if val != "" {
				dto.SellerWebsite = val
			}
		case "billing_tax_memory_id":
			if val != "" {
				dto.TaxMemoryID = val
			}
		case "billing_stamp_signature_url":
			dto.StampSignatureURL = val
		case "billing_vat_percent":
			if p, err := strconv.Atoi(val); err == nil && p >= 0 {
				dto.VATPercent = p
			}
		case "billing_vat_enabled":
			dto.VATEnabled = val == "true" || val == "1"
		case "billing_invoice_notes":
			if val != "" {
				dto.InvoiceNotes = val
			}
		}
	}

	return dto, nil
}

// UpdateBillingSettings updates official corporate seller billing settings in system table
func (l *BillingSettingsLogic) UpdateBillingSettings(req *types.UpdateBillingSettingsRequest) error {
	upsertSetting := func(key, value, settingType, desc string) error {
		var setting model.SystemSetting
		err := l.svcCtx.DB.WithContext(l.ctx).Where("category = ? AND key = ?", "billing", key).First(&setting).Error
		if err != nil {
			setting = model.SystemSetting{
				Category:  "billing",
				Key:       key,
				Value:     value,
				Type:      settingType,
				Desc:      desc,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			return l.svcCtx.DB.WithContext(l.ctx).Create(&setting).Error
		}
		setting.Value = value
		setting.UpdatedAt = time.Now()
		return l.svcCtx.DB.WithContext(l.ctx).Save(&setting).Error
	}

	if req.SellerCompanyName != nil {
		if err := upsertSetting("billing_seller_company_name", strings.TrimSpace(*req.SellerCompanyName), "string", "نام رسمی شرکت / فروشگاه جهت صدور فاکتور رسمی"); err != nil {
			return err
		}
	}
	if req.SellerBrandFA != nil {
		if err := upsertSetting("billing_seller_brand_fa", strings.TrimSpace(*req.SellerBrandFA), "string", "نام تجاری یا برند فارسی فروشگاه"); err != nil {
			return err
		}
	}
	if req.SellerNationalID != nil {
		if err := upsertSetting("billing_seller_national_id", strings.TrimSpace(*req.SellerNationalID), "string", "شناسه ملی شرکت فروشنده"); err != nil {
			return err
		}
	}
	if req.SellerEconomicCode != nil {
		if err := upsertSetting("billing_seller_economic_code", strings.TrimSpace(*req.SellerEconomicCode), "string", "کد اقتصادی ۱۲ رقمی دارایی"); err != nil {
			return err
		}
	}
	if req.SellerRegistrationNo != nil {
		if err := upsertSetting("billing_seller_registration_no", strings.TrimSpace(*req.SellerRegistrationNo), "string", "شماره ثبت شرکت"); err != nil {
			return err
		}
	}
	if req.SellerPostalCode != nil {
		if err := upsertSetting("billing_seller_postal_code", strings.TrimSpace(*req.SellerPostalCode), "string", "کد پستی ۱۰ رقمی اقامتگاه قانونی فروشنده"); err != nil {
			return err
		}
	}
	if req.SellerProvince != nil {
		if err := upsertSetting("billing_seller_province", strings.TrimSpace(*req.SellerProvince), "string", "استان محل اقامتگاه قانونی"); err != nil {
			return err
		}
	}
	if req.SellerCity != nil {
		if err := upsertSetting("billing_seller_city", strings.TrimSpace(*req.SellerCity), "string", "شهر محل اقامتگاه قانونی"); err != nil {
			return err
		}
	}
	if req.SellerAddress != nil {
		if err := upsertSetting("billing_seller_address", strings.TrimSpace(*req.SellerAddress), "string", "نشانی دقیق پستی دفتر مرکزی فروشنده"); err != nil {
			return err
		}
	}
	if req.SellerPhone != nil {
		if err := upsertSetting("billing_seller_phone", strings.TrimSpace(*req.SellerPhone), "string", "تلفن تماس ثابت و پشتیبانی صدور فاکتور"); err != nil {
			return err
		}
	}
	if req.SellerEmail != nil {
		if err := upsertSetting("billing_seller_email", strings.TrimSpace(*req.SellerEmail), "string", "ایمیل رسمی واحد مالی و صدور فاکتور"); err != nil {
			return err
		}
	}
	if req.SellerWebsite != nil {
		if err := upsertSetting("billing_seller_website", strings.TrimSpace(*req.SellerWebsite), "string", "آدرس وب‌سایت رسمی فروشنده"); err != nil {
			return err
		}
	}
	if req.VATPercent != nil {
		if err := upsertSetting("billing_vat_percent", strconv.Itoa(*req.VATPercent), "number", "درصد مالیات بر ارزش افزوده قانونی"); err != nil {
			return err
		}
	}
	if req.VATEnabled != nil {
		val := "false"
		if *req.VATEnabled {
			val = "true"
		}
		if err := upsertSetting("billing_vat_enabled", val, "bool", "فعال بودن محاسبه مالیات بر ارزش افزوده در فاکتور"); err != nil {
			return err
		}
	}
	if req.InvoiceNotes != nil {
		if err := upsertSetting("billing_invoice_notes", strings.TrimSpace(*req.InvoiceNotes), "string", "متن توضیحات و ضوابط حقوقی ذیل فاکتور رسمی"); err != nil {
			return err
		}
	}
	if req.TaxMemoryID != nil {
		if err := upsertSetting("billing_tax_memory_id", strings.TrimSpace(*req.TaxMemoryID), "string", "شناسه یکتای حافظه مالیاتی"); err != nil {
			return err
		}
	}
	if req.StampSignatureURL != nil {
		if err := upsertSetting("billing_stamp_signature_url", strings.TrimSpace(*req.StampSignatureURL), "string", "آدرس مهر و امضای دیجیتال رسمی فروشگاه"); err != nil {
			return err
		}
	}

	_ = audit.RecordAdminAction(l.ctx, l.svcCtx.DB, &model.AdminActionLog{
		AdminUsername: "admin",
		Action:        "UPDATE_BILLING_SETTINGS",
		Resource:      "settings",
		ResourceID:    "billing",
		Description:   "به‌روزرسانی مشخصات حقوقی و تنظیمات مالیاتی فاکتور رسمی فروشگاه",
	})

	return nil
}
