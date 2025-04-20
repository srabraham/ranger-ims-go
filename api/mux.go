package api

import (
	"database/sql"
	"github.com/srabraham/ranger-ims-go/auth"
	"github.com/srabraham/ranger-ims-go/conf"
	"log"
	"net/http"
	"time"
)

func AddToMux(mux *http.ServeMux, cfg *conf.IMSConfig, db, clubhouseDB *sql.DB) *http.ServeMux {
	j := auth.JWTer{SecretKey: cfg.Core.JWTSecret}

	mux.Handle("GET /ims/api/access/{$}",
		Adapt(
			GetEventAccesses{imsDB: db}.getEventAccesses,
			RequireAuthenticated(j),
		),
	)

	mux.HandleFunc("POST /ims/api/auth/{$}",
		PostAuth{
			imsDB:       db,
			clubhouseDB: clubhouseDB,
			jwtSecret:   cfg.Core.JWTSecret,
			jwtDuration: time.Duration(cfg.Core.TokenLifetime) * time.Second,
		}.postAuth,
	)

	mux.HandleFunc("GET /ims/api/auth/{$}",
		GetAuth{
			imsDB:       db,
			clubhouseDB: clubhouseDB,
			jwtSecret:   cfg.Core.JWTSecret,
			admins:      cfg.Core.Admins,
		}.getAuth,
	)

	mux.Handle("GET /ims/api/events/{eventName}/incidents/{$}",
		Adapt(
			GetIncidents{imsDB: db}.getIncidents,
			RequireAuthenticated(j),
		),
	)

	mux.Handle("GET /ims/api/events/{eventName}/field_reports/{$}",
		Adapt(
			GetFieldReports{imsDB: db}.getFieldReports,
			RequireAuthenticated(j),
		),
	)

	mux.Handle("GET /ims/api/events/{$}",
		Adapt(
			GetEvents{imsDB: db}.getEvents,
			RequireAuthenticated(j),
		),
	)

	mux.Handle("GET /ims/api/streets/{$}",
		Adapt(
			GetStreets{imsDB: db}.getStreets,
			RequireAuthenticated(j),
		),
	)

	mux.Handle("GET /ims/api/incident_types/{$}",
		Adapt(
			GetIncidentTypes{imsDB: db}.getIncidentTypes,
			RequireAuthenticated(j),
		),
	)

	//mux.HandleFunc("GET /ims/api/events/{$}",
	//	GetEvents{imsDB: db}.getEvents,
	//)
	return mux
}

func int32OrNil(v sql.NullInt32) *int32 {
	if v.Valid {
		return &v.Int32
	}
	return nil
}

func int16OrNil(v sql.NullInt16) *int16 {
	if v.Valid {
		return &v.Int16
	}
	return nil
}

func stringOrNil(v sql.NullString) *string {
	if v.Valid {
		return &v.String
	}
	return nil
}

func ptr[T any](t T) *T {
	return &t
}

type Adapter func(http.Handler) http.Handler

func LogBeforeAfter() Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println("before")
			defer log.Println("after, with defer")
			next.ServeHTTP(w, r)
			log.Println("called ServeHTTP")
		})
	}
}

func RequireAuthenticated(j auth.JWTer) Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("No Authorization token was provided"))
				return
			}
			_, err := j.AuthenticateJWT(header)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Invalid Authorization token"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func Adapt(h http.HandlerFunc, adapters ...Adapter) http.Handler {
	handler := http.Handler(h)
	for i := range adapters {
		adapter := adapters[len(adapters)-1-i] // range in reverse
		handler = adapter(h)
	}
	return handler
}
