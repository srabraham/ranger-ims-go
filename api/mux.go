package api

import (
	"context"
	"database/sql"
	"github.com/srabraham/ranger-ims-go/auth"
	"github.com/srabraham/ranger-ims-go/conf"
	"log"
	"log/slog"
	"net/http"
	"time"
)

func AddToMux(mux *http.ServeMux, cfg *conf.IMSConfig, db, clubhouseDB *sql.DB) *http.ServeMux {
	j := auth.JWTer{SecretKey: cfg.Core.JWTSecret}

	mux.Handle("GET /ims/api/access",
		Adapt(
			GetEventAccesses{imsDB: db}.getEventAccesses,
			ExtractClaimsToContext(j),
			RequireAuthenticated(),
		),
	)

	mux.Handle("POST /ims/api/auth",
		Adapt(
			PostAuth{
				imsDB:       db,
				clubhouseDB: clubhouseDB,
				jwtSecret:   cfg.Core.JWTSecret,
				jwtDuration: time.Duration(cfg.Core.TokenLifetime) * time.Second,
			}.postAuth,
		),
		// This endpoint does not require authentication, nor
		// should it even consider current Authorization header.
	)

	mux.Handle("GET /ims/api/auth",
		Adapt(
			GetAuth{
				imsDB:       db,
				clubhouseDB: clubhouseDB,
				jwtSecret:   cfg.Core.JWTSecret,
				admins:      cfg.Core.Admins,
			}.getAuth,
			ExtractClaimsToContext(j),
			// This endpoint does not require authentication
		),
	)

	mux.Handle("GET /ims/api/events/{eventName}/incidents",
		Adapt(
			GetIncidents{imsDB: db}.getIncidents,
			ExtractClaimsToContext(j),
			RequireAuthenticated(),
			RequireAuthorization(auth.ReadIncidents, db, cfg.Core.Admins),
		),
	)

	mux.Handle("GET /ims/api/events/{eventName}/incidents/{incidentNumber}",
		Adapt(
			GetIncident{imsDB: db}.getIncident,
			ExtractClaimsToContext(j),
			RequireAuthenticated(),
		),
	)

	mux.Handle("GET /ims/api/events/{eventName}/field_reports",
		Adapt(
			GetFieldReports{imsDB: db}.getFieldReports,
			ExtractClaimsToContext(j),
			RequireAuthenticated(),
			// TODO: need a READ permission
			RequireAuthorization(auth.ReadOwnFieldReports, db, cfg.Core.Admins),
		),
	)

	mux.Handle("GET /ims/api/events",
		Adapt(
			GetEvents{imsDB: db}.getEvents,
			ExtractClaimsToContext(j),
			RequireAuthenticated(),
		),
	)

	mux.Handle("GET /ims/api/streets",
		Adapt(
			GetStreets{imsDB: db}.getStreets,
			ExtractClaimsToContext(j),
			RequireAuthenticated(),
		),
	)

	mux.Handle("GET /ims/api/incident_types",
		Adapt(
			GetIncidentTypes{imsDB: db}.getIncidentTypes,
			ExtractClaimsToContext(j),
			RequireAuthenticated(),
		),
	)

	mux.Handle("GET /ims/api/personnel",
		Adapt(
			GetPersonnel{clubhouseDB: clubhouseDB}.getPersonnel,
			ExtractClaimsToContext(j),
			RequireAuthenticated(),
		),
	)

	mux.HandleFunc("GET /ims/api/ping",
		func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ack"))
		},
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

type ContextKey string

const JWTContextKey ContextKey = "JWTContext"

type JWTContext struct {
	Claims *auth.IMSClaims
	Error  error
}

func ExtractClaimsToContext(j auth.JWTer) Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			//if header == "" {
			//	w.WriteHeader(http.StatusUnauthorized)
			//	w.Write([]byte("No Authorization token was provided"))
			//	return
			//}
			claims, err := j.AuthenticateJWT(header)

			ctx := context.WithValue(r.Context(), JWTContextKey, JWTContext{
				Claims: claims,
				Error:  err,
			})

			next.ServeHTTP(w, r.WithContext(ctx))

			//r.WithContext(context.WithValue(r.Context())).Value()
			//if err != nil {
			//	w.WriteHeader(http.StatusUnauthorized)
			//	w.Write([]byte("Invalid Authorization token"))
			//	return
			//}
			//next.ServeHTTP(w, r)
		})
	}
}

func RequireAuthenticated() Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			jwtCtx, found := r.Context().Value(JWTContextKey).(JWTContext)
			if !found {
				slog.Error("the ExtractClaimsToContext adapter must be called before RequireAuthenticated")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if jwtCtx.Error != nil || jwtCtx.Claims == nil {
				slog.Error("JWT error", "error", jwtCtx.Error)
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Invalid Authorization token"))
				return
			}
			slog.Info("in require auth", "claims", jwtCtx.Claims)
			if jwtCtx.Claims.RangerHandle() == "" {
				slog.Error("No Ranger handle in JWT")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Invalid Authorization token"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireAuthorization(required auth.Permission, imsDB *sql.DB, imsAdmins []string) Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			jwtCtx, found := r.Context().Value(JWTContextKey).(JWTContext)
			if !found {
				slog.Error("the ExtractClaimsToContext adapter must be called before RequireAuthenticated")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			//if err := r.ParseForm(); err != nil {
			//	slog.Error("error parsing form", "error", err)
			//	w.WriteHeader(http.StatusBadRequest)
			//	return
			//}
			//permissions, err := auth.UserPermissions2(r.Context(), r.FormValue("event_id"), imsDB, imsAdmins, *jwtCtx.Claims)
			permissions, err := auth.UserPermissions2(r.Context(), r.PathValue("eventName"), imsDB, imsAdmins, *jwtCtx.Claims)
			if err != nil {
				slog.Error("Failed to compute permissions", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			slog.Info("testing authorization", "permissions", permissions, "required", required)
			if !permissions[required] {
				slog.Info("rejected!")
				w.WriteHeader(http.StatusUnauthorized)
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
		handler = adapter(handler)
	}
	return handler
}
