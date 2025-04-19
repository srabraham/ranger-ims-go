package api

import (
	"database/sql"
	"encoding/json"
	"github.com/srabraham/ranger-ims-go/store/queries"
	"log"
	"net/http"
)

type GetIncidentTypes struct {
	imsDB *sql.DB
}

func (ga GetIncidentTypes) getIncidentTypes(w http.ResponseWriter, req *http.Request) {
	// TODO: need the "hidden" feature

	types, err := queries.New(ga.imsDB).IncidentTypes(req.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	jjj, _ := json.Marshal(types)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jjj)
}
