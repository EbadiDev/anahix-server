package xerr

const (
	OK    uint32 = 200
	ERROR uint32 = 500

	// Common errors
	InvalidParams uint32 = 10001
	Unauthorized  uint32 = 10002
	Forbidden     uint32 = 10003
	NotFound      uint32 = 10004
	DatabaseError uint32 = 10005

	// Product & Inventory errors
	ProductNotFound    uint32 = 20001
	ProductInactive    uint32 = 20002
	OutOfStock         uint32 = 20003
	InvalidItemPayload uint32 = 20004

	// Order & Fulfillment errors
	OrderNotFound       uint32 = 30001
	OrderExpired        uint32 = 30002
	OrderAlreadyPaid    uint32 = 30003
	OrderInvalidStatus  uint32 = 30004
	InvalidAccessToken  uint32 = 30005
	TwoFactorRequired   uint32 = 30006
	TwoFactorExpired    uint32 = 30007

	// Payment errors
	PaymentFailed       uint32 = 40001
	PaymentVerification uint32 = 40002
	GatewayUnavailable  uint32 = 40003
)
