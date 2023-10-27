package user

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	authCookieName = "bietrunde"
	loginTime      = 265 * 24 * time.Hour
)

// User represents a logged in user.
type User struct {
	jwt.RegisteredClaims

	BieterID int  `json:"bieter-id"`
	IsAdmin  bool `json:"admin"`
}

// FromRequest reads the user from a request.
func FromRequest(r *http.Request, secred []byte) (User, error) {
	cookie, err := r.Cookie(authCookieName)
	if err != nil {
		return User{}, fmt.Errorf("reading cookie: %w", err)
	}

	var user User

	if _, err = jwt.ParseWithClaims(cookie.Value, &user, func(token *jwt.Token) (interface{}, error) {
		return secred, nil
	}); err != nil {
		return User{}, fmt.Errorf("parsing token: %w", err)
	}

	return user, nil
}

// FromID creates a user object from a bieterID.
func FromID(bieterID int) User {
	return User{
		BieterID: bieterID,
	}
}

// IsAnonymous tells, if the user is not logged in.
func (u User) IsAnonymous() bool {
	return u.BieterID == 0 && !u.IsAdmin
}

// SetCookie sets the cookie to the response.
func (u User) SetCookie(w http.ResponseWriter, secred []byte) error {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, u)

	tokenString, err := token.SignedString(secred)
	if err != nil {
		return fmt.Errorf("signing token: %w", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     authCookieName,
		Value:    tokenString,
		Path:     "/",
		MaxAge:   int(loginTime.Seconds()),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode, // This is necessary for the redirect from /?biet-id=123
	})

	return nil
}

// Logout removes the auth cookie
func Logout(w http.ResponseWriter) {
	c := &http.Cookie{
		Name:     authCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}

	http.SetCookie(w, c)
}
