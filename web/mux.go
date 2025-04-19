package web

import (
	"github.com/srabraham/ranger-ims-go/conf"
	"github.com/srabraham/ranger-ims-go/web/template"
	"net/http"
)

func AddToMux(mux *http.ServeMux, cfg *conf.IMSConfig) *http.ServeMux {
	mux.Handle("GET /ims/static/", http.StripPrefix("/ims/", http.FileServerFS(StaticFS)))

	mux.HandleFunc("GET /ims/app/{$}", func(w http.ResponseWriter, req *http.Request) {
		template.Root(cfg.Core.Deployment).Render(req.Context(), w)
	})

	mux.HandleFunc("GET /ims/app/events/{eventName}/incidents/{$}", func(w http.ResponseWriter, req *http.Request) {
		template.Incidents(cfg.Core.Deployment).Render(req.Context(), w)
	})

	mux.HandleFunc("GET /ims/auth/login/{$}", func(w http.ResponseWriter, req *http.Request) {
		template.Login(cfg.Core.Deployment).Render(req.Context(), w)
	})

	mux.HandleFunc("GET /ims/auth/logout/{$}", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "/ims/app?logout", http.StatusSeeOther)
	})

	return mux
}
