package api

import (
	"database/sql"
	"encoding/json"
	"github.com/srabraham/ranger-ims-go/store/queries"
	"log"
	"log/slog"
	"net/http"
	"slices"
)

type GetIncidentTypes struct {
	imsDB *sql.DB
}

type GetIncidentTypesResponse []string

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
	types, err := queries.New(hand.imsDB).IncidentTypes(req.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	result := make(GetIncidentTypesResponse, 0)
	for _, t := range types {
		if includeHidden || !t.Hidden {
			result = append(result, t.Name)
		}
	}
	slices.Sort(result)

	jjj, _ := json.Marshal(result)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jjj)
}
