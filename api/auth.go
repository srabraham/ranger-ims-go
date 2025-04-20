package api

import (
	"database/sql"
	"encoding/json"
	"github.com/davecgh/go-spew/spew"
	"github.com/srabraham/ranger-ims-go/api/access"
	"github.com/srabraham/ranger-ims-go/auth"
	"github.com/srabraham/ranger-ims-go/conf"
	clubhousequeries "github.com/srabraham/ranger-ims-go/directory/queries"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"io"
	"log"
	"net/http"
	"slices"
	"time"
)

type PostAuth struct {
	imsDB       *sql.DB
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
	jwt := auth.JWTer{SecretKey: conf.Cfg.Core.JWTSecret}.CreateJWT(vals.Identification, pa.jwtDuration)
	resp := PostAuthResponse{Token: jwt}
	marsh, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(marsh)
}

type GetAuth struct {
	imsDB       *sql.DB
	clubhouseDB *sql.DB
	jwtSecret   string
	admins      []string
}

type GetAuthResponse struct {
	Authenticated bool                      `json:"authenticated"`
	User          string                    `json:"user,omitzero"`
	Admin         bool                      `json:"admin"`
	EventAccess   map[string]AccessForEvent `json:"event_access"`
}

type AccessForEvent struct {
	ReadIncidents     bool `json:"readIncidents"`
	WriteIncidents    bool `json:"writeIncidents"`
	WriteFieldReports bool `json:"writeFieldReports"`
	AttachFiles       bool `json:"attachFiles"`
}

func (ga GetAuth) getAuth(w http.ResponseWriter, req *http.Request) {
	claims, err := auth.JWTer{SecretKey: conf.Cfg.Core.JWTSecret}.AuthenticateJWT(req.Header.Get("Authorization"))
	if err != nil {
		log.Printf("login failed: %v", err)
		resp := GetAuthResponse{Authenticated: false}
		marsh, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(marsh)
		return
	}

	log.Printf("got claims %v", claims)
	sub, _ := claims.GetSubject()
	resp := GetAuthResponse{
		Authenticated: true,
		User:          sub,
		Admin:         slices.Contains(ga.admins, sub),
		//EventAccess: map[string]AccessForEvent{
		//	"2023": {
		//		ReadIncidents:     true,
		//		WriteIncidents:    true,
		//		WriteFieldReports: true,
		//	},
		//},
	}

	req.ParseForm()

	eventName := req.Form.Get("event_id")
	if eventName != "" {
		ea, err := access.GetEventsAccess(req.Context(), ga.imsDB, eventName)
		if err != nil {
			log.Printf("failed to get events access: %v", err)
		} else {

			resp.EventAccess = map[string]AccessForEvent{
				eventName: {
					// TODO: need to check positions and teams as well for all three of these
					//  and need to check validity
					ReadIncidents: slices.ContainsFunc(ea[eventName].Readers, func(rule imsjson.AccessRule) bool {
						return rule.Expression == "person:"+sub
					}),
					WriteIncidents: slices.ContainsFunc(ea[eventName].Writers, func(rule imsjson.AccessRule) bool {
						return rule.Expression == "person:"+sub
					}),
					WriteFieldReports: slices.ContainsFunc(ea[eventName].Reporters, func(rule imsjson.AccessRule) bool {
						return rule.Expression == "person:"+sub
					}),
					AttachFiles: false,
				},
			}
			// TODO: this is ugly... don't do this
			sillyOverride := resp.EventAccess[eventName]
			if sillyOverride.WriteIncidents {
				sillyOverride.ReadIncidents = true
				sillyOverride.WriteFieldReports = true
			}
			resp.EventAccess[eventName] = sillyOverride
		}
	}

	marsh, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(marsh)
}
