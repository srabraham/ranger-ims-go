package integration

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/srabraham/ranger-ims-go/conf"
	"github.com/srabraham/ranger-ims-go/directory"
	"github.com/srabraham/ranger-ims-go/store"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"testing"
)

// mainTestInternal contains fields to be used only within main_test.go.
var mainTestInternal struct {
	imsDBContainer testcontainers.Container
}

// shared contains fields that may be used by any test in the integration package.
// These are fields from the common setup performed in main_test.go.
var shared struct {
	cfg   *conf.IMSConfig
	imsDB *store.DB
	// jwtAdmin, jwtNormalUser string
	userStore *directory.UserStore
}

// TestMain does the common setup and teardown for all tests in this package.
// It's slow to start up a MariaDB container, so we want to only have to do
// that once for the whole suite of test files.
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

func setup(ctx context.Context) {
	shared.cfg = conf.DefaultIMS()
	shared.cfg.Core.JWTSecret = rand.Text()
	shared.cfg.Core.Admins = []string{"TestAdminRanger"}
	shared.cfg.Store.MySQL.Database = "ims"
	shared.cfg.Store.MySQL.Username = "rangers"
	shared.cfg.Store.MySQL.Password = rand.Text()
	shared.cfg.Core.Directory = conf.DirectoryTypeTestUsers
	shared.cfg.Directory.TestUsers = []conf.TestUser{
		{
			Handle:      "RealTestUserInConfig",
			Email:       "realtestuser@rangers.brc",
			Status:      "active",
			DirectoryID: 80808,
			// password is "password"
			Password:  "salt-and-pepper:7c5b8e2d772d79374609c5c480fa93ce45e4ac5a",
			Onsite:    true,
			Positions: nil,
			Teams:     nil,
		},
		{
			Handle:      "TestAdminRanger",
			Email:       "testadminranger@rangers.brc",
			Status:      "active",
			DirectoryID: 70707,
			// password is ")'("
			Password:  "SoSalty:7de49a1fc515ef8531a6247f3d3a23405c399fe9",
			Onsite:    true,
			Positions: nil,
			Teams:     nil,
		},
	}
	userStore, err := directory.NewUserStore(shared.cfg.Directory.TestUsers, nil)
	if err != nil {
		panic(err)
	}
	shared.userStore = userStore
	req := testcontainers.ContainerRequest{
		Image:        "mariadb:10.5.27",
		ExposedPorts: []string{"3306/tcp"},
		WaitingFor:   wait.ForListeningPort("3306/tcp"),
		Env: map[string]string{
			"MARIADB_RANDOM_ROOT_PASSWORD": "true",
			"MARIADB_DATABASE":             shared.cfg.Store.MySQL.Database,
			"MARIADB_USER":                 shared.cfg.Store.MySQL.Username,
			"MARIADB_PASSWORD":             shared.cfg.Store.MySQL.Password,
		},
	}
	mainTestInternal.imsDBContainer, err = testcontainers.GenericContainer(ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
			// This logging is useful for debugging container startup issues
			//Logger:           log.New(os.Stdout, "MariaDB ", log.LstdFlags),
		},
	)
	if err != nil {
		panic(err)
	}
	endpoint, err := mainTestInternal.imsDBContainer.Endpoint(ctx, "")
	if err != nil {
		panic(err)
	}
	port, _ := strconv.Atoi(strings.TrimPrefix(endpoint, "localhost:"))
	shared.cfg.Store.MySQL.HostPort = int32(port)
	shared.imsDB = &store.DB{DB: store.MariaDB(shared.cfg)}
	script := "BEGIN NOT ATOMIC\n" + store.CurrentSchema + "\nEND"
	_, err = shared.imsDB.ExecContext(ctx, script)
	if err != nil {
		panic(err)
	}
}

func shutdown(ctx context.Context) {
	_ = shared.imsDB.Close()
	err := mainTestInternal.imsDBContainer.Terminate(ctx)
	if err != nil {
		// log and continue
		slog.Error("Failed to terminate container", "error", err)
	}
}
