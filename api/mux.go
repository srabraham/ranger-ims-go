package api

import (
	"context"
	"database/sql"
	"github.com/srabraham/ranger-ims-go/auth"
	"github.com/srabraham/ranger-ims-go/conf"
	"log/slog"
	"net/http"
	"time"
)

func AddToMux(mux *http.ServeMux, cfg *conf.IMSConfig, db, clubhouseDB *sql.DB) *http.ServeMux {
	if mux == nil {
		mux = http.NewServeMux()
	}

	j := auth.JWTer{SecretKey: cfg.Core.JWTSecret}

	es := NewEventSourcerer()

	mux.Handle("GET /ims/api/access",
		Adapt(
			GetEventAccesses{imsDB: db},
			LogBeforeAfter(),
			ExtractClaimsToContext(j),
			RequireAuthenticated(),
			RequireAuthorization(auth.AdministrateEvents, db, cfg.Core.Admins),
		),
	)

	mux.Handle("POST /ims/api/access",
		Adapt(
			PostEventAccess{imsDB: db},
			LogBeforeAfter(),
			ExtractClaimsToContext(j),
			RequireAuthenticated(),
			RequireAuthorization(auth.AdministrateEvents, db, cfg.Core.Admins),
		),
	)

	mux.Handle("POST /ims/api/auth",
		Adapt(
			PostAuth{
				imsDB:       db,
				clubhouseDB: clubhouseDB,
				jwtSecret:   cfg.Core.JWTSecret,
				jwtDuration: time.Duration(cfg.Core.TokenLifetime) * time.Second,
			},
			LogBeforeAfter(),
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
			},
			LogBeforeAfter(),
			ExtractClaimsToContext(j),
			// This endpoint does not require authentication or authorization, by design
		),
	)

	mux.Handle("GET /ims/api/events/{eventName}/incidents",
		Adapt(
			GetIncidents{imsDB: db},
			LogBeforeAfter(),
			ExtractClaimsToContext(j),
			RequireAuthenticated(),
			RequireAuthorization(auth.ReadIncidents, db, cfg.Core.Admins),
		),
	)

	mux.Handle("GET /ims/api/events/{eventName}/incidents/{incidentNumber}",
		Adapt(
			GetIncident{imsDB: db},
			LogBeforeAfter(),
			ExtractClaimsToContext(j),
			RequireAuthenticated(),
			RequireAuthorization(auth.ReadIncidents, db, cfg.Core.Admins),
		),
	)

	mux.Handle("POST /ims/api/events/{eventName}/incidents/{incidentNumber}/report_entries/{reportEntryId}",
		Adapt(
			EditIncidentReportEntry{imsDB: db, eventSource: es},
			LogBeforeAfter(),
			ExtractClaimsToContext(j),
			RequireAuthenticated(),
			RequireAuthorization(auth.WriteIncidents, db, cfg.Core.Admins),
		),
	)

	mux.Handle("GET /ims/api/events/{eventName}/field_reports",
		Adapt(
			GetFieldReports{imsDB: db},
			LogBeforeAfter(),
			ExtractClaimsToContext(j),
			RequireAuthenticated(),
			RequireAuthorization(auth.ReadFieldReports, db, cfg.Core.Admins),
		),
	)

	mux.Handle("POST /ims/api/events/{eventName}/field_reports",
		Adapt(
			NewFieldReport{imsDB: db, eventSource: es},
			LogBeforeAfter(),
			ExtractClaimsToContext(j),
			RequireAuthenticated(),
			RequireAuthorization(auth.WriteFieldReports, db, cfg.Core.Admins),
		),
	)

	mux.Handle("GET /ims/api/events/{eventName}/field_reports/{fieldReportNumber}",
		Adapt(
			GetFieldReport{imsDB: db},
			LogBeforeAfter(),
			ExtractClaimsToContext(j),
			RequireAuthenticated(),
			RequireAuthorization(auth.ReadFieldReports, db, cfg.Core.Admins),
		),
	)

	mux.Handle("POST /ims/api/events/{eventName}/field_reports/{fieldReportNumber}",
		Adapt(
			EditFieldReport{imsDB: db, eventSource: es},
			LogBeforeAfter(),
			ExtractClaimsToContext(j),
			RequireAuthenticated(),
			RequireAuthorization(auth.WriteFieldReports, db, cfg.Core.Admins),
		),
	)

	mux.Handle("POST /ims/api/events/{eventName}/field_reports/{fieldReportNumber}/report_entries/{reportEntryId}",
		Adapt(
			EditFieldReportReportEntry{imsDB: db, eventSource: es},
			LogBeforeAfter(),
			ExtractClaimsToContext(j),
			RequireAuthenticated(),
			RequireAuthorization(auth.WriteFieldReports, db, cfg.Core.Admins),
		),
	)

	mux.Handle("GET /ims/api/events",
		Adapt(
			GetEvents{imsDB: db, imsAdmins: cfg.Core.Admins},
			LogBeforeAfter(),
			ExtractClaimsToContext(j),
			RequireAuthenticated(),
			// ugh, no eventname in path
			//RequireAuthorization(auth.ReadEventName, db, cfg.Core.Admins),
		),
	)

	mux.Handle("POST /ims/api/events",
		Adapt(
			EditEvents{imsDB: db, imsAdmins: cfg.Core.Admins},
			LogBeforeAfter(),
			ExtractClaimsToContext(j),
			RequireAuthenticated(),
			RequireAuthorization(auth.AdministrateEvents, db, cfg.Core.Admins),
		),
	)

	mux.Handle("GET /ims/api/streets",
		Adapt(
			GetStreets{imsDB: db},
			LogBeforeAfter(),
			ExtractClaimsToContext(j),
			RequireAuthenticated(),
			// ugh, no eventname in path
			//RequireAuthorization(auth.ReadEventStreets, db, cfg.Core.Admins),
		),
	)

	mux.Handle("GET /ims/api/incident_types",
		Adapt(
			GetIncidentTypes{imsDB: db},
			LogBeforeAfter(),
			ExtractClaimsToContext(j),
			RequireAuthenticated(),
			RequireAuthorization(auth.ReadIncidentTypes, db, cfg.Core.Admins),
		),
	)

	mux.Handle("GET /ims/api/personnel",
		Adapt(
			GetPersonnel{clubhouseDB: clubhouseDB},
			LogBeforeAfter(),
			ExtractClaimsToContext(j),
			RequireAuthenticated(),
			RequireAuthorization(auth.ReadPersonnel, db, cfg.Core.Admins),
		),
	)

	mux.Handle("GET /ims/api/eventsource",
		Adapt(
			es.Server.Handler(EventSourceChannel),
			LogBeforeAfter(),
		),
	)

	mux.HandleFunc("GET /ims/api/ping",
		func(w http.ResponseWriter, req *http.Request) {
			http.Error(w, "ack", http.StatusOK)
		},
	)

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
			start := time.Now()
			next.ServeHTTP(w, r)
			slog.Info("Done serving request", "duration", time.Since(start).Round(100*time.Microsecond), "method", r.Method, "path", r.URL.Path)
		})
	}
}

type ContextKey string

const JWTContextKey ContextKey = "JWTContext"

type JWTContext struct {
	Claims            *auth.IMSClaims
	AuthenticatedUser string
	Error             error
}

func ExtractClaimsToContext(j auth.JWTer) Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			claims, err := j.AuthenticateJWT(header)
			ctx := context.WithValue(r.Context(), JWTContextKey, JWTContext{
				Claims:            claims,
				AuthenticatedUser: claims.RangerHandle(),
				Error:             err,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAuthenticated() Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			jwtCtx, found := r.Context().Value(JWTContextKey).(JWTContext)
			if !found {
				slog.Error("the ExtractClaimsToContext adapter must be called before RequireAuthenticated")
				http.Error(w, "This endpoint has been misconfigured. Please report this to the tech team",
					http.StatusInternalServerError)
				return
			}
			if jwtCtx.Error != nil || jwtCtx.Claims == nil {
				slog.Error("JWT error", "error", jwtCtx.Error)
				http.Error(w, "Invalid Authorization token", http.StatusUnauthorized)
				return
			}
			if jwtCtx.Claims.RangerHandle() == "" {
				slog.Error("No Ranger handle in JWT")
				http.Error(w, "Invalid Authorization token", http.StatusUnauthorized)
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
				http.Error(w, "This endpoint has been misconfigured. Please report this to the tech team",
					http.StatusInternalServerError)
				return
			}
			// TODO: this doesn't consider the ?event_id value, though maybe no endpoint needs it
			permissions, err := auth.UserPermissions2(r.Context(), r.PathValue("eventName"), imsDB, imsAdmins, *jwtCtx.Claims)
			if err != nil {
				slog.Error("Failed to compute permissions", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if !permissions[required] {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

//func Adapt(h http.HandlerFunc, adapters ...Adapter) http.Handler {
//	handler := http.Handler(h)
//	for i := range adapters {
//		adapter := adapters[len(adapters)-1-i] // range in reverse
//		handler = adapter(handler)
//	}
//	return handler
//}

func Adapt(handler http.Handler, adapters ...Adapter) http.Handler {
	for i := range adapters {
		adapter := adapters[len(adapters)-1-i] // range in reverse
		handler = adapter(handler)
	}
	return handler
}
