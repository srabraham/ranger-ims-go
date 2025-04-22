package api

import (
	"database/sql"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"github.com/srabraham/ranger-ims-go/store/imsdb"
	"log/slog"
	"net/http"
	"slices"
)

type GetIncidentTypes struct {
	imsDB *sql.DB
}

func (hand GetIncidentTypes) getIncidentTypes(w http.ResponseWriter, req *http.Request) {
	response := make(imsjson.IncidentTypes, 0)

	if success := parseForm(w, req); !success {
		return
	}
	includeHidden := req.Form.Get("hidden") == "true"
	typeRows, err := imsdb.New(hand.imsDB).IncidentTypes(req.Context())
	if err != nil {
		slog.Error("IncidentTypes query", "error", err)
		http.Error(w, "IncidentTypes query failed", http.StatusInternalServerError)
		return
	}

	for _, typeRow := range typeRows {
		t := typeRow.IncidentType
		if includeHidden || !t.Hidden {
			response = append(response, t.Name)
		}
	}
	slices.Sort(response)

	writeJSON(w, response)
}
