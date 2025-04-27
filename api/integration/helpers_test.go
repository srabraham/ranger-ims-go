package integration

import (
	"bytes"
	"encoding/json"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"
)

type ApiHelper struct {
	t         *testing.T
	serverURL *url.URL
	jwt       string
}

func (a ApiHelper) editTypes(req imsjson.EditIncidentTypesRequest) *http.Response {
	return a.imsPost(req, a.serverURL.JoinPath("/ims/api/incident_types").String())
}

func (a ApiHelper) getTypes(includeHidden bool) (imsjson.IncidentTypes, *http.Response) {
	path := a.serverURL.JoinPath("/ims/api/incident_types").String()
	if includeHidden {
		path = path + "?hidden=true"
	}
	bod, resp := a.imsGet(path, &imsjson.IncidentTypes{})
	return *bod.(*imsjson.IncidentTypes), resp
}

func (a ApiHelper) editEvent(req imsjson.EditEventsRequest) *http.Response {
	return a.imsPost(req, a.serverURL.JoinPath("/ims/api/events").String())
}

func (a ApiHelper) getEvents() (imsjson.Events, *http.Response) {
	bod, resp := a.imsGet(a.serverURL.JoinPath("/ims/api/events").String(), &imsjson.Events{})
	return *bod.(*imsjson.Events), resp
}

func (a ApiHelper) editAccess(req imsjson.EventsAccess) *http.Response {
	return a.imsPost(req, a.serverURL.JoinPath("/ims/api/access").String())
}

func (a ApiHelper) getAccess() (imsjson.EventsAccess, *http.Response) {
	bod, resp := a.imsGet(a.serverURL.JoinPath("/ims/api/access").String(), &imsjson.EventsAccess{})
	return *bod.(*imsjson.EventsAccess), resp
	// return *a.imsGet(a.serverURL.JoinPath("/ims/api/access").String(), &imsjson.EventsAccess{}).(*imsjson.EventsAccess)
}

func (a ApiHelper) imsPost(body any, path string) *http.Response {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	postBody, err := json.Marshal(body)
	require.NoError(a.t, err)
	httpPost, err := http.NewRequest("POST", path, bytes.NewReader(postBody))
	require.NoError(a.t, err)
	if a.jwt != "" {
		httpPost.Header.Set("Authorization", "Bearer "+a.jwt)
	}
	resp, err := httpClient.Do(httpPost)
	require.NoError(a.t, err)
	return resp
	// require.Equal(a.t, http.StatusNoContent, resp.StatusCode)
}

func (a ApiHelper) imsGet(path string, resp any) (any, *http.Response) {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	httpReq, err := http.NewRequest("GET", path, nil)
	require.NoError(a.t, err)
	if a.jwt != "" {
		httpReq.Header.Set("Authorization", "Bearer "+a.jwt)
	}
	get, err := httpClient.Do(httpReq)
	require.NoError(a.t, err)
	defer get.Body.Close()
	b, err := io.ReadAll(get.Body)
	require.NoError(a.t, err)
	if get.StatusCode != http.StatusOK {
		return resp, get
	}
	err = json.Unmarshal(b, &resp)
	require.NoError(a.t, err)
	return resp, get
}
