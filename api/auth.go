package api

import (
	"database/sql"
	"encoding/json"
	"github.com/srabraham/ranger-ims-go/api/access"
	"github.com/srabraham/ranger-ims-go/auth"
	"github.com/srabraham/ranger-ims-go/conf"
	clubhousequeries "github.com/srabraham/ranger-ims-go/directory/queries"
	"github.com/srabraham/ranger-ims-go/store/queries"
	"io"
	"log"
	"log/slog"
	"net/http"
	"slices"
	"strings"
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
	//for _, r := range results {
	//	spew.Dump(r)
	//}
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
	slog.InfoContext(req.Context(), "attempting login", "identification", vals.Identification)

	var storedPassHash string
	var userID int64
	var onsite bool
	for _, person := range results {
		callsignMatch := person.Callsign == vals.Identification
		emailMatch := person.Email.Valid && strings.ToLower(person.Email.String) == strings.ToLower(vals.Identification)
		if callsignMatch || emailMatch {
			userID = person.ID
			storedPassHash = person.Password.String
			onsite = person.OnSite
			break
		}
	}

	correct, _ := auth.VerifyPassword(vals.Password, storedPassHash)
	log.Printf("password was correct? %v", correct)
	if !correct {
		slog.ErrorContext(req.Context(), "invalid credentials", "identification", vals.Identification)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// TODO: cache all this
	teams, _ := clubhousequeries.New(pa.clubhouseDB).Teams(req.Context())
	positions, _ := clubhousequeries.New(pa.clubhouseDB).Positions(req.Context())
	personTeams, _ := clubhousequeries.New(pa.clubhouseDB).PersonTeams(req.Context())
	personPositions, _ := clubhousequeries.New(pa.clubhouseDB).PersonPositions(req.Context())

	var foundPositions []uint64
	var foundPositionNames []string
	var foundTeams []int32
	var foundTeamNames []string
	for _, pp := range personPositions {
		if pp.PersonID == uint64(userID) {
			foundPositions = append(foundPositions, pp.PositionID)
		}
	}
	for _, pos := range positions {
		if slices.Contains(foundPositions, pos.ID) {
			foundPositionNames = append(foundPositionNames, pos.Title)
		}
	}
	for _, pt := range personTeams {
		if pt.PersonID == int32(userID) {
			foundTeams = append(foundTeams, pt.TeamID)
		}
	}
	for _, team := range teams {
		if slices.Contains(foundTeams, int32(team.ID)) {
			foundTeamNames = append(foundTeamNames, team.Title)
		}
	}

	jwt := auth.JWTer{SecretKey: conf.Cfg.Core.JWTSecret}.CreateJWT(vals.Identification, userID, foundPositionNames, foundTeamNames, onsite, pa.jwtDuration)
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
	handle := claims.RangerHandle()
	var roles []auth.Role
	if slices.Contains(ga.admins, handle) {
		roles = append(roles, auth.Administrator)
	}
	resp := GetAuthResponse{
		Authenticated: true,
		User:          handle,
		Admin:         slices.Contains(roles, auth.Administrator),
	}

	if err := req.ParseForm(); err != nil {
		slog.ErrorContext(req.Context(), "parseForm error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	eventName := req.Form.Get("event_id")
	if eventName != "" {
		//ea, err := access.GetEventsAccess(req.Context(), ga.imsDB, eventName)
		eventID, err := queries.New(ga.imsDB).QueryEventID(req.Context(), eventName)
		if err != nil {
			slog.ErrorContext(req.Context(), "queryEventID", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		ea, err := queries.New(ga.imsDB).EventAccess(req.Context(), eventID)
		if err != nil {
			slog.ErrorContext(req.Context(), "EventAccess", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		permissions := access.UserPermissions(ea, ga.admins, handle, claims.RangerOnSite(), claims.RangerPositions(), claims.RangerTeams())

		resp.EventAccess = map[string]AccessForEvent{
			eventName: {
				ReadIncidents:     permissions[auth.ReadIncidents],
				WriteIncidents:    permissions[auth.WriteIncidents],
				WriteFieldReports: permissions[auth.ReadWriteOwnFieldReports],
				AttachFiles:       false,
			},
		}
	}

	marsh, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(marsh)
}
