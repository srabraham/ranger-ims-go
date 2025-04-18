package api

import (
	"database/sql"
	"encoding/json"
	"github.com/davecgh/go-spew/spew"
	"github.com/golang-jwt/jwt/v5"
	"github.com/srabraham/ranger-ims-go/auth"
	clubhousequeries "github.com/srabraham/ranger-ims-go/directory/queries"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type PostAuth struct {
	clubhouseDB *sql.DB
	jwtSecret   string
	jwtDuration time.Duration
}

type PostAuthResponse struct {
	Token string `json:"token"`
}

func (pa PostAuth) postAuth(w http.ResponseWriter, req *http.Request) {
	results, err := clubhousequeries.New(pa.clubhouseDB).RangersById(req.Context())
	if err != nil {
		return
	}
	req.Context()
	for _, r := range results {
		spew.Dump(r)
	}
	b, err := io.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		return
	}
	type AuthReq struct {
		Identification string `json:"identification"`
		Password       string `json:"password"`
	}

	vals := AuthReq{}
	if err = json.Unmarshal(b, &vals); err != nil {
		return
	}
	log.Printf("logging in as %v", vals.Identification)
	correct, _ := auth.VerifyPassword(vals.Password, "$2b$12$ls.WSuMjGignj1.Ob1jh8e:fcef3deeb59ec359a7b82f836d6fef9fc42042a3")
	log.Printf("password was valid? %v", correct)
	if !correct {
		return
	}
	jwt := auth.GetJWT(vals.Identification, pa.jwtSecret, pa.jwtDuration)
	resp := PostAuthResponse{Token: jwt}
	marsh, _ := json.Marshal(resp)
	_, _ = w.Write(marsh)
}

type GetAuth struct {
	clubhouseDB *sql.DB
	jwtSecret   string
}

type GetAuthResponse struct {
	Authenticated bool   `json:"authenticated"`
	User          string `json:"user"`
}

func (ga GetAuth) getAuth(w http.ResponseWriter, req *http.Request) {
	authHead := req.Header.Get("Authorization")
	authHead = strings.TrimPrefix(authHead, "Bearer ")
	tok, err := jwt.Parse(authHead, func(token *jwt.Token) (any, error) {
		return []byte(ga.jwtSecret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
	if err != nil {
		log.Printf("error parsing token: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if tok == nil {
		log.Println("token is nil")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if !tok.Valid {
		log.Println("token is invalid")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	sub, _ := tok.Claims.GetSubject()
	resp := GetAuthResponse{
		Authenticated: true,
		User:          sub,
	}
	marsh, _ := json.Marshal(resp)
	w.Write(marsh)
}
