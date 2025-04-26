package integration

import (
	"encoding/json"
	"github.com/srabraham/ranger-ims-go/api"
	"github.com/srabraham/ranger-ims-go/auth"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
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

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	testEventName := "MyNewEvent"

	editEvent := imsjson.EditEventsRequest{
		Add: []string{testEventName},
	}
	editEventBytes, _ := json.Marshal(editEvent)
	httpPost, _ := http.NewRequest("POST", serverURL.JoinPath("/ims/api/events").String(), strings.NewReader(string(editEventBytes)))
	httpPost.Header.Set("Authorization", "Bearer "+jwt)
	resp, err := httpClient.Do(httpPost)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

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
	accessReqBytes, _ := json.Marshal(accessReq)
	httpPost, _ = http.NewRequest("POST", serverURL.JoinPath("/ims/api/access").String(), strings.NewReader(string(accessReqBytes)))
	httpPost.Header.Set("Authorization", "Bearer "+jwt)
	resp, err = httpClient.Do(httpPost)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	httpReq, _ := http.NewRequest("GET", serverURL.JoinPath("/ims/api/events").String(), nil)
	httpReq.Header.Set("Authorization", "Bearer "+jwt)
	get, err := httpClient.Do(httpReq)
	require.NoError(t, err)
	defer get.Body.Close()
	b, err := io.ReadAll(get.Body)
	require.NoError(t, err)

	var events imsjson.Events
	err = json.Unmarshal(b, &events)
	require.NoError(t, err)

	require.Equal(t, imsjson.Events{
		{
			ID:   1,
			Name: "MyNewEvent",
		},
	}, events)
}
