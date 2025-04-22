package api

import (
	"database/sql"
	clubhousequeries "github.com/srabraham/ranger-ims-go/directory/clubhousedb"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"log/slog"
	"net/http"
)

type GetPersonnel struct {
	clubhouseDB *sql.DB
}

type GetPersonnelResponse []imsjson.Person

func (action GetPersonnel) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	response := make(GetPersonnelResponse, 0)

	// TODO: cache the personnel

	results, err := clubhousequeries.New(action.clubhouseDB).RangersById(req.Context())
	if err != nil {
		slog.Error("Error getting Rangers", "error", err)
		http.Error(w, "Failed to get Rangers", http.StatusInternalServerError)
		return
	}

	for _, r := range results {
		response = append(response, imsjson.Person{
			Handle:      r.Callsign,
			Status:      string(r.Status),
			Onsite:      r.OnSite,
			DirectoryID: r.ID,
		})
	}
	w.Header().Set("Cache-Control", "max-age=1200, private")
	writeJSON(w, response)
}
