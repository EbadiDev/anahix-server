package notification

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/EbadiDev/anahix-server/internal/logic/provider/smsir"
	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
)

// SMSPatternNotifier manages sending high-priority SMS messages using SMS.ir Verify/Pattern templates
// to guarantee 100% delivery bypassing telecom blacklists (تبلیغات مسدود).
type SMSPatternNotifier struct {
	svcCtx *svc.ServiceContext
}

func NewSMSPatternNotifier(svcCtx *svc.ServiceContext) *SMSPatternNotifier {
	return &SMSPatternNotifier{svcCtx: svcCtx}
}

// getSMSConfig retrieves relevant SMS settings from the database system table
func (n *SMSPatternNotifier) getSMSConfig(ctx context.Context) (apiKey string, lineNum int64, templates map[string]int) {
	templates = map[string]int{
		"otp":            100000,
		"two_factor":     100001,
		"order_complete": 100002,
		"renewal":        100003,
	}

	var settings []model.SystemSetting
	_ = n.svcCtx.DB.WithContext(ctx).Where("category = ?", "sms").Find(&settings).Error

	configMap := make(map[string]string)
	for _, s := range settings {
		configMap[s.Key] = s.Value
	}

	apiKey = strings.TrimSpace(configMap["smsir_api_key"])
	lineStr := strings.Trim(configMap["smsir_line_number"], "'\" \t\r\n")
	lineNum, _ = strconv.ParseInt(lineStr, 10, 64)
	if lineNum <= 0 {
		lineNum = 30004505000017
	}

	parseTID := func(key string, fallback int) int {
		cleaned := strings.Trim(configMap[key], "'\" \t\r\n")
		if tid, err := strconv.Atoi(cleaned); err == nil && tid > 0 {
			return tid
		}
		return fallback
	}

	// OTP Template
	if tid := parseTID("smsir_template_otp", 0); tid > 0 {
		templates["otp"] = tid
	} else if tid := parseTID("smsir_template_id", 0); tid > 0 {
		templates["otp"] = tid
	}

	// 2FA Template
	if tid := parseTID("smsir_template_2fa", 0); tid > 0 {
		templates["two_factor"] = tid
	}

	// Order Complete Template
	if tid := parseTID("smsir_template_order_complete", 0); tid > 0 {
		templates["order_complete"] = tid
	}

	// Subscription Renewal Template
	if tid := parseTID("smsir_template_renewal", 0); tid > 0 {
		templates["renewal"] = tid
	}

	return apiKey, lineNum, templates
}

// SendWelcomeOTP sends OTP code using SMS.ir verify template (bypasses blacklist)
func (n *SMSPatternNotifier) SendWelcomeOTP(ctx context.Context, mobile string, code string) error {
	phone := smsir.NormalizeMobile(mobile)
	if phone == "" {
		return smsir.ErrInvalidMobile
	}

	apiKey, _, templates := n.getSMSConfig(ctx)
	templateID := templates["otp"]

	_, err := n.svcCtx.SMSIRClient.SendPattern(ctx, apiKey, phone, templateID, []smsir.VerifyParameter{
		{Name: "Code", Value: code},
	})
	if err != nil {
		log.Printf("❌ [SMSPatternNotifier] Failed to send OTP SMS to %s: %v", phone, err)
		return err
	}

	log.Printf("📱 [SMSPatternNotifier] OTP SMS sent successfully to %s (Template: %d)", phone, templateID)
	return nil
}

// SendTwoFactorRequestAsync notifies customer that operator requested a 2FA verification code
func (n *SMSPatternNotifier) SendTwoFactorRequestAsync(mobile string, orderNumber string, promptMessage string) {
	if strings.TrimSpace(mobile) == "" {
		return
	}

	go func(phone string, ordNum string, prompt string) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("⚠️ Panic recovered in 2FA SMS notification: %v", r)
			}
		}()

		bgCtx := context.Background()
		apiKey, lineNum, templates := n.getSMSConfig(bgCtx)
		templateID := templates["two_factor"]

		normMobile := smsir.NormalizeMobile(phone)
		if normMobile == "" {
			return
		}

		if prompt == "" {
			prompt = "لطفاً جهت ورود و تکمیل سفارش، کد تایید را در پنل کاربری ثبت کنید"
		}

		// Try sending via SMS.ir pattern (bypasses blacklist)
		if templateID > 0 {
			params := []smsir.VerifyParameter{
				{Name: "Order", Value: ordNum},
				{Name: "Prompt", Value: prompt},
			}
			_, err := n.svcCtx.SMSIRClient.SendPattern(bgCtx, apiKey, normMobile, templateID, params)
			if err == nil {
				log.Printf("📱 [SMSPatternNotifier] 2FA request pattern SMS sent to %s for order #%s (Template: %d)", normMobile, ordNum, templateID)
				return
			}
			log.Printf("⚠️ [SMSPatternNotifier] 2FA pattern SMS failed, attempting bulk fallback: %v", err)
		}

		// Fallback to text SMS if template fails or not configured
		fallbackText := fmt.Sprintf("آناهیکس: مشتری گرامی، جهت پردازش و تحویل سفارش #%s نیاز به کد تایید ۲FA است: %s\nورود به سفارش: anahix.ir/profile?tab=orders", ordNum, prompt)
		_, _ = n.svcCtx.SMSIRClient.SendBulk(bgCtx, apiKey, lineNum, fallbackText, []string{normMobile})
	}(mobile, orderNumber, promptMessage)
}

// SendOrderCompletedAsync notifies customer upon order completion and license delivery
func (n *SMSPatternNotifier) SendOrderCompletedAsync(mobile string, orderNumber string, licenseOrPayload string) {
	if strings.TrimSpace(mobile) == "" {
		return
	}

	go func(phone string, ordNum string, payload string) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("⚠️ Panic recovered in Order Completed SMS notification: %v", r)
			}
		}()

		bgCtx := context.Background()
		apiKey, lineNum, templates := n.getSMSConfig(bgCtx)
		templateID := templates["order_complete"]

		normMobile := smsir.NormalizeMobile(phone)
		if normMobile == "" {
			return
		}

		licenseSnippet := payload
		if len(licenseSnippet) > 80 {
			licenseSnippet = licenseSnippet[:80] + "..."
		}
		if licenseSnippet == "" {
			licenseSnippet = "تحویل داده شد (مشاهده در پنل)"
		}

		// Try sending via SMS.ir pattern (bypasses blacklist)
		if templateID > 0 {
			params := []smsir.VerifyParameter{
				{Name: "Order", Value: ordNum},
				{Name: "License", Value: licenseSnippet},
			}
			_, err := n.svcCtx.SMSIRClient.SendPattern(bgCtx, apiKey, normMobile, templateID, params)
			if err == nil {
				log.Printf("📱 [SMSPatternNotifier] Order completed pattern SMS sent to %s for order #%s (Template: %d)", normMobile, ordNum, templateID)
				return
			}
			log.Printf("⚠️ [SMSPatternNotifier] Order complete pattern SMS failed, attempting bulk fallback: %v", err)
		}

		// Fallback to text SMS
		fallbackText := fmt.Sprintf("آناهیکس: مشتری گرامی، سفارش #%s با موفقیت تکمیل شد.\nاطلاعات لایسنس/تحویل: %s\nمشاهده در حساب کاربری: anahix.ir/profile?tab=orders", ordNum, licenseSnippet)
		_, _ = n.svcCtx.SMSIRClient.SendBulk(bgCtx, apiKey, lineNum, fallbackText, []string{normMobile})
	}(mobile, orderNumber, licenseOrPayload)
}

// SendRenewalReminderAsync sends an alert reminding customer of subscription expiration
func (n *SMSPatternNotifier) SendRenewalReminderAsync(mobile string, serviceName string, daysRemaining int, renewURL string) {
	if strings.TrimSpace(mobile) == "" {
		return
	}

	go func(phone string, sName string, days int, url string) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("⚠️ Panic recovered in Renewal Reminder SMS notification: %v", r)
			}
		}()

		bgCtx := context.Background()
		apiKey, lineNum, templates := n.getSMSConfig(bgCtx)
		templateID := templates["renewal"]

		normMobile := smsir.NormalizeMobile(phone)
		if normMobile == "" {
			return
		}

		if url == "" {
			url = "anahix.ir/profile?tab=orders"
		}
		daysStr := strconv.Itoa(days)

		// Try sending via SMS.ir pattern (bypasses blacklist)
		if templateID > 0 {
			params := []smsir.VerifyParameter{
				{Name: "Service", Value: sName},
				{Name: "Days", Value: daysStr},
				{Name: "Link", Value: url},
			}
			_, err := n.svcCtx.SMSIRClient.SendPattern(bgCtx, apiKey, normMobile, templateID, params)
			if err == nil {
				log.Printf("📱 [SMSPatternNotifier] Renewal reminder pattern SMS sent to %s for %s (Template: %d)", normMobile, sName, templateID)
				return
			}
			log.Printf("⚠️ [SMSPatternNotifier] Renewal pattern SMS failed, attempting bulk fallback: %v", err)
		}

		// Fallback to text SMS
		fallbackText := fmt.Sprintf("آناهیکس: کاربر گرامی، اشتراک سرویس %s شما %s روز دیگر منقضی می‌شود.\nجهت تمدید سریع و بدون وقفه به لینک زیر مراجعه فرمایید:\n%s", sName, daysStr, url)
		_, _ = n.svcCtx.SMSIRClient.SendBulk(bgCtx, apiKey, lineNum, fallbackText, []string{normMobile})
	}(mobile, serviceName, daysRemaining, renewURL)
}

// NotifyRestock dispatches SMS alerts to all customers who requested restock notifications
func (n *SMSPatternNotifier) NotifyRestock(ctx context.Context, productID string, productTitle string, slug string) {
	go func(pID, pTitle, pSlug string) {
		bgCtx := context.Background()
		var alerts []model.StockAlert
		if err := n.svcCtx.DB.WithContext(bgCtx).
			Where("(product_id = ? OR product_id = ?) AND notified = ?", pID, pSlug, false).
			Find(&alerts).Error; err != nil || len(alerts) == 0 {
			return
		}

		apiKey, lineNum, templates := n.getSMSConfig(bgCtx)
		if apiKey == "" {
			return
		}

		brandName := "پلی شاپ"
		var brandSetting model.SystemSetting
		if err := n.svcCtx.DB.WithContext(bgCtx).Where("key = ?", "site_name_fa").First(&brandSetting).Error; err == nil && brandSetting.Value != "" {
			brandName = brandSetting.Value
		}

		templateID := templates["restock"]
		if templateID <= 0 {
			templateID = 100004
		}

		for _, alert := range alerts {
			normMobile := smsir.NormalizeMobile(alert.Mobile)
			if normMobile == "" {
				continue
			}

			// Try pattern first
			params := []smsir.VerifyParameter{
				{Name: "Brand", Value: brandName},
				{Name: "Product", Value: pTitle},
			}
			_, pErr := n.svcCtx.SMSIRClient.SendPattern(bgCtx, apiKey, normMobile, templateID, params)
			if pErr != nil {
				// Fallback to bulk SMS
				text := fmt.Sprintf("%s: کاربر گرامی، محصول «%s» موجود شد و هم‌اکنون آماده سفارش است.\nمشاهده و خرید: anahix.ir/products/%s", brandName, pTitle, pSlug)
				_, _ = n.svcCtx.SMSIRClient.SendBulk(bgCtx, apiKey, lineNum, text, []string{normMobile})
			}

			// Mark alert as notified
			_ = n.svcCtx.DB.WithContext(bgCtx).Model(&alert).Update("notified", true).Error
			log.Printf("📱 [StockAlert] Sent restock SMS to %s for product %s", normMobile, pTitle)
		}
	}(productID, productTitle, slug)
}
