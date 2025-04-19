package auth

import (
	"log"
	"testing"
	"time"
)

func TestGetJWT(t *testing.T) {
	log.Println(JWTer{"some-secret"}.CreateJWT("Hardware", time.Hour))
}
