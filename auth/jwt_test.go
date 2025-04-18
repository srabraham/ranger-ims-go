package auth

import (
	"log"
	"testing"
)

func TestGetJWT(t *testing.T) {
	log.Println(GetJWT("Hardware", "some-secret"))
}
