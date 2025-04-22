package api

import (
	"database/sql"
	"encoding/json"
	"github.com/srabraham/ranger-ims-go/store/imsdb"
	"io"
	"log/slog"
	"net/http"
)

func parseForm(w http.ResponseWriter, req *http.Request) (success bool) {
	if err := req.ParseForm(); err != nil {
		slog.Error("Failed to parse form", "error", err, "path", req.URL.Path)
		http.Error(w, "Failed to parse HTTP form", http.StatusBadRequest)
		return false
	}
	return true
}

func readBody(w http.ResponseWriter, req *http.Request) (bytes []byte, success bool) {
	defer req.Body.Close()
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		slog.Error("Failed to read request body", "error", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return nil, false
	}
	return bodyBytes, true
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

func eventFromFormValue(w http.ResponseWriter, req *http.Request, imsDB *sql.DB) (event imsdb.Event, success bool) {
	if ok := parseForm(w, req); !ok {
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

func eventFromName(w http.ResponseWriter, req *http.Request, eventName string, imsDB *sql.DB) (event imsdb.Event, success bool) {
	if eventName == "" {
		slog.Error("No eventName was provided")
		http.Error(w, "No eventName was provided", http.StatusInternalServerError)
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

func writeJSON(w http.ResponseWriter, resp any) (success bool) {
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
