package integration

import (
	"github.com/srabraham/ranger-ims-go/api"
	"github.com/srabraham/ranger-ims-go/directory"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestPostAuthAPIAuthorization(t *testing.T) {
	userStore, err := directory.NewUserStore(shared.cfg.Directory.TestUsers, nil)
	require.NoError(t, err)
	s := httptest.NewServer(api.AddToMux(nil, shared.cfg, shared.imsDB, userStore))
	defer s.Close()
	serverURL, err := url.Parse(s.URL)
	require.NoError(t, err)

	apisNotAuthenticated := ApiHelper{t: t, serverURL: serverURL, jwt: ""}

	// A user who doesn't exist gets s 401
	statusCode, body, token := apisNotAuthenticated.postAuth(api.PostAuthRequest{
		Identification: "Not a real user",
		Password:       "password123",
	})
	require.Equal(t, http.StatusUnauthorized, statusCode)
	require.Contains(t, body, "bad credentials")
	require.Empty(t, token)

	// A user with the correct password gets logged in and gets a JWT
	statusCode, body, token = apisNotAuthenticated.postAuth(api.PostAuthRequest{
		Identification: "realtestuser@rangers.brc",
		Password:       "password",
	})
	require.Equal(t, http.StatusOK, statusCode)
	require.NotEmpty(t, token)

	// That same valid user can also log in by handle
	statusCode, body, token = apisNotAuthenticated.postAuth(api.PostAuthRequest{
		Identification: "RealTestUserInConfig",
		Password:       "password",
	})
	require.Equal(t, http.StatusOK, statusCode)
	require.NotEmpty(t, token)

	// A valid user with the wrong password gets denied entry
	statusCode, body, token = apisNotAuthenticated.postAuth(api.PostAuthRequest{
		Identification: "RealTestUserInConfig",
		Password:       "not my password",
	})
	require.Equal(t, http.StatusUnauthorized, statusCode)
	require.Contains(t, body, "bad credentials")
	require.Empty(t, token)
}
