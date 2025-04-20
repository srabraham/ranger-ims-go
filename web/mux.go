package web

import (
	"github.com/srabraham/ranger-ims-go/conf"
	"github.com/srabraham/ranger-ims-go/web/template"
	"net/http"
)

func AddToMux(mux *http.ServeMux, cfg *conf.IMSConfig) *http.ServeMux {
	mux.Handle("GET /ims/static/",
		Adapt(
			http.StripPrefix("/ims/", http.FileServerFS(StaticFS)).ServeHTTP,
		),
	)
	mux.Handle("GET /ims/app/{$}",
		Adapt(
			func(w http.ResponseWriter, req *http.Request) {
				template.Root(cfg.Core.Deployment).Render(req.Context(), w)
			},
		),
	)
	mux.Handle("GET /ims/app/admin/{$}",
		Adapt(
			func(w http.ResponseWriter, req *http.Request) {
				template.AdminRoot(cfg.Core.Deployment).Render(req.Context(), w)
			},
		),
	)
	mux.Handle("GET /ims/app/admin/events/{$}",
		Adapt(
			func(w http.ResponseWriter, req *http.Request) {
				template.AdminEvents(cfg.Core.Deployment).Render(req.Context(), w)
			},
		),
	)
	mux.Handle("GET /ims/app/admin/streets/{$}",
		Adapt(
			func(w http.ResponseWriter, req *http.Request) {
				template.AdminStreets(cfg.Core.Deployment).Render(req.Context(), w)
			},
		),
	)
	mux.Handle("GET /ims/app/admin/types/{$}",
		Adapt(
			func(w http.ResponseWriter, req *http.Request) {
				template.AdminTypes(cfg.Core.Deployment).Render(req.Context(), w)
			},
		),
	)
	mux.Handle("GET /ims/app/events/{eventName}/field_reports/{$}",
		Adapt(
			func(w http.ResponseWriter, req *http.Request) {
				template.FieldReports(cfg.Core.Deployment).Render(req.Context(), w)
			},
		),
	)
	mux.Handle("GET /ims/app/events/{eventName}/field_reports/{fieldReportNumber}/{$}",
		Adapt(
			func(w http.ResponseWriter, req *http.Request) {
				template.FieldReport(cfg.Core.Deployment).Render(req.Context(), w)
			},
		),
	)
	mux.Handle("GET /ims/app/events/{eventName}/incidents/{$}",
		Adapt(
			func(w http.ResponseWriter, req *http.Request) {
				template.Incidents(cfg.Core.Deployment).Render(req.Context(), w)
			},
		),
	)
	mux.Handle("GET /ims/app/events/{eventName}/incidents/{incidentNumber}/{$}",
		Adapt(
			func(w http.ResponseWriter, req *http.Request) {
				template.Incident(cfg.Core.Deployment).Render(req.Context(), w)
			},
		),
	)
	mux.Handle("GET /ims/auth/login/{$}",
		Adapt(
			func(w http.ResponseWriter, req *http.Request) {
				template.Login(cfg.Core.Deployment).Render(req.Context(), w)
			},
		),
	)
	mux.Handle("GET /ims/auth/logout/{$}",
		Adapt(
			func(w http.ResponseWriter, req *http.Request) {
				http.Redirect(w, req, "/ims/app?logout", http.StatusSeeOther)
			},
		),
	)

	return mux
}

type Adapter func(http.Handler) http.Handler

func Adapt(h http.HandlerFunc, adapters ...Adapter) http.Handler {
	handler := http.Handler(h)
	for i := range adapters {
		adapter := adapters[len(adapters)-1-i] // range in reverse
		handler = adapter(h)
	}
	return handler
}
