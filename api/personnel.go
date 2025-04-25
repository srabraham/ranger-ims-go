package api

import (
	"github.com/srabraham/ranger-ims-go/directory"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"log/slog"
	"net/http"
)

type GetPersonnel struct {
	//clubhouseDB *directory.DB
	userStore *directory.UserStore
}

type GetPersonnelResponse []imsjson.Person

func (action GetPersonnel) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	response := make(GetPersonnelResponse, 0)

	rangers, err := action.userStore.GetRangers(req.Context())
	if err != nil {
		slog.Error("Failed to get personnel", "error", err)
		http.Error(w, "Failed to get personnel", http.StatusInternalServerError)
		return
	}

	for _, ranger := range rangers {
		response = append(response, imsjson.Person{
			Handle: ranger.Handle,
			// Don't send out email addresses in the API
			Email: "",
			// Don't send out passwords in the API
			Password:    "",
			Status:      ranger.Status,
			Onsite:      ranger.Onsite,
			DirectoryID: ranger.DirectoryID,
		})
	}

	w.Header().Set("Cache-Control", "max-age=1200, private")
	mustWriteJSON(w, response)
}
