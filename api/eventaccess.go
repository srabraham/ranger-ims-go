package api

import (
	"database/sql"
	"encoding/json"
	"github.com/srabraham/ranger-ims-go/api/access"
	"log"
	"net/http"
)

type GetEventAccesses struct {
	imsDB *sql.DB
}

func (gea GetEventAccesses) getEventAccesses(w http.ResponseWriter, req *http.Request) {

	eventName := req.PathValue("eventName")

	resp, err := access.GetEventsAccess(req.Context(), gea.imsDB, eventName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	jjj, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jjj)
	//
	//var events []queries.Event
	//
	//if eventName != "" {
	//	eventID, err := queries.New(gea.imsDB).QueryEventID(req.Context(), eventName)
	//	if err != nil {
	//		w.WriteHeader(http.StatusInternalServerError)
	//		log.Println(err)
	//		return
	//	}
	//	events = append(events, queries.Event{
	//		ID:   eventID,
	//		Name: eventName,
	//	})
	//} else {
	//	allEvents, err := queries.New(gea.imsDB).Events(req.Context())
	//	if err != nil {
	//		w.WriteHeader(http.StatusInternalServerError)
	//		log.Println(err)
	//		return
	//	}
	//	events = append(events, allEvents...)
	//}
	//
	//resp := make(imsjson.EventsAccess)
	//
	//for _, e := range events {
	//	accesses, err := queries.New(gea.imsDB).EventAccess(req.Context(), e.ID)
	//	if err != nil {
	//		log.Println(err)
	//		return
	//	}
	//	ea := imsjson.EventAccess{}
	//	for _, access := range accesses {
	//		rule := imsjson.AccessRule{Expression: access.Expression, Validity: string(access.Validity)}
	//		switch access.Mode {
	//		case queries.EventAccessModeRead:
	//			ea.Readers = append(ea.Readers, rule)
	//		case queries.EventAccessModeWrite:
	//			ea.Writers = append(ea.Writers, rule)
	//		case queries.EventAccessModeReport:
	//			ea.Reporters = append(ea.Reporters, rule)
	//		}
	//	}
	//	resp[e.Name] = ea
	//}

}
