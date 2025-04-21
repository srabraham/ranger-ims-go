package api

import (
	"database/sql"
	"encoding/json"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"github.com/srabraham/ranger-ims-go/store/queries"
	"log"
	"log/slog"
	"net/http"
	"slices"
)

type GetIncidentTypes struct {
	imsDB *sql.DB
}

func (hand GetIncidentTypes) getIncidentTypes(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		slog.ErrorContext(req.Context(), "getIncidentTypes failed to parse form", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	includeHidden := false
	if req.Form.Get("hidden") == "true" {
		includeHidden = true
	}
	typeRows, err := queries.New(hand.imsDB).IncidentTypes(req.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	result := make(imsjson.IncidentTypes, 0)
	for _, typeRow := range typeRows {
		t := typeRow.IncidentType
		if includeHidden || !t.Hidden {
			result = append(result, t.Name)
		}
	}
	slices.Sort(result)

	jjj, _ := json.Marshal(result)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jjj)
}
