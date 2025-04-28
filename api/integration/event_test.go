package integration

import (
	"github.com/srabraham/ranger-ims-go/api"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestEventAPIAuthorization(t *testing.T) {
	s := httptest.NewServer(api.AddToMux(nil, shared.cfg, shared.imsDB, nil))
	defer s.Close()
	serverURL, err := url.Parse(s.URL)
	require.NoError(t, err)

	apisAdmin := ApiHelper{t: t, serverURL: serverURL, jwt: jwtForTestAdminRanger(t)}
	apisNonAdmin := ApiHelper{t: t, serverURL: serverURL, jwt: jwtForRealTestUser(t)}
	apisNotAuthenticated := ApiHelper{t: t, serverURL: serverURL, jwt: ""}

	// Any authenticated user can call GetEvents
	_, resp := apisNotAuthenticated.getEvents()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	_, resp = apisNonAdmin.getEvents()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_, resp = apisAdmin.getEvents()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Only admins can hit the EditEvents endpoint
	// An unauthenticated client will get a 401
	// An unauthorized user will get a 403
	editEventReq := imsjson.EditEventsRequest{}
	resp = apisNotAuthenticated.editEvent(editEventReq)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp = apisNonAdmin.editEvent(editEventReq)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp = apisAdmin.editEvent(editEventReq)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestGetAndEditEvent(t *testing.T) {
	s := httptest.NewServer(api.AddToMux(nil, shared.cfg, shared.imsDB, nil))
	defer s.Close()
	serverURL, err := url.Parse(s.URL)
	require.NoError(t, err)

	apisAdmin := ApiHelper{t: t, serverURL: serverURL, jwt: jwtForTestAdminRanger(t)}

	testEventName := "MyNewEvent"

	editEventReq := imsjson.EditEventsRequest{
		Add: []string{testEventName},
	}

	resp := apisAdmin.editEvent(editEventReq)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	accessReq := imsjson.EventsAccess{
		testEventName: {
			Writers: []imsjson.AccessRule{
				{
					Expression: "person:TestAdminRanger",
					Validity:   "always",
				},
			},
		},
	}
	resp = apisAdmin.editAccess(accessReq)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	events, resp := apisAdmin.getEvents()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, imsjson.Events{
		{
			ID:   1,
			Name: "MyNewEvent",
		},
	}, events)
}
