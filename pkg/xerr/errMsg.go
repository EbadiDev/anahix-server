package xerr

var message = map[uint32]string{
	OK:                  "عملیات با موفقیت انجام شد",
	ERROR:               "خطای داخلی سرور رخ داده است",
	InvalidParams:       "پارامترهای ارسالی نامعتبر است",
	Unauthorized:        "احراز هویت انجام نشده است",
	Forbidden:           "عدم دسترسی به منبع درخواستی",
	NotFound:            "موردی یافت نشد",
	DatabaseError:       "خطای پایگاه داده",
	ProductNotFound:     "محصول مورد نظر یافت نشد",
	ProductInactive:     "محصول در حال حاضر غیرفعال است",
	OutOfStock:          "موجودی این محصول به اتمام رسیده است",
	InvalidItemPayload:  "اطلاعات آیتم دیجیتال نامعتبر است",
	OrderNotFound:       "سفارش مورد نظر یافت نشد",
	OrderExpired:        "مهلت پرداخت سفارش منقضی شده است",
	OrderAlreadyPaid:    "سفارش قبلاً پرداخت شده است",
	OrderInvalidStatus:  "وضعیت سفارش برای این عملیات مجاز نیست",
	InvalidAccessToken:  "توکن دسترسی سفارش نامعتبر است",
	TwoFactorRequired:   "نیاز به ارسال کد تایید دومرحله‌ای است",
	TwoFactorExpired:    "مهلت ارسال کد تایید دومرحله‌ای منقضی شده است",
	PaymentFailed:       "پرداخت ناموفق بود",
	PaymentVerification: "خطا در تایید تراکنش بانکی",
	GatewayUnavailable:  "درگاه پرداخت در دسترس نیست",
}

func MapErrMsg(errCode uint32) string {
	if msg, ok := message[errCode]; ok {
		return msg
	}
	return "خطای ناشناخته رخ داده است"
}
