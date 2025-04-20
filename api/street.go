package api

import (
	"database/sql"
	"encoding/json"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"github.com/srabraham/ranger-ims-go/store/queries"
	"log"
	"net/http"
)

type GetStreets struct {
	imsDB *sql.DB
}

func (ga GetStreets) getStreets(w http.ResponseWriter, req *http.Request) {
	// TODO: check if event_id set, use that if so

	events, err := queries.New(ga.imsDB).Events(req.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	// eventName --> street ID --> street name

	resp := make(imsjson.EventsStreets)

	for _, event := range events {
		streets, err := queries.New(ga.imsDB).ConcentricStreets(req.Context(), event.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
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
