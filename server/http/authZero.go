package http

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
)

func extractAuthZeroSubId(r *http.Request) string {
	if userInfo := r.Context().Value(userKey); userInfo != nil {
		if t, ok := userInfo.(*jwt.Token); ok {
			if claims, k := t.Claims.(jwt.MapClaims); k {
				return fmt.Sprintf("%v", claims[subKey])
			}
		}
	}
	return ""
}
