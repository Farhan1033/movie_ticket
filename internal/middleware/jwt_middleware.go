package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"movie-ticket/config"
	redis_config "movie-ticket/infra/redis"
	"movie-ticket/internal/auth_module/dto"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	ContextKeyRole  = "user_role"
	ContextKeyID    = "user_id"
	ContextKeyEmail = "user_email"
	TokenExpiry     = 24 * time.Hour
)

// Custom JWT Claims
// ID disimpan sebagai string dalam JWT untuk kompatibilitas, tapi akan dikonversi ke UUID saat digunakan
type CustomClaims struct {
	ID    string `json:"id"` // UUID sebagai string
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

// CreateToken membuat JWT token baru dan menyimpan session ke Redis
func CreateToken(id uuid.UUID, role, email string) (string, error) {
	// Buat claims
	claims := CustomClaims{
		ID:    id.String(),
		Email: email,
		Role:  role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "movie-ticket-app",
			Subject:   id.String(),
		},
	}

	// Buat token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.Get("JWT_SECRET")))
	if err != nil {
		return "", fmt.Errorf("failed to sign jwt token: %w", err)
	}

	// Simpan session ke Redis
	session := dto.UserSession{
		ID:    id,
		Email: email,
		Role:  role,
	}

	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return "", fmt.Errorf("failed to marshal session: %w", err)
	}

	// Set ke Redis dengan TTL
	ctx := context.Background()
	err = redis_config.RedisClient.Set(ctx, tokenString, sessionJSON, TokenExpiry).Err()
	if err != nil {
		return "", fmt.Errorf("failed to save session to redis: %w", err)
	}

	return tokenString, nil
}

// ParseToken memvalidasi dan parse JWT token
func ParseToken(tokenString string) (*jwt.Token, *CustomClaims, error) {
	// Parse token dengan custom claims
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validasi signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.Get("JWT_SECRET")), nil
	})

	if err != nil {
		// Dalam jwt/v5, error handling sudah lebih sederhana
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, nil, errors.New("token has expired")
		}
		if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, nil, errors.New("token not valid yet")
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, nil, errors.New("malformed token")
		}
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return nil, nil, errors.New("invalid token signature")
		}

		return nil, nil, fmt.Errorf("token validation failed: %w", err)
	}

	// Validasi token
	if !token.Valid {
		return nil, nil, errors.New("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, nil, errors.New("invalid token claims")
	}

	return token, claims, nil
}

// ValidateTokenInRedis memeriksa apakah token ada di Redis dan valid
func ValidateTokenInRedis(tokenString string) (*dto.UserSession, error) {
	ctx := context.Background()

	// Cek apakah token ada di Redis
	val, err := redis_config.RedisClient.Get(ctx, tokenString).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			return nil, errors.New("token not found or expired")
		}
		return nil, fmt.Errorf("redis error: %w", err)
	}

	// Parse session data
	var session dto.UserSession
	if err := json.Unmarshal([]byte(val), &session); err != nil {
		return nil, fmt.Errorf("invalid session data: %w", err)
	}

	return &session, nil
}

// RevokeToken menghapus token dari Redis (untuk logout)
func RevokeToken(tokenString string) error {
	ctx := context.Background()
	err := redis_config.RedisClient.Del(ctx, tokenString).Err()
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}
	return nil
}

// RefreshToken membuat token baru dan menghapus token lama
func RefreshToken(oldTokenString string) (string, error) {
	// Validasi token lama
	_, claims, err := ParseToken(oldTokenString)
	if err != nil {
		return "", fmt.Errorf("invalid old token: %w", err)
	}

	// Validasi session di Redis
	session, err := ValidateTokenInRedis(oldTokenString)
	if err != nil {
		return "", fmt.Errorf("session validation failed: %w", err)
	}

	// Hapus token lama
	if err := RevokeToken(oldTokenString); err != nil {
		return "", fmt.Errorf("failed to revoke old token: %w", err)
	}

	// Buat token baru
	userID, err := uuid.Parse(claims.ID)
	if err != nil {
		return "", fmt.Errorf("invalid user ID in claims: %w", err)
	}

	newToken, err := CreateToken(userID, session.Role, session.Email)
	if err != nil {
		return "", fmt.Errorf("failed to create new token: %w", err)
	}

	return newToken, nil
}

// JwtMiddleware adalah middleware untuk autentikasi JWT
func JwtMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Ambil Authorization header
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			return
		}

		// Validasi format Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			return
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Token is required",
			})
			return
		}

		// Parse dan validasi JWT token
		_, claims, err := ParseToken(tokenString)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": fmt.Sprintf("Token validation failed: %s", err.Error()),
			})
			return
		}

		// Validasi session di Redis
		session, err := ValidateTokenInRedis(tokenString)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": fmt.Sprintf("Session validation failed: %s", err.Error()),
			})
			return
		}

		// Cross-validation antara JWT claims dan Redis session
		if claims.ID != session.ID.String() || claims.Email != session.Email || claims.Role != session.Role {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Token and session data mismatch",
			})
			return
		}

		// Set data ke context
		ctx.Set(ContextKeyID, session.ID)
		ctx.Set(ContextKeyRole, session.Role)
		ctx.Set(ContextKeyEmail, session.Email)

		// Lanjutkan ke handler berikutnya
		ctx.Next()
	}
}

// RequireRole middleware untuk otorisasi berdasarkan role
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userRole, exists := ctx.Get(ContextKeyRole)
		if !exists {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "User role not found",
			})
			return
		}

		role, ok := userRole.(string)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Invalid user role type",
			})
			return
		}

		// Check if user role is in allowed roles
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				ctx.Next()
				return
			}
		}

		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "Insufficient permissions",
		})
	}
}

// Helper Functions

// GetUserIDFromContext mengambil user ID dari context
// Ambil token dari Authorization header
func getTokenFromHeader(ctx *gin.Context) (string, error) {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", errors.New("invalid authorization header format")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == "" {
		return "", errors.New("token is required")
	}

	return tokenString, nil
}

// GetUserIDFromRedis mengambil user ID dari Redis
func GetUserIDFromRedis(ctx *gin.Context) (uuid.UUID, error) {
	tokenString, err := getTokenFromHeader(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	session, err := ValidateTokenInRedis(tokenString)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get session from redis: %w", err)
	}

	return session.ID, nil
}

// GetUserRoleFromRedis mengambil role user dari Redis
func GetUserRoleFromRedis(ctx *gin.Context) (string, error) {
	tokenString, err := getTokenFromHeader(ctx)
	if err != nil {
		return "", err
	}

	session, err := ValidateTokenInRedis(tokenString)
	if err != nil {
		return "", fmt.Errorf("failed to get session from redis: %w", err)
	}

	return session.Role, nil
}

// GetUserEmailFromRedis mengambil email user dari Redis
func GetUserEmailFromRedis(ctx *gin.Context) (string, error) {
	tokenString, err := getTokenFromHeader(ctx)
	if err != nil {
		return "", err
	}

	session, err := ValidateTokenInRedis(tokenString)
	if err != nil {
		return "", fmt.Errorf("failed to get session from redis: %w", err)
	}

	return session.Email, nil
}

// GetAllUserDataFromRedis mengambil semua data user dari Redis
func GetAllUserDataFromRedis(ctx *gin.Context) (uuid.UUID, string, string, error) {
	tokenString, err := getTokenFromHeader(ctx)
	if err != nil {
		return uuid.Nil, "", "", err
	}

	session, err := ValidateTokenInRedis(tokenString)
	if err != nil {
		return uuid.Nil, "", "", fmt.Errorf("failed to get session from redis: %w", err)
	}

	return session.ID, session.Role, session.Email, nil
}

