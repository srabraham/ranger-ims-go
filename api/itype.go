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

func (action GetIncidentTypes) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	response := make(imsjson.IncidentTypes, 0)

	if success := mustParseForm(w, req); !success {
		return
	}
	includeHidden := req.Form.Get("hidden") == "true"
	typeRows, err := imsdb.New(action.imsDB).IncidentTypes(req.Context())
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

	w.Header().Set("Cache-Control", "max-age=1200, private")
	writeJSON(w, response)
}

type EditIncidentTypes struct {
	imsDB *sql.DB
}

func (action EditIncidentTypes) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	typesReq, ok := mustReadBodyAs[imsjson.EditIncidentTypesRequest](w, req)
	if !ok {
		return
	}
	for _, it := range typesReq.Add {
		err := imsdb.New(action.imsDB).CreateIncidentTypeOrIgnore(ctx, imsdb.CreateIncidentTypeOrIgnoreParams{
			Name:   it,
			Hidden: false,
		})
		if err != nil {
			slog.Error("Failed to create incident type", "error", err)
			http.Error(w, "Failed to create incident type", http.StatusInternalServerError)
			return
		}
	}
	for _, it := range typesReq.Hide {
		err := imsdb.New(action.imsDB).HideShowIncidentType(ctx, imsdb.HideShowIncidentTypeParams{
			Name:   it,
			Hidden: true,
		})
		if err != nil {
			slog.Error("Failed to hide incident type", "error", err)
			http.Error(w, "Failed to hide incident type", http.StatusInternalServerError)
			return
		}
	}
	for _, it := range typesReq.Show {
		err := imsdb.New(action.imsDB).HideShowIncidentType(ctx, imsdb.HideShowIncidentTypeParams{
			Name:   it,
			Hidden: false,
		})
		if err != nil {
			slog.Error("Failed to unhide incident type", "error", err)
			http.Error(w, "Failed to unhide incident type", http.StatusInternalServerError)
			return
		}
	}
	http.Error(w, "Success", http.StatusOK)
}
