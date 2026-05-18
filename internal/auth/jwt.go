package auth

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type JWTService struct {
	secret []byte
}

func NewJWTService(secret string) *JWTService {
	if secret == "" {
		secret = generateRandom(32)
	}
	return &JWTService{secret: []byte(secret)}
}

func (j *JWTService) GenerateToken(user *User) (string, error) {
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *JWTService) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return j.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// ResolveAuth resolves a Bearer token to user info.
// Returns (user, nil) on success.
// If token starts with "tok_", it's a user API token.
// Otherwise it's treated as a JWT.
func ResolveAuth(header string, store *UserStore, jwtSvc *JWTService) *User {
	token := strings.TrimPrefix(header, "Bearer ")
	if token == "" || token == header {
		return nil
	}

	if strings.HasPrefix(token, "tok_") {
		return store.FindByToken(token)
	}

	claims, err := jwtSvc.ValidateToken(token)
	if err != nil {
		return nil
	}
	return store.FindByID(claims.UserID)
}
