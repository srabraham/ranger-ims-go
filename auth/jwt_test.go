package auth

import (
	"log"
	"log/slog"
	"testing"
	"time"
)

func TestGetJWT(t *testing.T) {
	jwter := JWTer{"some-secret"}
	j := jwter.CreateJWT("Hardware", time.Hour)
	log.Println(j)
	claims, err := jwter.AuthenticateJWT(j)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	slog.Info("teams are", "teams", claims.RangerTeams(), "len", len(claims.RangerTeams()))
}
