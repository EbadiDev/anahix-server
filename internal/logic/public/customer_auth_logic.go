package public

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/EbadiDev/anahix-server/internal/logic/provider/smsir"
	"github.com/EbadiDev/anahix-server/internal/middleware"
	"github.com/EbadiDev/anahix-server/internal/model"
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/internal/types"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var mobileRegex = regexp.MustCompile(`^09[0-9]{9}$`)

type CustomerAuthLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCustomerAuthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CustomerAuthLogic {
	return &CustomerAuthLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// SendOTP handles phone validation, rate-limiting, code generation, and SMS.ir dispatch
func (l *CustomerAuthLogic) SendOTP(req *types.SendOTPRequest) (*types.SendOTPResponse, error) {
	phone := smsir.NormalizeMobile(req.Mobile)
	if !mobileRegex.MatchString(phone) {
		return nil, xerr.NewErrMsg("شماره موبایل وارد شده معتبر نمی‌باشد (فرمت صحیح: ۰۹xxxxxxxxx)")
	}

	// 1. Check 60s cooldown
	var recentOTP model.OTPVerification
	cooldownCutoff := time.Now().Add(-60 * time.Second)
	err := l.svcCtx.DB.WithContext(l.ctx).
		Where("phone = ? AND used = ? AND created_at > ?", phone, false, cooldownCutoff).
		Order("created_at desc").
		First(&recentOTP).Error

	if err == nil {
		elapsed := int(time.Since(recentOTP.CreatedAt).Seconds())
		remaining := 60 - elapsed
		if remaining < 1 {
			remaining = 1
		}
		return nil, xerr.NewErrMsg(fmt.Sprintf("لطفاً پیش از درخواست مجدد کد تایید، %d ثانیه شکیبا باشید", remaining))
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	// 2. Mark any older unused OTPs for this phone as used/invalidated
	l.svcCtx.DB.WithContext(l.ctx).
		Model(&model.OTPVerification{}).
		Where("phone = ? AND used = ?", phone, false).
		Update("used", true)

	// 3. Generate secure 5-digit code (10000 - 99999)
	n, err := rand.Int(rand.Reader, big.NewInt(90000))
	if err != nil {
		return nil, fmt.Errorf("failed to generate secure random code: %w", err)
	}
	code := fmt.Sprintf("%05d", n.Int64()+10000)
	log.Printf("📱 [Customer Auth OTP] Mobile: %s | Code: %s", phone, code)

	// 4. Retrieve SMS.ir configuration from DB
	var settings []model.SystemSetting
	l.svcCtx.DB.WithContext(l.ctx).Where("category = ?", "sms").Find(&settings)

	apiKey := ""
	templateID := 0
	for _, s := range settings {
		switch s.Key {
		case "smsir_api_key":
			apiKey = strings.TrimSpace(s.Value)
		case "smsir_template_id":
			cleaned := strings.Trim(s.Value, "'\" \t\r\n")
			if tid, err := strconv.Atoi(cleaned); err == nil && tid > 0 {
				templateID = tid
			}
		}
	}

	if templateID <= 0 {
		return nil, fmt.Errorf("شناسه قالب پیامک در سیستم پیکربندی نشده است")
	}

	// 5. Send via SMS.ir client
	_, err = l.svcCtx.SMSIRClient.SendVerify(l.ctx, apiKey, phone, templateID, code)
	if err != nil {
		return nil, fmt.Errorf("خطا در ارسال پیامک تایید: %w", err)
	}

	// 6. Save OTP record (120s TTL)
	otp := model.OTPVerification{
		Phone:     phone,
		Code:      code,
		ExpiresAt: time.Now().Add(120 * time.Second),
		Attempts:  0,
		Used:      false,
	}
	if err := l.svcCtx.DB.WithContext(l.ctx).Create(&otp).Error; err != nil {
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	return &types.SendOTPResponse{
		Sent:      true,
		Cooldown:  60,
		ExpiresIn: 120,
		Message:   "کد تایید ۵ رقمی به شماره موبایل شما ارسال شد",
	}, nil
}

// VerifyOTP validates the entered code and provisions or logs in the customer
func (l *CustomerAuthLogic) VerifyOTP(req *types.VerifyOTPRequest) (*types.VerifyOTPResponse, error) {
	phone := smsir.NormalizeMobile(req.Mobile)
	code := strings.TrimSpace(req.Code)

	if !mobileRegex.MatchString(phone) {
		return nil, xerr.NewErrMsg("شماره موبایل وارد شده معتبر نمی‌باشد")
	}
	if len(code) != 5 {
		return nil, xerr.NewErrMsg("کد تایید باید ۵ رقم باشد")
	}

	// 1. Locate active OTP record
	var otp model.OTPVerification
	err := l.svcCtx.DB.WithContext(l.ctx).
		Where("phone = ? AND used = ?", phone, false).
		Order("created_at desc").
		First(&otp).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, xerr.NewErrMsg("کد تاییدی برای این شماره یافت نشد یا منقضی شده است. لطفاً مجدداً درخواست کد دهید")
		}
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	// Check expiration
	if time.Now().After(otp.ExpiresAt) {
		l.svcCtx.DB.WithContext(l.ctx).Model(&otp).Update("used", true)
		return nil, xerr.NewErrMsg("کد تایید منقضی شده است. لطفاً کد جدید دریافت کنید")
	}

	// Check brute force attempts
	if otp.Attempts >= 5 {
		l.svcCtx.DB.WithContext(l.ctx).Model(&otp).Update("used", true)
		return nil, xerr.NewErrMsg("تعداد تلاش‌های ناموفق بیش از حد مجاز است. لطفاً مجدداً درخواست کد ارسال فرمایید")
	}

	// Validate code
	if otp.Code != code {
		l.svcCtx.DB.WithContext(l.ctx).Model(&otp).Update("attempts", otp.Attempts+1)
		remainingAttempts := 5 - (otp.Attempts + 1)
		return nil, xerr.NewErrMsg(fmt.Sprintf("کد تایید وارد شده نادرست است (%d فرصت باقی‌مانده)", remainingAttempts))
	}

	// Mark OTP as successfully used
	l.svcCtx.DB.WithContext(l.ctx).Model(&otp).Update("used", true)

	// 2. Find or create customer
	var customer model.Customer
	err = l.svcCtx.DB.WithContext(l.ctx).Where("phone = ?", phone).First(&customer).Error
	now := time.Now()
	isNew := false

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			isNew = true
			customer = model.Customer{
				Phone:        phone,
				FullName:     "",
				BalanceToman: 0,
				IsActive:     true,
				IsAdmin:      phone == "09928376377",
				LastLoginAt:  &now,
			}
			if err := l.svcCtx.DB.WithContext(l.ctx).Create(&customer).Error; err != nil {
				return nil, xerr.NewErrCode(xerr.DatabaseError)
			}
		} else {
			return nil, xerr.NewErrCode(xerr.DatabaseError)
		}
	} else {
		if !customer.IsActive {
			return nil, xerr.NewErrMsg("حساب کاربری شما مسدود شده است. با پشتیبانی تماس بگیرید")
		}
		if customer.FullName == "" || customer.FullName == "کاربر آناهیکس" {
			isNew = true
		}
		if customer.Phone == "09928376377" && !customer.IsAdmin {
			customer.IsAdmin = true
			l.svcCtx.DB.WithContext(l.ctx).Model(&customer).Update("is_admin", true)
		}
		customer.LastLoginAt = &now
		l.svcCtx.DB.WithContext(l.ctx).Model(&customer).Update("last_login_at", now)
	}

	// 3. Generate JWT customer session token (30 days validity)
	secret := l.svcCtx.Config.JWT.AccessSecret
	token, err := middleware.GenerateCustomerToken(secret, customer.ID.String(), customer.Phone, 30)
	if err != nil {
		return nil, fmt.Errorf("خطا در ایجاد توکن ورود: %w", err)
	}

	return &types.VerifyOTPResponse{
		Token:    token,
		IsNew:    isNew,
		Customer: mapCustomerResponse(&customer),
	}, nil
}

func mapCustomerResponse(customer *model.Customer) types.CustomerProfileResponse {
	firstName := customer.FirstName
	lastName := customer.LastName
	if firstName == "" && lastName == "" && customer.FullName != "" {
		parts := strings.Fields(customer.FullName)
		if len(parts) == 1 {
			firstName = parts[0]
		} else if len(parts) > 1 {
			firstName = parts[0]
			lastName = strings.Join(parts[1:], " ")
		}
	}

	return types.CustomerProfileResponse{
		ID:                   customer.ID.String(),
		Phone:                customer.Phone,
		Mobile:               customer.Phone,
		FirstName:            firstName,
		LastName:             lastName,
		FullName:             customer.FullName,
		Email:                customer.Email,
		BalanceToman:         customer.BalanceToman,
		Balance:              customer.BalanceToman,
		ReferralCode:         customer.ReferralCode,
		ReferralBalanceToman: customer.ReferralBalanceToman,
		IsActive:             customer.IsActive,
		IsAdmin:              customer.IsAdmin,
		IsCorporate:          customer.IsCorporate,
		CompanyName:          customer.CompanyName,
		NationalID:           customer.NationalID,
		EconomicCode:         customer.EconomicCode,
		RegistrationNo:       customer.RegistrationNo,
		PostalCode:           customer.PostalCode,
		CompanyAddress:       customer.CompanyAddress,
		CompanyPhone:         customer.CompanyPhone,
		CreatedAt:            customer.CreatedAt,
	}
}

// GetMe returns current logged-in customer profile and wallet balance
func (l *CustomerAuthLogic) GetMe(customerIDStr string) (*types.CustomerProfileResponse, error) {
	custUUID, err := uuid.Parse(customerIDStr)
	if err != nil {
		return nil, xerr.NewErrMsg("شناسه مشتری نامعتبر است")
	}

	var customer model.Customer
	if err := l.svcCtx.DB.WithContext(l.ctx).First(&customer, "id = ?", custUUID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, xerr.NewErrMsg("حساب کاربری یافت نشد")
		}
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	resp := mapCustomerResponse(&customer)
	return &resp, nil
}

// UpdateProfile modifies editable fields for customer
func (l *CustomerAuthLogic) UpdateProfile(customerIDStr string, req *types.UpdateCustomerProfileRequest) (*types.CustomerProfileResponse, error) {
	custUUID, err := uuid.Parse(customerIDStr)
	if err != nil {
		return nil, xerr.NewErrMsg("شناسه مشتری نامعتبر است")
	}

	var customer model.Customer
	if err := l.svcCtx.DB.WithContext(l.ctx).First(&customer, "id = ?", custUUID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, xerr.NewErrMsg("حساب کاربری یافت نشد")
		}
		return nil, xerr.NewErrCode(xerr.DatabaseError)
	}

	updates := map[string]interface{}{}
	if req.FirstName != "" {
		trimmed := strings.TrimSpace(req.FirstName)
		updates["first_name"] = trimmed
		customer.FirstName = trimmed
	}
	if req.LastName != "" {
		trimmed := strings.TrimSpace(req.LastName)
		updates["last_name"] = trimmed
		customer.LastName = trimmed
	}
	if req.FullName != "" {
		trimmed := strings.TrimSpace(req.FullName)
		updates["full_name"] = trimmed
		customer.FullName = trimmed
	} else if req.FirstName != "" || req.LastName != "" {
		full := strings.TrimSpace(customer.FirstName + " " + customer.LastName)
		updates["full_name"] = full
		customer.FullName = full
	}
	if req.Email != "" {
		trimmed := strings.TrimSpace(req.Email)
		updates["email"] = trimmed
		customer.Email = trimmed
	}
	if req.IsCorporate != nil {
		updates["is_corporate"] = *req.IsCorporate
		customer.IsCorporate = *req.IsCorporate
	}
	if req.CompanyName != "" {
		trimmed := strings.TrimSpace(req.CompanyName)
		updates["company_name"] = trimmed
		customer.CompanyName = trimmed
	}
	if req.NationalID != "" {
		trimmed := strings.TrimSpace(req.NationalID)
		updates["national_id"] = trimmed
		customer.NationalID = trimmed
	}
	if req.EconomicCode != "" {
		trimmed := strings.TrimSpace(req.EconomicCode)
		updates["economic_code"] = trimmed
		customer.EconomicCode = trimmed
	}
	if req.RegistrationNo != "" {
		trimmed := strings.TrimSpace(req.RegistrationNo)
		updates["registration_no"] = trimmed
		customer.RegistrationNo = trimmed
	}
	if req.PostalCode != "" {
		trimmed := strings.TrimSpace(req.PostalCode)
		updates["postal_code"] = trimmed
		customer.PostalCode = trimmed
	}
	if req.CompanyAddress != "" {
		trimmed := strings.TrimSpace(req.CompanyAddress)
		updates["company_address"] = trimmed
		customer.CompanyAddress = trimmed
	}
	if req.CompanyPhone != "" {
		trimmed := strings.TrimSpace(req.CompanyPhone)
		updates["company_phone"] = trimmed
		customer.CompanyPhone = trimmed
	}

	if len(updates) > 0 {
		if err := l.svcCtx.DB.WithContext(l.ctx).Model(&customer).Updates(updates).Error; err != nil {
			return nil, xerr.NewErrCode(xerr.DatabaseError)
		}
	}

	resp := mapCustomerResponse(&customer)
	return &resp, nil
}
