package api

import (
	"database/sql"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"github.com/srabraham/ranger-ims-go/store/queries"
	"log/slog"
	"net/http"
)

type GetStreets struct {
	imsDB *sql.DB
}

func (hand GetStreets) getStreets(w http.ResponseWriter, req *http.Request) {
	// eventName --> street ID --> street name
	resp := make(imsjson.EventsStreets)

	if ok := parseForm(w, req); !ok {
		return
	}
	eventName := req.Form.Get("event_id")
	var events []queries.Event
	if eventName != "" {
		event, ok := eventFromFormValue(w, req, hand.imsDB)
		if !ok {
			return
		}
		events = append(events, queries.Event{ID: event.ID, Name: event.Name})
	} else {
		eventRows, err := queries.New(hand.imsDB).Events(req.Context())
		if err != nil {
			slog.Error("Failed to get events", "error", err)
			http.Error(w, "Failed to get events", http.StatusInternalServerError)
			return
		}
		for _, er := range eventRows {
			events = append(events, er.Event)
		}
	}

	for _, event := range events {
		streets, err := queries.New(hand.imsDB).ConcentricStreets(req.Context(), event.ID)
		if err != nil {
			slog.Error("Failed to get streets", "error", err)
			http.Error(w, "Failed to get streets", http.StatusInternalServerError)
			return
		}
		resp[event.Name] = make(imsjson.EventStreets)
		for _, street := range streets {
			resp[event.Name][street.ConcentricStreet.ID] = street.ConcentricStreet.Name
		}
	}
	writeJSON(w, resp)
}
