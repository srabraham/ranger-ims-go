package integration

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/srabraham/ranger-ims-go/auth"
	"github.com/srabraham/ranger-ims-go/conf"
	"github.com/srabraham/ranger-ims-go/store"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

var (
	imsDBContainer testcontainers.Container
	imsTestCfg     *conf.IMSConfig
	imsDB          *store.DB

	jwtAdmin, jwtNormalUser string
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
	imsTestCfg = conf.DefaultIMS()
	imsTestCfg.Core.JWTSecret = rand.Text()
	imsTestCfg.Core.Admins = []string{"AdminRanger"}
	imsTestCfg.Store.MySQL.Database = "ims"
	imsTestCfg.Store.MySQL.Username = "rangers"
	imsTestCfg.Store.MySQL.Password = rand.Text()
	imsTestCfg.Core.Directory = conf.DirectoryTypeTestUsers
	imsTestCfg.Directory.TestUsers = []conf.TestUser{
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
	}
	req := testcontainers.ContainerRequest{
		Image:        "mariadb:10.5.27",
		ExposedPorts: []string{"3306/tcp"},
		WaitingFor:   wait.ForListeningPort("3306/tcp"),
		Env: map[string]string{
			"MARIADB_RANDOM_ROOT_PASSWORD": "true",
			"MARIADB_DATABASE":             imsTestCfg.Store.MySQL.Database,
			"MARIADB_USER":                 imsTestCfg.Store.MySQL.Username,
			"MARIADB_PASSWORD":             imsTestCfg.Store.MySQL.Password,
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
	imsTestCfg.Store.MySQL.HostPort = int32(port)
	imsDB = &store.DB{DB: store.MariaDB(imsTestCfg)}
	script := "BEGIN NOT ATOMIC\n" + store.CurrentSchema + "\nEND"
	_, err = imsDB.ExecContext(ctx, script)
	if err != nil {
		panic(err)
	}

	jwtAdmin = auth.JWTer{SecretKey: imsTestCfg.Core.JWTSecret}.CreateJWT(
		imsTestCfg.Core.Admins[0], 65483, nil, nil, true, 1*time.Hour,
	)
	jwtNormalUser = auth.JWTer{SecretKey: imsTestCfg.Core.JWTSecret}.CreateJWT(
		"NonAdmin Ranger", 3289, nil, nil, true, 1*time.Hour,
	)
}

func shutdown(ctx context.Context) {
	_ = imsDB.Close()
	err := imsDBContainer.Terminate(ctx)
	if err != nil {
		// log and continue
		slog.Error("Failed to terminate container", "error", err)
	}
}
