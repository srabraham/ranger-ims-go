package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"log"
	"time"
)

func GetJWT(rangerName, secret string, duration time.Duration) string {
	//secret := conf.Cfg.Core.JWTSecret
	//log.Println("secret:", secret)
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			// TODO
			"exp": time.Now().Add(duration).Unix(),
			"iat": time.Now().Unix(),
			"iss": "ranger-ims-go",
			
			// TODO
			"preferred_username": rangerName,
			"ranger_on_site":     true,
			"ranger_positions":   "Dirt - Green Dot,Green Dot Lead,Tech Ops,Green Dot Sanctuary,Operator,Tech On Call,Green Dot Lead Intern",
			"ranger_teams":       "Green Dot Team,Operator Team,Tech Cadre",
			"sub":                rangerName,
		},
	).SignedString([]byte(secret))
	if err != nil {
		log.Panic(err)
	}
	return token
}
