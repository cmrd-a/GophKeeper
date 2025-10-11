package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var hmacSampleSecret []byte

func init() {
	hmacSampleSecret = []byte(os.Getenv("JWT_SECRET"))
}

// Claims holds the jwt claims we use.
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// CreateToken creates a signed JWT for given user id.
func CreateToken(userID string, ttl time.Duration) (string, error) {
	if len(hmacSampleSecret) == 0 {
		// fallback to a short-lived insecure secret if not provided
		hmacSampleSecret = []byte("dev-secret")
	}
	now := time.Now()
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(hmacSampleSecret)
}

// ParseAndValidate parses a token string and returns user id if valid.
func ParseAndValidate(tokenStr string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return hmacSampleSecret, nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.UserID, nil
	}
	return "", errors.New("invalid token")
}
