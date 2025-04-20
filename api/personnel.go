package api

import (
	"database/sql"
	"encoding/json"
	clubhousequeries "github.com/srabraham/ranger-ims-go/directory/queries"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"log/slog"
	"net/http"
)

type GetPersonnel struct {
	clubhouseDB *sql.DB
}

func (gp GetPersonnel) getPersonnel(w http.ResponseWriter, req *http.Request) {
	//if err := req.ParseForm(); err != nil {
	//	slog.ErrorContext(req.Context(), "getPersonnel failed to parse form", "error", err)
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}
	//includeHidden := false
	//if req.Form.Get("hidden") == "true" {
	//	includeHidden = true
	//}
	results, err := clubhousequeries.New(gp.clubhouseDB).RangersById(req.Context())
	if err != nil {
		slog.ErrorContext(req.Context(), "Error getting rangers", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := make(imsjson.Personnel, 0)
	for _, r := range results {
		response = append(response, imsjson.Person{
			Handle:      r.Callsign,
			Status:      string(r.Status),
			Onsite:      r.OnSite,
			DirectoryID: r.ID,
		})
	}

	jjj, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jjj)
}
