package api

import (
	"database/sql"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"github.com/srabraham/ranger-ims-go/store/imsdb"
	"log/slog"
	"net/http"
)

type GetStreets struct {
	imsDB *sql.DB
}

func (action GetStreets) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// eventName --> street ID --> street name
	resp := make(imsjson.EventsStreets)

	if ok := mustParseForm(w, req); !ok {
		return
	}
	eventName := req.Form.Get("event_id")
	var events []imsdb.Event
	if eventName != "" {
		event, ok := mustEventFromFormValue(w, req, action.imsDB)
		if !ok {
			return
		}
		events = append(events, imsdb.Event{ID: event.ID, Name: event.Name})
	} else {
		eventRows, err := imsdb.New(action.imsDB).Events(req.Context())
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
		streets, err := imsdb.New(action.imsDB).ConcentricStreets(req.Context(), event.ID)
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
	w.Header().Set("Cache-Control", "max-age=1200, private")
	mustWriteJSON(w, resp)
}

type EditStreets struct {
	imsDB *sql.DB
}

func (action EditStreets) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	eventsStreets, ok := mustReadBodyAs[imsjson.EventsStreets](w, req)
	if !ok {
		return
	}
	for eventName, newEventStreets := range eventsStreets {
		event, ok := mustGetEvent(w, req, eventName, action.imsDB)
		if !ok {
			return
		}
		currentStreets, err := imsdb.New(action.imsDB).ConcentricStreets(req.Context(), event.ID)
		if err != nil {
			slog.Error("Failed to get streets", "error", err)
			http.Error(w, "Failed to get streets", http.StatusInternalServerError)
			return
		}
		currentStreetIDs := make(map[string]bool)
		for _, street := range currentStreets {
			currentStreetIDs[street.ConcentricStreet.ID] = true
		}
		for streetID, streetName := range newEventStreets {
			if !currentStreetIDs[streetID] {
				err = imsdb.New(action.imsDB).CreateConcentricStreet(ctx, imsdb.CreateConcentricStreetParams{
					Event: event.ID,
					ID:    streetID,
					Name:  streetName,
				})
				if err != nil {
					slog.Error("Failed to create concentric street", "error", err)
					http.Error(w, "Failed to create concentric street", http.StatusInternalServerError)
					return
				}
			}
		}
	}
	http.Error(w, "Success", http.StatusNoContent)
}
