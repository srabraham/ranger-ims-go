package integration

import (
	"github.com/srabraham/ranger-ims-go/api"
	"github.com/srabraham/ranger-ims-go/auth"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestGetEvent(t *testing.T) {
	s := httptest.NewServer(api.AddToMux(nil, imsCfg, imsDB, nil))
	defer s.Close()
	serverURL, err := url.Parse(s.URL)
	require.NoError(t, err)

	jwt := auth.JWTer{SecretKey: imsCfg.Core.JWTSecret}.CreateJWT(
		imsCfg.Core.Admins[0], 123, nil, nil, true, 1*time.Hour,
	)
	apis := ApiHelper{t: t, serverURL: serverURL, jwt: jwt}

	jwtForNonAdminUser := auth.JWTer{SecretKey: imsCfg.Core.JWTSecret}.CreateJWT(
		"Not an admin!", 1234, nil, nil, true, 1*time.Hour,
	)
	apisNonAdminUser := ApiHelper{t: t, serverURL: serverURL, jwt: jwtForNonAdminUser}

	testEventName := "MyNewEvent"

	editEventReq := imsjson.EditEventsRequest{
		Add: []string{testEventName},
	}
	resp := apis.editEvent(editEventReq)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	resp = apisNonAdminUser.editEvent(editEventReq)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)

	// editEventBytes, _ := json.Marshal(editEvent)
	// httpPost, _ := http.NewRequest("POST", serverURL.JoinPath("/ims/api/events").String(), strings.NewReader(string(editEventBytes)))
	// httpPost.Header.Set("Authorization", "Bearer "+jwt)
	// resp, err := httpClient.Do(httpPost)
	// require.NoError(t, err)
	// require.Equal(t, http.StatusNoContent, resp.StatusCode)

	accessReq := imsjson.EventsAccess{
		testEventName: {
			Writers: []imsjson.AccessRule{
				{
					Expression: "person:" + imsCfg.Core.Admins[0],
					Validity:   "always",
				},
			},
		},
	}
	resp = apis.editAccess(accessReq)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	// access := getAccess(t, serverURL, jwt)

	events, resp := apis.getEvents()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, imsjson.Events{
		{
			ID:   1,
			Name: "MyNewEvent",
		},
	}, events)
}
