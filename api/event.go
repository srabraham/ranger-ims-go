package api

import (
	"database/sql"
	"encoding/json"
	"github.com/srabraham/ranger-ims-go/auth"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"github.com/srabraham/ranger-ims-go/store/imsdb"
	"log/slog"
	"net/http"
	"regexp"
)

type GetEvents struct {
	imsDB     *sql.DB
	imsAdmins []string
}

func (action GetEvents) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	resp := make(imsjson.Events, 0)
	ctx := req.Context()

	eventRows, err := imsdb.New(action.imsDB).Events(ctx)
	if err != nil {
		slog.Error("Failed to get events", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// TODO: need to apply authorization per event

	claims := ctx.Value(JWTContextKey).(JWTContext).Claims

	for _, er := range eventRows {
		perms, err := auth.UserPermissions2(
			ctx,
			er.Event.Name,
			action.imsDB,
			action.imsAdmins,
			*claims,
		)
		if err != nil {
			slog.Error("UserPermissions", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if perms[auth.ReadEventName] {
			resp = append(resp, imsjson.Event{
				// TODO: eventually change this to actually be the numeric ID
				ID:   er.Event.Name,
				Name: er.Event.Name,
			})
		}
	}
	w.Header().Set("Cache-Control", "max-age=1200, private")
	writeJSON(w, resp)
}

type EditEvents struct {
	imsDB     *sql.DB
	imsAdmins []string
}

// Require basic cleanliness for EventID, since it's used in IMS URLs
// and in filesystem directory paths.
var allowedEventNames = regexp.MustCompile(`^[\w-]+$`)

func (action EditEvents) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	if ok := parseForm(w, req); !ok {
		return
	}
	bodyBytes, ok := readBody(w, req)
	if !ok {
		return
	}
	var editRequest imsjson.EditEventsRequest
	if err := json.Unmarshal(bodyBytes, &editRequest); err != nil {
		slog.Error("Failed to parse body", "error", err)
		http.Error(w, "Failed to parse body", http.StatusBadRequest)
		return
	}

	for _, eventName := range editRequest.Add {
		if !allowedEventNames.MatchString(eventName) {
			slog.Error("Invalid event name", "eventName", eventName)
			http.Error(w, "Event names must match the pattern "+allowedEventNames.String(), http.StatusBadRequest)
			return
		}
		id, err := imsdb.New(action.imsDB).CreateEvent(ctx, eventName)
		if err != nil {
			slog.Error("Failed to create event", "eventName", eventName, "error", err)
			http.Error(w, "Failed to create event", http.StatusInternalServerError)
			return
		}
		slog.Info("Created event", "eventName", eventName, "id", id)
	}
	http.Error(w, "Success", http.StatusNoContent)
}
