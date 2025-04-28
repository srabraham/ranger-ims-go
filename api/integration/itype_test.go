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

func TestIncidentTypesAPIAuthorization(t *testing.T) {
	s := httptest.NewServer(api.AddToMux(nil, shared.cfg, shared.imsDB, nil))
	defer s.Close()
	serverURL, err := url.Parse(s.URL)
	require.NoError(t, err)

	apisAdmin := ApiHelper{t: t, serverURL: serverURL, jwt: shared.jwtAdmin}
	apisNonAdmin := ApiHelper{t: t, serverURL: serverURL, jwt: shared.jwtNormalUser}
	apisNotAuthenticated := ApiHelper{t: t, serverURL: serverURL, jwt: ""}

	// Any authenticated user can call GetIncidentTypes
	_, resp := apisNotAuthenticated.getTypes(false)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	_, resp = apisNonAdmin.getTypes(false)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_, resp = apisAdmin.getTypes(false)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Only admins can hit the EditIncidentTypes endpoint
	// An unauthenticated client will get a 401
	// An unauthorized user will get a 403
	editTypesReq := imsjson.EditIncidentTypesRequest{}
	resp = apisNotAuthenticated.editTypes(editTypesReq)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp = apisNonAdmin.editTypes(editTypesReq)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp = apisAdmin.editTypes(editTypesReq)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestCreateIncident(t *testing.T) {
	s := httptest.NewServer(api.AddToMux(nil, shared.cfg, shared.imsDB, nil))
	defer s.Close()
	serverURL, err := url.Parse(s.URL)
	require.NoError(t, err)

	jwt := auth.JWTer{SecretKey: shared.cfg.Core.JWTSecret}.CreateJWT(
		shared.cfg.Core.Admins[0], 123, nil, nil, true, 1*time.Hour,
	)
	apis := ApiHelper{t: t, serverURL: serverURL, jwt: jwt}

	// Make three new incident types
	typeA, typeB, typeC := "Cat", "Dog", "Emu"
	createTypes := imsjson.EditIncidentTypesRequest{
		Add:  imsjson.IncidentTypes{typeA, typeB, typeC},
		Hide: nil,
		Show: nil,
	}
	apis.editTypes(createTypes)

	// All three types should now be retrievable and non-hidden
	typesResp, resp := apis.getTypes(false)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, typesResp, typeA)
	require.Contains(t, typesResp, typeB)
	require.Contains(t, typesResp, typeC)

	// Hide one of those types
	hideOne := imsjson.EditIncidentTypesRequest{
		Hide: imsjson.IncidentTypes{typeA},
	}
	apis.editTypes(hideOne)

	// That type should no longer appear from the standard incident type query
	typesVisibleOnly, resp := apis.getTypes(false)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotContains(t, typesVisibleOnly, typeA)
	require.Contains(t, typesVisibleOnly, typeB)
	require.Contains(t, typesVisibleOnly, typeC)
	// but it will still appears when includeHidden=true
	typesIncludeHidden, resp := apis.getTypes(true)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, typesIncludeHidden, typeA)
	require.Contains(t, typesIncludeHidden, typeB)
	require.Contains(t, typesIncludeHidden, typeC)

	// Unhide that type we previously hid
	showItAgain := imsjson.EditIncidentTypesRequest{
		Show: imsjson.IncidentTypes{typeA, typeB},
	}
	apis.editTypes(showItAgain)
	// and see that it's back in the standard incident type query results
	typesVisibleOnly, resp = apis.getTypes(false)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, typesVisibleOnly, typeA)
	require.Contains(t, typesVisibleOnly, typeB)
	require.Contains(t, typesVisibleOnly, typeC)
}
