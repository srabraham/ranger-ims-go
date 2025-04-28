package integration

import (
	"github.com/srabraham/ranger-ims-go/api"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"
)

func TestIncident(t *testing.T) {
	s := httptest.NewServer(api.AddToMux(nil, shared.cfg, shared.imsDB, nil))
	defer s.Close()
	serverURL, err := url.Parse(s.URL)
	require.NoError(t, err)

	apisAdmin := ApiHelper{t: t, serverURL: serverURL, jwt: jwtForTestAdminRanger(t)}
	apisNonAdmin := ApiHelper{t: t, serverURL: serverURL, jwt: jwtForRealTestUser(t)}

	// Use the admin JWT to create a new event,
	// then give the normal user Writer role on that event
	eventName := "IncidentEvent"
	resp := apisAdmin.editEvent(imsjson.EditEventsRequest{Add: []string{eventName}})
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp = apisAdmin.editAccess(imsjson.EventsAccess{
		eventName: imsjson.EventAccess{
			Writers: []imsjson.AccessRule{{
				Expression: "person:RealTestUserInConfig",
				Validity:   "always",
			}},
		},
	})
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Use normal user to create a new incident
	resp = apisNonAdmin.newIncident(imsjson.Incident{
		Event:         eventName,
		Summary:       ptr("my summary!"),
		RangerHandles: &[]string{"SomeOne", "SomeTwo"},
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	numStr := resp.Header.Get("X-IMS-Incident-Number")
	require.NotEmpty(t, numStr)
	num, err := strconv.ParseInt(numStr, 10, 32)
	require.NoError(t, err)
	require.GreaterOrEqual(t, num, int64(1))

	// Use normal user to get that incident back and check it over
	retrievedIncident, resp := apisNonAdmin.getIncident(eventName, int32(num))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, retrievedIncident)
	require.Equal(t, *retrievedIncident.Summary, "my summary!")
	require.Equal(t, retrievedIncident.State, "new")
	require.Equal(t, retrievedIncident.Priority, int8(3))
	require.NotNil(t, retrievedIncident.RangerHandles)
	require.Contains(t, *retrievedIncident.RangerHandles, "SomeOne")
	require.Contains(t, *retrievedIncident.RangerHandles, "SomeTwo")
	require.WithinDuration(t, time.Now(), retrievedIncident.Created, 5*time.Minute)
}

func ptr[T any](s T) *T {
	return &s
}
