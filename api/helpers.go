package api

import (
	"encoding/json"
	"github.com/srabraham/ranger-ims-go/auth"
	"github.com/srabraham/ranger-ims-go/store"
	"github.com/srabraham/ranger-ims-go/store/imsdb"
	"io"
	"log/slog"
	"net/http"
)

func mustParseForm(w http.ResponseWriter, req *http.Request) (success bool) {
	if err := req.ParseForm(); err != nil {
		slog.Error("Failed to parse form", "error", err, "path", req.URL.Path)
		http.Error(w, "Failed to parse HTTP form", http.StatusBadRequest)
		return false
	}
	return true
}

func mustReadBodyAs[T any](w http.ResponseWriter, req *http.Request) (t T, success bool) {
	defer req.Body.Close()
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		slog.Error("Failed to read request body", "error", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return t, false
	}
	if err = json.Unmarshal(bodyBytes, &t); err != nil {
		slog.Error("Failed to unmarshal request body", "error", err)
		http.Error(w, "Failed to unmarshal request body", http.StatusBadRequest)
		return t, false
	}
	return t, true
}

//func eventFromPathValue(w http.ResponseWriter, req *http.Request, imsDB *sql.DB) (event queries.Event, success bool) {
//	eventName := req.PathValue("eventName")
//	if eventName == "" {
//		slog.Error("No eventName was found in the URL path", "path", req.URL.Path)
//		http.Error(w, "No eventName was found in the URL", http.StatusBadRequest)
//		return queries.Event{}, false
//	}
//	eventRow, err := queries.New(imsDB).QueryEventID(req.Context(), eventName)
//	if err != nil {
//		slog.Error("Failed to get event ID", "error", err)
//		http.Error(w, "Failed to get event ID", http.StatusInternalServerError)
//		return queries.Event{}, false
//	}
//	return eventRow.Event, true
//}

func mustEventFromFormValue(w http.ResponseWriter, req *http.Request, imsDB *store.DB) (event imsdb.Event, success bool) {
	if ok := mustParseForm(w, req); !ok {
		return imsdb.Event{}, false
	}
	eventName := req.FormValue("event_id")
	if eventName == "" {
		slog.Error("No event_id was found in the URL path", "path", req.URL.Path)
		http.Error(w, "No event_id was found in the URL", http.StatusBadRequest)
		return imsdb.Event{}, false
	}
	eventRow, err := imsdb.New(imsDB).QueryEventID(req.Context(), eventName)
	if err != nil {
		slog.Error("Failed to get event ID", "error", err)
		http.Error(w, "Failed to get event ID", http.StatusInternalServerError)
		return imsdb.Event{}, false
	}
	return eventRow.Event, true
}

func mustGetEvent(w http.ResponseWriter, req *http.Request, eventName string, imsDB *store.DB) (event imsdb.Event, success bool) {
	if eventName == "" {
		slog.Error("No eventName was provided")
		http.Error(w, "No eventName was provided", http.StatusInternalServerError)
		return imsdb.Event{}, false
	}

	eventRow, err := imsdb.New(imsDB).QueryEventID(req.Context(), eventName)
	if err != nil {
		slog.Error("Failed to fetch event", "error", err)
		http.Error(w, "Event not found", http.StatusNotFound)
		return imsdb.Event{}, false
	}
	return eventRow.Event, true
}

func mustWriteJSON(w http.ResponseWriter, resp any) (success bool) {
	marshalled, err := json.Marshal(resp)
	if err != nil {
		slog.Error("Failed to marshal JSON", "error", err)
		http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
		return false
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(marshalled)
	if err != nil {
		slog.Error("Failed to write JSON", "error", err)
		http.Error(w, "Failed to write JSON", http.StatusInternalServerError)
		return false
	}
	return true
}

func mustGetJwtCtx(w http.ResponseWriter, req *http.Request) (JWTContext, bool) {
	jwtCtx, found := req.Context().Value(JWTContextKey).(JWTContext)
	if !found {
		slog.Error("the ExtractJWTOptionalAuthN adapter must be called before ExtractJWTRequireAuthN")
		http.Error(w, "This endpoint has been misconfigured. Please report this to the tech team",
			http.StatusInternalServerError)
		return JWTContext{}, false
	}
	return jwtCtx, true
}

func mustGetPermissionsCtx(w http.ResponseWriter, req *http.Request) (PermissionsContext, bool) {
	permsCtx, found := req.Context().Value(PermissionsContextKey).(PermissionsContext)
	if !found {
		slog.Error("the ExtractPermissionsToContext adapter must be called before mustGetPermissionsCtx")
		http.Error(w, "This endpoint has been misconfigured. Please report this to the tech team",
			http.StatusInternalServerError)
		return PermissionsContext{}, false
	}
	return permsCtx, true
}

func mustGetEventPermissions(w http.ResponseWriter, req *http.Request, imsDB *store.DB, imsAdmins []string) (imsdb.Event, JWTContext, auth.EventPermissionMask, bool) {
	event, ok := mustGetEvent(w, req, req.PathValue("eventName"), imsDB)
	if !ok {
		return imsdb.Event{}, JWTContext{}, 0, false
	}
	jwtCtx, ok := mustGetJwtCtx(w, req)
	if !ok {
		return imsdb.Event{}, JWTContext{}, 0, false
	}
	eventPermissions, _, err := auth.UserPermissions2(req.Context(), &event.ID, imsDB, imsAdmins, *jwtCtx.Claims)
	if err != nil {
		slog.Error("Failed to compute permissions", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return imsdb.Event{}, JWTContext{}, 0, false
	}
	return event, jwtCtx, eventPermissions[event.ID], true
}
