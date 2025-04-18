package web

import (
	"net/http"
)

func AddToMux(mux *http.ServeMux) *http.ServeMux {
	mux.Handle("GET /ims/app/static/", http.StripPrefix("/ims/app/", http.FileServerFS(StaticFS)))
	return mux
}
