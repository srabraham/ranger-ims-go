package api

import (
	"context"
	"database/sql"
	"github.com/srabraham/ranger-ims-go/auth"
	"github.com/srabraham/ranger-ims-go/conf"
	"github.com/srabraham/ranger-ims-go/directory"
	"github.com/srabraham/ranger-ims-go/store"
	"log/slog"
	"net/http"
	"time"
)

func AddToMux(mux *http.ServeMux, cfg *conf.IMSConfig, db *store.DB, userStore *directory.UserStore) *http.ServeMux {
	if mux == nil {
		mux = http.NewServeMux()
	}

	j := auth.JWTer{SecretKey: cfg.Core.JWTSecret}

	es := NewEventSourcerer()

	mux.Handle("GET /ims/api/access",
		Adapt(
			GetEventAccesses{imsDB: db},
			LogBeforeAfter(),
			ExtractJWTRequireAuthN(j),
			RequireGlobalPermission(auth.GlobalAdministrateEvents, db, cfg.Core.Admins),
		),
	)

	mux.Handle("POST /ims/api/access",
		Adapt(
			PostEventAccess{imsDB: db},
			LogBeforeAfter(),
			ExtractJWTRequireAuthN(j),
			RequireGlobalPermission(auth.GlobalAdministrateEvents, db, cfg.Core.Admins),
		),
	)

	mux.Handle("POST /ims/api/auth",
		Adapt(
			PostAuth{
				imsDB:       db,
				userStore:   userStore,
				jwtSecret:   cfg.Core.JWTSecret,
				jwtDuration: time.Duration(cfg.Core.TokenLifetime) * time.Second,
			},
			LogBeforeAfter(),
		),
		// This endpoint does not require authentication, nor
		// should it even consider the request's Authorization header.
	)

	mux.Handle("GET /ims/api/auth",
		Adapt(
			GetAuth{
				imsDB:     db,
				jwtSecret: cfg.Core.JWTSecret,
				admins:    cfg.Core.Admins,
			},
			LogBeforeAfter(),
			// This endpoint does not require authentication or authorization, by design
			ExtractJWTOptionalAuthN(j),
		),
	)

	mux.Handle("GET /ims/api/events/{eventName}/incidents",
		Adapt(
			GetIncidents{imsDB: db, imsAdmins: cfg.Core.Admins},
			LogBeforeAfter(),
			ExtractJWTRequireAuthN(j),
		),
	)

	mux.Handle("POST /ims/api/events/{eventName}/incidents",
		Adapt(
			NewIncident{imsDB: db, es: es, imsAdmins: cfg.Core.Admins},
			LogBeforeAfter(),
			ExtractJWTRequireAuthN(j),
		),
	)

	mux.Handle("GET /ims/api/events/{eventName}/incidents/{incidentNumber}",
		Adapt(
			GetIncident{imsDB: db, imsAdmins: cfg.Core.Admins},
			LogBeforeAfter(),
			ExtractJWTRequireAuthN(j),
		),
	)

	mux.Handle("POST /ims/api/events/{eventName}/incidents/{incidentNumber}",
		Adapt(
			EditIncident{imsDB: db, es: es, imsAdmins: cfg.Core.Admins},
			LogBeforeAfter(),
			ExtractJWTRequireAuthN(j),
		),
	)

	mux.Handle("POST /ims/api/events/{eventName}/incidents/{incidentNumber}/report_entries/{reportEntryId}",
		Adapt(
			EditIncidentReportEntry{imsDB: db, eventSource: es, imsAdmins: cfg.Core.Admins},
			LogBeforeAfter(),
			ExtractJWTRequireAuthN(j),
		),
	)

	mux.Handle("GET /ims/api/events/{eventName}/field_reports",
		Adapt(
			GetFieldReports{imsDB: db},
			LogBeforeAfter(),
			ExtractJWTRequireAuthN(j),
		),
	)

	mux.Handle("POST /ims/api/events/{eventName}/field_reports",
		Adapt(
			NewFieldReport{imsDB: db, eventSource: es},
			LogBeforeAfter(),
			ExtractJWTRequireAuthN(j),
		),
	)

	mux.Handle("GET /ims/api/events/{eventName}/field_reports/{fieldReportNumber}",
		Adapt(
			GetFieldReport{imsDB: db},
			LogBeforeAfter(),
			ExtractJWTRequireAuthN(j),
		),
	)

	mux.Handle("POST /ims/api/events/{eventName}/field_reports/{fieldReportNumber}",
		Adapt(
			EditFieldReport{imsDB: db, eventSource: es},
			LogBeforeAfter(),
			ExtractJWTRequireAuthN(j),
		),
	)

	mux.Handle("POST /ims/api/events/{eventName}/field_reports/{fieldReportNumber}/report_entries/{reportEntryId}",
		Adapt(
			EditFieldReportReportEntry{imsDB: db, eventSource: es, imsAdmins: cfg.Core.Admins},
			LogBeforeAfter(),
			ExtractJWTRequireAuthN(j),
		),
	)

	mux.Handle("GET /ims/api/events",
		Adapt(
			GetEvents{imsDB: db, imsAdmins: cfg.Core.Admins},
			LogBeforeAfter(),
			ExtractJWTRequireAuthN(j),
			// ugh, no eventname in path
			//RequireEventPermission(auth.EventReadEventName, db, cfg.Core.Admins),
		),
	)

	mux.Handle("POST /ims/api/events",
		Adapt(
			EditEvents{imsDB: db, imsAdmins: cfg.Core.Admins},
			LogBeforeAfter(),
			ExtractJWTRequireAuthN(j),
			RequireGlobalPermission(auth.GlobalAdministrateEvents, db, cfg.Core.Admins),
		),
	)

	mux.Handle("GET /ims/api/streets",
		Adapt(
			GetStreets{imsDB: db},
			LogBeforeAfter(),
			ExtractJWTRequireAuthN(j),
			RequireGlobalPermission(auth.GlobalReadStreets, db, cfg.Core.Admins),
		),
	)

	mux.Handle("POST /ims/api/streets",
		Adapt(
			EditStreets{imsDB: db},
			LogBeforeAfter(),
			ExtractJWTRequireAuthN(j),
			RequireGlobalPermission(auth.GlobalAdministrateStreets, db, cfg.Core.Admins),
		),
	)

	mux.Handle("GET /ims/api/incident_types",
		Adapt(
			GetIncidentTypes{imsDB: db},
			LogBeforeAfter(),
			ExtractJWTRequireAuthN(j),
			RequireGlobalPermission(auth.GlobalReadIncidentTypes, db, cfg.Core.Admins),
		),
	)

	mux.Handle("POST /ims/api/incident_types",
		Adapt(
			EditIncidentTypes{imsDB: db},
			LogBeforeAfter(),
			ExtractJWTRequireAuthN(j),
			RequireGlobalPermission(auth.GlobalAdministrateIncidentTypes, db, cfg.Core.Admins),
		),
	)

	mux.Handle("GET /ims/api/personnel",
		Adapt(
			GetPersonnel{userStore: userStore},
			LogBeforeAfter(),
			ExtractJWTRequireAuthN(j),
			RequireGlobalPermission(auth.GlobalReadPersonnel, db, cfg.Core.Admins),
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
			slog.Debug("Done serving request", "duration", time.Since(start).Round(100*time.Microsecond), "method", r.Method, "path", r.URL.Path)
		})
	}
}

type ContextKey string

const JWTContextKey ContextKey = "JWTContext"

type JWTContext struct {
	Claims *auth.IMSClaims
	Error  error
}

const PermissionsContextKey ContextKey = "PermissionsContext"

type PermissionsContext struct {
	EventPermissions  map[int32]auth.EventPermissionMask
	GlobalPermissions auth.GlobalPermissionMask
}

func ExtractJWTOptionalAuthN(j auth.JWTer) Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			claims, err := j.AuthenticateJWT(header)
			ctx := context.WithValue(r.Context(), JWTContextKey, JWTContext{
				Claims: claims,
				Error:  err,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ExtractJWTRequireAuthN(j auth.JWTer) Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			claims, err := j.AuthenticateJWT(header)
			if err != nil || claims == nil {
				slog.Error("JWT error", "error", err)
				http.Error(w, "Invalid Authorization token", http.StatusUnauthorized)
				return
			}
			if claims.RangerHandle() == "" {
				slog.Error("No Ranger handle in JWT")
				http.Error(w, "Invalid Authorization token", http.StatusUnauthorized)
				return
			}
			jwtCtx := context.WithValue(r.Context(), JWTContextKey, JWTContext{
				Claims: claims,
				Error:  err,
			})
			next.ServeHTTP(w, r.WithContext(jwtCtx))
		})
	}
}

func ExtractPermissionsToContext(imsDB *store.DB, imsAdmins []string) Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			jwtCtx, ok := mustGetJwtCtx(w, r)
			if !ok {
				return
			}
			var eventID *int32
			if r.PathValue("eventName") != "" {
				event, ok := mustGetEvent(w, r, r.PathValue("eventName"), imsDB)
				if !ok {
					return
				}
				eventID = &event.ID
			}

			// TODO: this doesn't consider the ?event_id value, though maybe no endpoint needs it
			eventPermissions, globalPermissions, err := auth.UserPermissions2(r.Context(), eventID, imsDB, imsAdmins, *jwtCtx.Claims)
			if err != nil {
				slog.Error("Failed to compute permissions", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), PermissionsContextKey, PermissionsContext{
				EventPermissions:  eventPermissions,
				GlobalPermissions: globalPermissions,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireEventPermission(required auth.EventPermissionMask, imsDB *store.DB, imsAdmins []string) Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			jwtCtx, ok := mustGetJwtCtx(w, r)
			if !ok {
				return
			}
			var eventID *int32
			if r.PathValue("eventName") != "" {
				event, ok := mustGetEvent(w, r, r.PathValue("eventName"), imsDB)
				if !ok {
					return
				}
				eventID = &event.ID
			}

			// TODO: this doesn't consider the ?event_id value, though maybe no endpoint needs it
			eventPermissions, _, err := auth.UserPermissions2(r.Context(), eventID, imsDB, imsAdmins, *jwtCtx.Claims)
			if err != nil {
				slog.Error("Failed to compute permissions", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if eventID == nil || eventPermissions[*eventID]&required == 0 {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireGlobalPermission(required auth.GlobalPermissionMask, imsDB *store.DB, imsAdmins []string) Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			jwtCtx, ok := mustGetJwtCtx(w, r)
			if !ok {
				return
			}

			_, globalPermissions, err := auth.UserPermissions2(r.Context(), nil, imsDB, imsAdmins, *jwtCtx.Claims)
			if err != nil {
				slog.Error("Failed to compute permissions", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if globalPermissions&required == 0 {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
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
