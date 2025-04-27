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

func (a ApiHelper) editTypes(req imsjson.EditIncidentTypesRequest) {
	a.imsPost(req, a.serverURL.JoinPath("/ims/api/incident_types").String())
}

func (a ApiHelper) getTypes(includeHidden bool) imsjson.IncidentTypes {
	path := a.serverURL.JoinPath("/ims/api/incident_types").String()
	if includeHidden {
		path = path + "?hidden=true"
	}
	return *a.imsGet(path, &imsjson.IncidentTypes{}).(*imsjson.IncidentTypes)
}

func (a ApiHelper) editEvent(req imsjson.EditEventsRequest) {
	a.imsPost(req, a.serverURL.JoinPath("/ims/api/events").String())
}

func (a ApiHelper) getEvents() imsjson.Events {
	return *a.imsGet(a.serverURL.JoinPath("/ims/api/events").String(), &imsjson.Events{}).(*imsjson.Events)
}

func (a ApiHelper) editAccess(req imsjson.EventsAccess) {
	a.imsPost(req, a.serverURL.JoinPath("/ims/api/access").String())
}

func (a ApiHelper) getAccess() imsjson.EventsAccess {
	return *a.imsGet(a.serverURL.JoinPath("/ims/api/access").String(), &imsjson.EventsAccess{}).(*imsjson.EventsAccess)
}

func (a ApiHelper) imsPost(body any, path string) {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	postBody, err := json.Marshal(body)
	require.NoError(a.t, err)
	httpPost, err := http.NewRequest("POST", path, bytes.NewReader(postBody))
	require.NoError(a.t, err)
	httpPost.Header.Set("Authorization", "Bearer "+a.jwt)
	resp, err := httpClient.Do(httpPost)
	require.NoError(a.t, err)
	require.Equal(a.t, http.StatusNoContent, resp.StatusCode)
}

func (a ApiHelper) imsGet(path string, resp any) any {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	httpReq, _ := http.NewRequest("GET", path, nil)
	httpReq.Header.Set("Authorization", "Bearer "+a.jwt)
	get, err := httpClient.Do(httpReq)
	require.NoError(a.t, err)
	defer get.Body.Close()
	b, err := io.ReadAll(get.Body)
	require.NoError(a.t, err)
	err = json.Unmarshal(b, &resp)
	require.NoError(a.t, err)
	return resp
}
