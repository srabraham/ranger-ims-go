package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"github.com/srabraham/ranger-ims-go/store/queries"
	"log"
	"net/http"
)

type GetEventAccesses struct {
	imsDB *sql.DB
}

func (hand GetEventAccesses) getEventAccesses(w http.ResponseWriter, req *http.Request) {

	eventName := req.PathValue("eventName")

	resp, err := GetEventsAccess(req.Context(), hand.imsDB, eventName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	jjj, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jjj)
}

func GetEventsAccess(ctx context.Context, imsDB *sql.DB, eventName string) (imsjson.EventsAccess, error) {
	var events []queries.Event
	if eventName != "" {
		eventRow, err := queries.New(imsDB).QueryEventID(ctx, eventName)
		if err != nil {
			return nil, fmt.Errorf("[QueryEventID]: %w", err)
		}
		events = append(events, eventRow.Event)
	} else {
		allEventRows, err := queries.New(imsDB).Events(ctx)
		if err != nil {
			return nil, fmt.Errorf("[Events]: %w", err)
		}
		for _, aer := range allEventRows {
			events = append(events, aer.Event)
		}
	}

	resp := make(imsjson.EventsAccess)

	for _, e := range events {
		accessRows, err := queries.New(imsDB).EventAccess(ctx, e.ID)
		if err != nil {
			return nil, fmt.Errorf("[EventAccess]: %w", err)
		}
		ea := imsjson.EventAccess{
			Readers:   []imsjson.AccessRule{},
			Writers:   []imsjson.AccessRule{},
			Reporters: []imsjson.AccessRule{},
		}
		for _, accessRow := range accessRows {
			access := accessRow.EventAccess
			rule := imsjson.AccessRule{Expression: access.Expression, Validity: string(access.Validity)}
			switch access.Mode {
			case queries.EventAccessModeRead:
				ea.Readers = append(ea.Readers, rule)
			case queries.EventAccessModeWrite:
				ea.Writers = append(ea.Writers, rule)
			case queries.EventAccessModeReport:
				ea.Reporters = append(ea.Reporters, rule)
			}
		}
		resp[e.Name] = ea
	}
	return resp, nil
}
