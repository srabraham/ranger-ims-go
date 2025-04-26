package api

import (
	"github.com/srabraham/ranger-ims-go/auth"
	"github.com/srabraham/ranger-ims-go/conf"
	"github.com/srabraham/ranger-ims-go/directory"
	"github.com/srabraham/ranger-ims-go/store"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"
)

type PostAuth struct {
	imsDB       *store.DB
	userStore   *directory.UserStore
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
	// This endpoint is unauthenticated (doesn't require an Authorization header)
	// as the point of this is to take a username and password to create a new JWT.

	vals, ok := mustReadBodyAs[PostAuthRequest](w, req)
	if !ok {
		return
	}

	rangers, err := action.userStore.GetRangers(req.Context())
	if err != nil {
		slog.Error("Failed to get personnel", "error", err)
		http.Error(w, "Failed to get personnel", http.StatusInternalServerError)
		return
	}
	var storedPassHash string
	var userID int64
	var onsite bool
	for _, person := range rangers {
		callsignMatch := person.Handle == vals.Identification
		emailMatch := person.Email != "" && strings.ToLower(person.Email) == strings.ToLower(vals.Identification)
		if callsignMatch || emailMatch {
			userID = person.DirectoryID
			storedPassHash = person.Password
			onsite = person.Onsite
			break
		}
	}

	correct, _ := auth.VerifyPassword(vals.Password, storedPassHash)
	if correct {
		slog.Error("Failed login attempt (bad credentials)", "identification", vals.Identification)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	slog.Info("Successful login", "identification", vals.Identification)

	foundPositionNames, foundTeamNames, err := action.userStore.GetUserPositionsTeams(req.Context(), userID)
	if err != nil {
		slog.Error("Failed to fetch Clubhouse positions/teams data", "error", err)
		http.Error(w, "Failed to fetch Clubhouse positions/teams data", http.StatusInternalServerError)
		return
	}

	jwt := auth.JWTer{SecretKey: conf.Cfg.Core.JWTSecret}.
		CreateJWT(vals.Identification, userID, foundPositionNames, foundTeamNames, onsite, action.jwtDuration)
	resp := PostAuthResponse{Token: jwt}

	mustWriteJSON(w, resp)
}

type GetAuth struct {
	imsDB     *store.DB
	jwtSecret string
	admins    []string
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
	// This endpoint is unauthenticated (doesn't require an Authorization header).
	resp := GetAuthResponse{}

	jwtCtx, found := req.Context().Value(JWTContextKey).(JWTContext)
	if !found || jwtCtx.Error != nil || jwtCtx.Claims == nil {
		resp.Authenticated = false
		mustWriteJSON(w, resp)
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
		slog.Error("mustParseForm error", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	eventName := req.Form.Get("event_id")
	if eventName != "" {

		event, ok := mustEventFromFormValue(w, req, action.imsDB)
		if !ok {
			return
		}

		eventPermissions, _, err := auth.EventPermissions(req.Context(), &event.ID, action.imsDB, action.admins, *claims)
		if err != nil {
			slog.Error("Failed to compute permissions", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		resp.EventAccess = map[string]AccessForEvent{
			eventName: {
				ReadIncidents:     eventPermissions[event.ID]&auth.EventReadIncidents != 0,
				WriteIncidents:    eventPermissions[event.ID]&auth.EventWriteIncidents != 0,
				WriteFieldReports: eventPermissions[event.ID]&(auth.EventWriteOwnFieldReports|auth.EventWriteAllFieldReports) != 0,
				AttachFiles:       false,
			},
		}
	}

	mustWriteJSON(w, resp)
}
