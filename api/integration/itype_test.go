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

func TestCreateIncident(t *testing.T) {
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

	typeA, typeB, typeC := "Cat", "Dog", "Emu"

	createTypes := imsjson.EditIncidentTypesRequest{
		Add:  imsjson.IncidentTypes{typeA, typeB, typeC},
		Hide: nil,
		Show: nil,
	}

	editTypes(t, serverURL, createTypes, jwt, httpClient)

	typesResp := getTypes(t, serverURL, jwt, httpClient)
	require.Contains(t, typesResp, typeA)
	require.Contains(t, typesResp, typeB)
	require.Contains(t, typesResp, typeC)
}

func editTypes(t *testing.T, serverURL *url.URL, req imsjson.EditIncidentTypesRequest, jwt string, httpClient *http.Client) {
	createTypesBytes, err := json.Marshal(req)
	require.NoError(t, err)
	httpPost, _ := http.NewRequest("POST", serverURL.JoinPath("/ims/api/incident_types").String(), strings.NewReader(string(createTypesBytes)))
	httpPost.Header.Set("Authorization", "Bearer "+jwt)
	resp, err := httpClient.Do(httpPost)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func getTypes(t *testing.T, serverURL *url.URL, jwt string, httpClient *http.Client) imsjson.IncidentTypes {
	httpReq, _ := http.NewRequest("GET", serverURL.JoinPath("/ims/api/incident_types").String(), nil)
	httpReq.Header.Set("Authorization", "Bearer "+jwt)
	get, err := httpClient.Do(httpReq)
	require.NoError(t, err)
	defer get.Body.Close()
	b, err := io.ReadAll(get.Body)
	require.NoError(t, err)
	var response imsjson.IncidentTypes
	err = json.Unmarshal(b, &response)
	require.NoError(t, err)
	return response
}
