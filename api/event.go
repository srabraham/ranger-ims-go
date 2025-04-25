package api

import (
	"cmp"
	"github.com/srabraham/ranger-ims-go/auth"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"github.com/srabraham/ranger-ims-go/store"
	"github.com/srabraham/ranger-ims-go/store/imsdb"
	"log/slog"
	"net/http"
	"regexp"
	"slices"
)

type GetEvents struct {
	imsDB     *store.DB
	imsAdmins []string
}

func (action GetEvents) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	resp := make(imsjson.Events, 0)
	ctx := req.Context()

	eventRows, err := imsdb.New(action.imsDB).Events(ctx)
	if err != nil {
		slog.Error("Failed to get events", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	accessRows, err := imsdb.New(action.imsDB).EventAccessAll(ctx)
	if err != nil {
		panic(err)
	}
	accessRowByEventID := make(map[int32][]imsdb.EventAccess)
	for _, ar := range accessRows {
		accessRowByEventID[ar.EventAccess.Event] = append(accessRowByEventID[ar.EventAccess.Event], ar.EventAccess)
	}

	claims := ctx.Value(JWTContextKey).(JWTContext).Claims

	permissionsByEvent, _ := auth.UserPermissions3(
		accessRowByEventID,
		action.imsAdmins,
		claims.RangerHandle(),
		claims.RangerOnSite(),
		claims.RangerPositions(),
		claims.RangerTeams(),
	)

	for _, er := range eventRows {
		if permissionsByEvent[er.Event.ID]&auth.EventReadEventName != 0 {
			resp = append(resp, imsjson.Event{
				ID:   er.Event.ID,
				Name: er.Event.Name,
			})
		}
	}
	//
	//for _, er := range eventRows {
	//	perms := auth.UserPermissions(
	//		accessRowByEventID[er.Event.ID],
	//		action.imsAdmins,
	//		claims.RangerHandle(),
	//		claims.RangerOnSite(),
	//		claims.RangerPositions(),
	//		claims.RangerTeams(),
	//	)
	//	if perms[auth.EventReadEventName] {
	//		resp = append(resp, imsjson.Event{
	//			ID:   er.Event.ID,
	//			Name: er.Event.Name,
	//		})
	//	}
	//}

	slices.SortFunc(resp, func(a, b imsjson.Event) int {
		return cmp.Compare(a.ID, b.ID)
	})

	w.Header().Set("Cache-Control", "max-age=1200, private")
	mustWriteJSON(w, resp)
}

type EditEvents struct {
	imsDB     *store.DB
	imsAdmins []string
}

// Require basic cleanliness for EventName, since it's used in IMS URLs
// and in filesystem directory paths.
var allowedEventNames = regexp.MustCompile(`^[\w-]+$`)

func (action EditEvents) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	if ok := mustParseForm(w, req); !ok {
		return
	}

	editRequest, ok := mustReadBodyAs[imsjson.EditEventsRequest](w, req)
	if !ok {
		return
	}

	for _, eventName := range editRequest.Add {
		if !allowedEventNames.MatchString(eventName) {
			slog.Error("Invalid event name", "eventName", eventName)
			http.Error(w, "Event names must match the pattern "+allowedEventNames.String(), http.StatusBadRequest)
			return
		}
		id, err := imsdb.New(action.imsDB).CreateEvent(ctx, eventName)
		if err != nil {
			slog.Error("Failed to create event", "eventName", eventName, "error", err)
			http.Error(w, "Failed to create event", http.StatusInternalServerError)
			return
		}
		slog.Info("Created event", "eventName", eventName, "id", id)
	}
	http.Error(w, "Success", http.StatusNoContent)
}
