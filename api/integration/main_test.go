package integration

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/srabraham/ranger-ims-go/conf"
	"github.com/srabraham/ranger-ims-go/store"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"testing"
)

var (
	imsDBContainer testcontainers.Container
	imsCfg         *conf.IMSConfig
	imsDB          *store.DB
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
			// This logging is useful for debugging container startup issues
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
	imsDB = &store.DB{DB: store.MariaDB(imsCfg)}
	script := "BEGIN NOT ATOMIC\n" + store.CurrentSchema + "\nEND"
	_, err = imsDB.ExecContext(ctx, script)
	if err != nil {
		panic(err)
	}
}

func shutdown(ctx context.Context) {
	_ = imsDB.Close()
	err := imsDBContainer.Terminate(ctx)
	if err != nil {
		// log and continue
		slog.Error("Failed to terminate container", "error", err)
	}
}
