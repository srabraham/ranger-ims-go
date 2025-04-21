package api

import (
	"database/sql"
	"encoding/json"
	"github.com/srabraham/ranger-ims-go/auth"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"github.com/srabraham/ranger-ims-go/store/queries"
	"log"
	"log/slog"
	"net/http"
)

type GetEvents struct {
	imsDB     *sql.DB
	imsAdmins []string
}

type GetEventsResponse []imsjson.Event

func (hand GetEvents) getEvents(w http.ResponseWriter, req *http.Request) {
	eventRows, err := queries.New(hand.imsDB).Events(req.Context())
	if err != nil {
		slog.Error("Failed to get events", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// TODO: need to apply authorization per event

	resp := make(GetEventsResponse, 0)

	claims := req.Context().Value(JWTContextKey).(JWTContext).Claims

	for _, er := range eventRows {
		perms, err := auth.UserPermissions2(
			req.Context(),
			er.Event.Name,
			hand.imsDB,
			hand.imsAdmins,
			*claims,
		)
		if err != nil {
			slog.Error("UserPermissions", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if perms[auth.ReadEventName] {
			resp = append(resp, imsjson.Event{
				// TODO: eventually change this to actually be the numeric ID
				ID:   er.Event.Name,
				Name: er.Event.Name,
			})
		}
	}

	jjj, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jjj)
	log.Println("returned events")
}
