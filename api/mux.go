package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/srabraham/ranger-ims-go/conf"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"github.com/srabraham/ranger-ims-go/store/queries"
	"log"
	"net/http"
	"time"
)

func AddToMux(mux *http.ServeMux, db, clubhouseDB *sql.DB) *http.ServeMux {
	// TODO: this is clearly wrong
	mux.HandleFunc("POST /ims/api/auth", PostAuth{
		clubhouseDB: clubhouseDB,
		jwtSecret:   conf.Cfg.Core.JWTSecret,
		jwtDuration: time.Duration(conf.Cfg.Core.TokenLifetime) * time.Second,
	}.postAuth)
	mux.HandleFunc("GET /ims/api/auth", GetAuth{
		clubhouseDB: clubhouseDB,
		jwtSecret:   conf.Cfg.Core.JWTSecret,
	}.getAuth)

	mux.HandleFunc("GET /ims/api/events/{eventName}/incidents/{$}", func(w http.ResponseWriter, req *http.Request) {

		event := req.PathValue("eventName")

		generatedLTE := false // false should mean to exclude

		reportEntries, err := queries.New(db).Incidents_ReportEntries(req.Context(), queries.Incidents_ReportEntriesParams{Name: event, Generated: generatedLTE})
		if err != nil {
			log.Println(err)
			return
		}

		entriesByIncident := make(map[int32][]imsjson.ReportEntry)
		for _, re := range reportEntries {
			entriesByIncident[re.IncidentNumber] = append(entriesByIncident[re.IncidentNumber], imsjson.ReportEntry{
				ID:            re.ID,
				Created:       time.Unix(int64(re.Created), 0),
				Author:        re.Author,
				SystemEntry:   re.Generated,
				Text:          re.Text,
				Stricken:      re.Stricken,
				HasAttachment: re.AttachedFile.String != "",
			})
		}

		rows, err := queries.New(db).Incidents(req.Context(), event)
		if err != nil {
			log.Println(err)
			return
		}

		var resp []imsjson.Incident

		for _, r := range rows {
			//spew.Dump(r)
			var incidentTypes, rangerHandles []string
			var fieldReportNumbers []int32
			json.Unmarshal(r.IncidentTypes.([]byte), &incidentTypes)
			json.Unmarshal(r.RangerHandles.([]byte), &rangerHandles)
			json.Unmarshal(r.FieldReportNumbers.([]byte), &fieldReportNumbers)
			resp = append(resp, imsjson.Incident{
				// TODO: use event from the db, not the request
				Event:   event,
				Number:  r.Incident.Number,
				Created: time.Unix(int64(r.Incident.Created), 0),
				// LastModified
				State:    string(r.Incident.State),
				Priority: r.Incident.Priority,
				Summary:  r.Incident.Summary.String,
				Location: imsjson.Location{
					Name:         r.Incident.LocationName.String,
					Concentric:   r.Incident.LocationConcentric.String,
					RadialHour:   valOrNil(r.Incident.LocationRadialHour),
					RadialMinute: valOrNil(r.Incident.LocationRadialMinute),
					Description:  r.Incident.LocationDescription.String,
					Type:         "garett",
				},
				IncidentTypes: incidentTypes,
				FieldReports:  fieldReportNumbers,
				RangerHandles: rangerHandles,
				ReportEntries: entriesByIncident[r.Incident.Number],
			})
		}

		w.Header().Set("Content-Type", "application/json")
		//jjj, _ := json.MarshalIndent(resp, "", "  ")
		jjj, _ := json.Marshal(resp)
		w.Write(jjj)
		//
		//rows, _ := db.Query("select * from incident")
		//for rows.Next() {
		//	log.Println(rows.Scan())
		//}
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		// The "/" pattern matches everything, so we need to check
		// that we're at the root here.
		if req.URL.Path != "/" {
			http.NotFound(w, req)
			return
		}
		fmt.Fprintf(w, "Welcome to the home page!")
	})
	return mux
}

func valOrNil(v sql.NullInt16) *int16 {
	if v.Valid {
		return &v.Int16
	}
	return nil
}
