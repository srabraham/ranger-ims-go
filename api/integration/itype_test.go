package integration

import (
	"bytes"
	"encoding/json"
	"github.com/srabraham/ranger-ims-go/api"
	"github.com/srabraham/ranger-ims-go/auth"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestCreateIncident(t *testing.T) {
	s := httptest.NewServer(api.AddToMux(nil, imsCfg, imsDB, nil))
	defer s.Close()
	serverURL, err := url.Parse(s.URL)
	require.NoError(t, err)

	jwt := auth.JWTer{SecretKey: imsCfg.Core.JWTSecret}.CreateJWT(
		imsCfg.Core.Admins[0], 123, nil, nil, true, 1*time.Hour,
	)

	// Make three new incident types
	typeA, typeB, typeC := "Cat", "Dog", "Emu"
	createTypes := imsjson.EditIncidentTypesRequest{
		Add:  imsjson.IncidentTypes{typeA, typeB, typeC},
		Hide: nil,
		Show: nil,
	}
	editTypes(t, serverURL, createTypes, jwt)

	// All three types should now be retrievable and non-hidden
	typesResp := getTypes(t, serverURL, jwt, false)
	require.Contains(t, typesResp, typeA)
	require.Contains(t, typesResp, typeB)
	require.Contains(t, typesResp, typeC)

	// Hide one of those types
	hideOne := imsjson.EditIncidentTypesRequest{
		Hide: imsjson.IncidentTypes{typeA},
	}
	editTypes(t, serverURL, hideOne, jwt)

	// That type should no longer appear from the standard incident type query
	typesVisibleOnly := getTypes(t, serverURL, jwt, false)
	require.NotContains(t, typesVisibleOnly, typeA)
	require.Contains(t, typesVisibleOnly, typeB)
	require.Contains(t, typesVisibleOnly, typeC)
	// but it will still appears when includeHidden=true
	typesIncludeHidden := getTypes(t, serverURL, jwt, true)
	require.Contains(t, typesIncludeHidden, typeA)
	require.Contains(t, typesIncludeHidden, typeB)
	require.Contains(t, typesIncludeHidden, typeC)

	// Unhide that type we previously hid
	showItAgain := imsjson.EditIncidentTypesRequest{
		Show: imsjson.IncidentTypes{typeA, typeB},
	}
	editTypes(t, serverURL, showItAgain, jwt)
	// and see that it's back in the standard incident type query results
	typesVisibleOnly = getTypes(t, serverURL, jwt, false)
	require.Contains(t, typesVisibleOnly, typeA)
	require.Contains(t, typesVisibleOnly, typeB)
	require.Contains(t, typesVisibleOnly, typeC)
}

func editTypes(t *testing.T, serverURL *url.URL, req imsjson.EditIncidentTypesRequest, jwt string) {
	imsPost(t, req, serverURL.JoinPath("/ims/api/incident_types").String(), jwt)
}

func getTypes(t *testing.T, serverURL *url.URL, jwt string, includeHidden bool) imsjson.IncidentTypes {
	path := serverURL.JoinPath("/ims/api/incident_types").String()
	if includeHidden {
		path = path + "?hidden=true"
	}
	return imsGetAs[imsjson.IncidentTypes](t, path, jwt)
}

func imsPost[T any](t *testing.T, body T, path, jwt string) {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	postBody, err := json.Marshal(body)
	require.NoError(t, err)
	httpPost, err := http.NewRequest("POST", path, bytes.NewReader(postBody))
	require.NoError(t, err)
	httpPost.Header.Set("Authorization", "Bearer "+jwt)
	resp, err := httpClient.Do(httpPost)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func imsGetAs[T any](t *testing.T, path, jwt string) T {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	httpReq, _ := http.NewRequest("GET", path, nil)
	httpReq.Header.Set("Authorization", "Bearer "+jwt)
	get, err := httpClient.Do(httpReq)
	require.NoError(t, err)
	defer get.Body.Close()
	b, err := io.ReadAll(get.Body)
	require.NoError(t, err)
	var response T
	err = json.Unmarshal(b, &response)
	require.NoError(t, err)
	return response
}
