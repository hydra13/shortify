package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/hydra13/shortify/internal/models"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

const (
	cookieTokenExp = time.Hour * 3
	cookieKey      = "t"
	secretKey      = "supersecretkey"
)

type AuthService struct{}

func New() *AuthService {
	return &AuthService{}
}

func (a *AuthService) GetUser(r *http.Request) (userID string, err error) {
	fmt.Println("cookie", r.Cookies())
	cookie, err := r.Cookie(cookieKey)
	if err != nil {
		return "", err
	}

	return a.getUserIDFromToken(cookie.Value)
}

func (a *AuthService) GetOrCreateUser(r *http.Request) (userID string, isNew bool) {
	userID, err := a.GetUser(r)
	if err != nil {
		return a.generateUserID(), true
	}

	return userID, false
}

func (a *AuthService) SetAuthCookie(w http.ResponseWriter, userID string) {
	tokenString, err := a.BuildJWTString(userID)
	if err == nil {
		http.SetCookie(w, &http.Cookie{
			Name:     cookieKey,
			Value:    tokenString,
			Expires:  time.Now().Add(cookieTokenExp),
			HttpOnly: true,
			// SameSite: http.SameSiteStrictMode,
			SameSite: http.SameSiteNoneMode,
		})
	}
}

func (a *AuthService) generateUserID() string {
	return uuid.New().String()
}

func (a *AuthService) BuildJWTString(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cookieTokenExp)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (a *AuthService) getUserIDFromToken(tokenString string) (string, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", models.ErrTokenNotValid
	}

	return claims.UserID, nil
}
