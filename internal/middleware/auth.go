package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/EbadiDev/anahix-server/pkg/result"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AdminClaims struct {
	AdminID  string `json:"admin_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateAdminToken generates a signed JWT token for an admin user
func GenerateAdminToken(secret string, adminID string, username string, role string, expireHours int) (string, error) {
	claims := AdminClaims{
		AdminID:  adminID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "anahix-server",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// AdminAuthMiddleware validates the Bearer token in the Authorization header
func AdminAuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, result.Error(xerr.Unauthorized, "توکن احراز هویت ادمین ارسال نشده است"))
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, result.Error(xerr.Unauthorized, "فرمت توکن نامعتبر است"))
			return
		}

		tokenStr := parts[1]
		claims := &AdminClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, result.Error(xerr.Unauthorized, "توکن منقضی شده یا نامعتبر است"))
			return
		}

		c.Set("admin_id", claims.AdminID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// CorsMiddleware sets standard CORS headers for storefront and admin panel
func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Order-Token")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

type CustomerClaims struct {
	CustomerID string `json:"customer_id"`
	Mobile     string `json:"mobile"`
	jwt.RegisteredClaims
}

// GenerateCustomerToken generates a signed JWT token for a customer
func GenerateCustomerToken(secret string, customerID string, mobile string, expireDays int) (string, error) {
	claims := CustomerClaims{
		CustomerID: customerID,
		Mobile:     mobile,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireDays) * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "anahix-server",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// CustomerAuthMiddleware validates customer Bearer token
func CustomerAuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, result.Error(xerr.Unauthorized, "لطفاً ابتدا وارد حساب کاربری خود شوید"))
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, result.Error(xerr.Unauthorized, "فرمت توکن ورود نامعتبر است"))
			return
		}

		tokenStr := parts[1]
		claims := &CustomerClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, result.Error(xerr.Unauthorized, "نشست شما منقضی شده است، لطفاً مجدداً وارد شوید"))
			return
		}

		c.Set("customer_id", claims.CustomerID)
		c.Set("mobile", claims.Mobile)
		c.Next()
	}
}

// CustomerOptionalAuthMiddleware extracts customer identity if token is present, but doesn't abort if absent
func CustomerOptionalAuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				tokenStr := parts[1]
				claims := &CustomerClaims{}
				token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, errors.New("unexpected signing method")
					}
					return []byte(secret), nil
				})

				if err == nil && token.Valid {
					c.Set("customer_id", claims.CustomerID)
					c.Set("mobile", claims.Mobile)
				}
			}
		}
		c.Next()
	}
}

