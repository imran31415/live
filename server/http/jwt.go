package http

import (
	pb "admin/protos"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
)

const (
	subKey      = "sub"
	issClaimUrl = "https://auth.via.live/"
	pemCertUrl  = "https://auth.via.live/.well-known/jwks.json"
	userKey     = "user"
)

type Jwks struct {
	Keys []JSONWebKeys `json:"keys"`
}

type JSONWebKeys struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

func getPemCert(token *jwt.Token) (string, error) {
	cert := ""
	resp, err := http.Get(pemCertUrl)

	if err != nil {
		return cert, err
	}
	defer resp.Body.Close()

	var jwks = Jwks{}
	err = json.NewDecoder(resp.Body).Decode(&jwks)

	if err != nil {
		return cert, err
	}

	for k, _ := range jwks.Keys {
		if token.Header["kid"] == jwks.Keys[k].Kid {
			cert = "-----BEGIN CERTIFICATE-----\n" + jwks.Keys[k].X5c[0] + "\n-----END CERTIFICATE-----"
		}
	}

	if cert == "" {
		return cert, errors.New("Unable to find appropriate key.")
	}

	return cert, nil
}

func jwtMiddleWare() *jwtmiddleware.JWTMiddleware {
	return jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			// fmt.Println(token.Claims["sub"])
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				_ = getUserFromBackend(fmt.Sprintf("%v", claims[subKey]))

			}
			// Verify 'aud' claim
			aud := "api.via.live"
			checkAud := token.Claims.(jwt.MapClaims).VerifyAudience(aud, false)
			if !checkAud {
				return nil, errors.New("invalid audience")
			}
			// Verify 'iss' claim
			iss := issClaimUrl
			checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(iss, false)
			if !checkIss {
				return nil, errors.New("invalid issuer")
			}

			cert, err := getPemCert(token)
			if err != nil {
				log.Println("returned error in getPerm err: ", err)
				return nil, err
			}

			result, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
			return result, nil
		},
		SigningMethod: jwt.SigningMethodRS256,
		Debug:         false,
	})
}

func getUserFromBackend(sub string) *pb.User {
	u, err := httpClient.grpcClient.GetOrCreateUserBySubId(context.Background(), &pb.SubId{Id: sub})
	if err != nil {
		log.Println("Unable to get or create user from auth0 sub id")
		return nil
	}
	return u
}
