package auth

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"strings"
	"time"
)

type JWTer struct {
	SecretKey string
}

func (j JWTer) CreateJWT(rangerName string, duration time.Duration) string {
	token, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		NewIMSClaims().
			WithIssuedAt(time.Now()).
			WithExpiration(time.Now().Add(duration)).
			WithIssuer("ranger-ims-go").
			WithRangerHandle(rangerName).
			// TODO
			WithRangerOnSite(true).
			WithRangerPositions("Dirt - Green Dot", "Green Dot Lead").
			WithRangerTeams("Green Dot Team", "Operator Team").
			WithSubject("12345"),
	).SignedString([]byte(j.SecretKey))
	if err != nil {
		log.Panic(err)
	}
	return token
}

func (j JWTer) AuthenticateJWT(authHeader string) (*IMSClaims, error) {
	authHeader = strings.TrimPrefix(authHeader, "Bearer ")
	if authHeader == "" {
		return nil, fmt.Errorf("no token provided")
	}
	claims := IMSClaims{}
	tok, err := jwt.ParseWithClaims(authHeader, &claims, func(token *jwt.Token) (any, error) {
		return []byte(j.SecretKey), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
	if err != nil {
		return nil, fmt.Errorf("[jwt.Parse]: %w", err)
	}
	if tok == nil {
		return nil, fmt.Errorf("token is nil")
	}
	if !tok.Valid {
		return nil, fmt.Errorf("token is invalid")
	}
	if claims.RangerHandle() == "" {
		return nil, fmt.Errorf("ranger handle is required")
	}
	return &claims, nil
}
