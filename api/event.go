package api

import (
	"database/sql"
	"encoding/json"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"github.com/srabraham/ranger-ims-go/store/queries"
	"log"
	"net/http"
)

type GetEvents struct {
	imsDB *sql.DB
}

func (ge GetEvents) getEvents(w http.ResponseWriter, req *http.Request) {
	events, err := queries.New(ge.imsDB).Events(req.Context())
	if err != nil {
		log.Println(err)
		return
	}

	// TODO: need to apply authorization per event

	var resp []imsjson.Event

	for _, e := range events {
		resp = append(resp, imsjson.Event{
			// TODO: eventually change this to actually be the numeric ID
			ID:   e.Name,
			Name: e.Name,
		})
	}

	jjj, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jjj)
	log.Println("returned events")
}
