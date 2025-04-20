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

type GetPersonnelResponse []imsjson.Person

func (hand GetPersonnel) getPersonnel(w http.ResponseWriter, req *http.Request) {

	// TODO: cache the personnel

	results, err := clubhousequeries.New(hand.clubhouseDB).RangersById(req.Context())
	if err != nil {
		slog.ErrorContext(req.Context(), "Error getting rangers", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := make(GetPersonnelResponse, 0)
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
