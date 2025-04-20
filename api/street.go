package api

import (
	"database/sql"
	"encoding/json"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"github.com/srabraham/ranger-ims-go/store/queries"
	"log/slog"
	"net/http"
)

type GetStreets struct {
	imsDB *sql.DB
}

type GetStreetsResponse map[string]imsjson.EventStreets

func (hand GetStreets) getStreets(w http.ResponseWriter, req *http.Request) {
	var events []queries.Event
	if err := req.ParseForm(); err != nil {
		slog.Error("ParseForm: ", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	eventName := req.Form.Get("event_id")
	if eventName != "" {
		eventID, err := queries.New(hand.imsDB).QueryEventID(req.Context(), eventName)
		if err != nil {
			slog.Error("Failed to get event ID", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		events = append(events, queries.Event{ID: eventID, Name: eventName})
	} else {
		eventsFromDB, err := queries.New(hand.imsDB).Events(req.Context())
		if err != nil {
			slog.Error("Failed to get events", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		events = append(events, eventsFromDB...)
	}

	// eventName --> street ID --> street name
	resp := make(GetStreetsResponse)

	for _, event := range events {
		streets, err := queries.New(hand.imsDB).ConcentricStreets(req.Context(), event.ID)
		if err != nil {
			slog.Error("Failed to get streets", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp[event.Name] = make(imsjson.EventStreets)
		for _, street := range streets {
			resp[event.Name][street.ID] = street.Name
		}
	}
	jjj, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jjj)
}
