package api

import (
	"database/sql"
	"github.com/srabraham/ranger-ims-go/store/queries"
	"log/slog"
	"net/http"
)

func parseForm(w http.ResponseWriter, req *http.Request) (success bool) {
	if err := req.ParseForm(); err != nil {
		slog.Error("Failed to parse form", "error", err, "path", req.URL.Path)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to parse HTTP form"))
		return false
	}
	return true
}

func eventFromPathValue(w http.ResponseWriter, req *http.Request, imsDB *sql.DB) (event queries.Event, success bool) {
	eventName := req.PathValue("eventName")
	if eventName == "" {
		slog.Error("No eventName was found in the URL path", "path", req.URL.Path)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("No eventName was found in the URL"))
		return queries.Event{}, false
	}
	eventRow, err := queries.New(imsDB).QueryEventID(req.Context(), eventName)
	if err != nil {
		slog.Error("Failed to get event ID", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return queries.Event{}, false
	}
	return eventRow.Event, true
}
