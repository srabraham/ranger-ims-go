package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/srabraham/ranger-ims-go/auth"
	"github.com/srabraham/ranger-ims-go/conf"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"github.com/srabraham/ranger-ims-go/store"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic")
			shutdown(ctx)
		}
	}()
	setup(ctx)
	code := m.Run()
	shutdown(ctx)
	os.Exit(code)
}

var (
	imsDBContainer testcontainers.Container
	imsCfg         *conf.IMSConfig
	imsDB          *sql.DB
)

func setup(ctx context.Context) {
	imsCfg = conf.DefaultIMS()
	imsCfg.Core.JWTSecret = uuid.New().String()
	imsCfg.Core.Admins = []string{"AdminRanger"}
	imsCfg.Store.MySQL.Database = "ims"
	imsCfg.Store.MySQL.Username = "rangers"
	imsCfg.Store.MySQL.Password = uuid.New().String()
	req := testcontainers.ContainerRequest{
		Image:        "mariadb:10.5.27",
		ExposedPorts: []string{"3306/tcp"},
		WaitingFor:   wait.ForListeningPort("3306/tcp"),
		Env: map[string]string{
			"MARIADB_RANDOM_ROOT_PASSWORD": "true",
			"MARIADB_DATABASE":             imsCfg.Store.MySQL.Database,
			"MARIADB_USER":                 imsCfg.Store.MySQL.Username,
			"MARIADB_PASSWORD":             imsCfg.Store.MySQL.Password,
		},
	}
	var err error
	imsDBContainer, err = testcontainers.GenericContainer(ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
			//Logger:           log.New(os.Stdout, "MariaDB ", log.LstdFlags),
		},
	)
	if err != nil {
		panic(err)
	}
	endpoint, err := imsDBContainer.Endpoint(ctx, "")
	if err != nil {
		panic(err)
	}
	port, _ := strconv.Atoi(strings.TrimPrefix(endpoint, "localhost:"))
	imsCfg.Store.MySQL.HostPort = int32(port)
	imsDB = store.MariaDB(imsCfg)
}

func shutdown(ctx context.Context) {
	_ = imsDB.Close()
	err := imsDBContainer.Terminate(ctx)
	if err != nil {
		// log and continue
		slog.Error("Failed to terminate container", "error", err)
	}
}

func TestGetEvent(t *testing.T) {
	ctx := t.Context()

	script := "BEGIN NOT ATOMIC\n" + store.CurrentSchema + "\nEND"
	_, err := imsDB.ExecContext(ctx, script)
	require.NoError(t, err)

	s := httptest.NewServer(AddToMux(nil, imsCfg, imsDB, nil))
	defer s.Close()
	serverURL, _ := url.Parse(s.URL)

	jwt := auth.JWTer{SecretKey: imsCfg.Core.JWTSecret}.CreateJWT(
		imsCfg.Core.Admins[0], 123, nil, nil, true, 1*time.Hour,
	)

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	editEvent := imsjson.EditEventsRequest{
		Add: []string{"MyNewEvent"},
	}
	editEventBytes, _ := json.Marshal(editEvent)
	httpPost, _ := http.NewRequest("POST", serverURL.JoinPath("/ims/api/events").String(), strings.NewReader(string(editEventBytes)))
	httpPost.Header.Set("Authorization", "Bearer "+jwt)
	resp, err := httpClient.Do(httpPost)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	accessReq := imsjson.EventsAccess{
		"MyNewEvent": {
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
	log.Printf("have events %v", string(b))
}
