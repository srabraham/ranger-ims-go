package api

import (
	"database/sql"
	"encoding/json"
	"github.com/srabraham/ranger-ims-go/auth"
	"github.com/srabraham/ranger-ims-go/conf"
	clubhousequeries "github.com/srabraham/ranger-ims-go/directory/clubhousedb"
	"io"
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

type PostAuthRequest struct {
	Identification string `json:"identification"`
	Password       string `json:"password"`
}
type PostAuthResponse struct {
	Token string `json:"token"`
}

func (action PostAuth) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	results, err := clubhousequeries.New(action.clubhouseDB).RangersById(req.Context())
	if err != nil {
		slog.Error("Failed to fetch Rangers", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	b, err := io.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		return
	}

	vals := PostAuthRequest{}
	if err = json.Unmarshal(b, &vals); err != nil {
		slog.Error("Failed to parse request body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
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
	if !correct {
		slog.Error("invalid credentials", "identification", vals.Identification)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// TODO: cache all this
	teams, _ := clubhousequeries.New(action.clubhouseDB).Teams(req.Context())
	positions, _ := clubhousequeries.New(action.clubhouseDB).Positions(req.Context())
	personTeams, _ := clubhousequeries.New(action.clubhouseDB).PersonTeams(req.Context())
	personPositions, _ := clubhousequeries.New(action.clubhouseDB).PersonPositions(req.Context())

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

	jwt := auth.JWTer{SecretKey: conf.Cfg.Core.JWTSecret}.
		CreateJWT(vals.Identification, userID, foundPositionNames, foundTeamNames, onsite, action.jwtDuration)
	resp := PostAuthResponse{Token: jwt}

	writeJSON(w, resp)
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

func (action GetAuth) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	resp := GetAuthResponse{}

	jwtCtx, found := req.Context().Value(JWTContextKey).(JWTContext)
	if !found || jwtCtx.Error != nil || jwtCtx.Claims == nil {
		slog.Error("login failed", "error", jwtCtx.Error)
		resp.Authenticated = false
		writeJSON(w, resp)
		return
	}
	claims := jwtCtx.Claims
	handle := claims.RangerHandle()
	var roles []auth.Role
	if slices.Contains(action.admins, handle) {
		roles = append(roles, auth.Administrator)
	}
	resp.Authenticated = true
	resp.User = handle
	resp.Admin = slices.Contains(roles, auth.Administrator)

	if err := req.ParseForm(); err != nil {
		slog.Error("parseForm error", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	eventName := req.Form.Get("event_id")
	if eventName != "" {

		permissions, err := auth.UserPermissions2(req.Context(), eventName, action.imsDB, action.admins, *claims)
		if err != nil {
			slog.Error("Failed to compute permissions", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		resp.EventAccess = map[string]AccessForEvent{
			eventName: {
				ReadIncidents:     permissions[auth.ReadIncidents],
				WriteIncidents:    permissions[auth.WriteIncidents],
				WriteFieldReports: permissions[auth.WriteFieldReports],
				AttachFiles:       false,
			},
		}
	}

	writeJSON(w, resp)
}
