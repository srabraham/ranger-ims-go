package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/srabraham/ranger-ims-go/api"
	"github.com/srabraham/ranger-ims-go/conf"
	"github.com/srabraham/ranger-ims-go/directory"
	"github.com/srabraham/ranger-ims-go/store"
	"github.com/srabraham/ranger-ims-go/web"
	"log"
	"log/slog"
	"net/http"
	"time"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Launch the IMS server",
	Long: "Launch the IMS server\n\n" +
		"Configuration will be read from conf/imsd.toml, and can be overridden by environment variables.",
	Run: runServer,
}

func runServer(cmd *cobra.Command, args []string) {
	var logLevel slog.Level
	must(logLevel.UnmarshalText([]byte(conf.Cfg.Core.LogLevel)))
	slog.SetLogLoggerLevel(logLevel)

	log.Printf("Have config\n%v", conf.Cfg)

	var userStore *directory.UserStore
	var err error
	switch conf.Cfg.Core.Directory {
	case conf.DirectoryTypeClubhouseDB:
		userStore, err = directory.NewUserStore(nil, directory.MariaDB())
	case conf.DirectoryTypeTestUsers:
		userStore, err = directory.NewUserStore(conf.Cfg.Directory.TestUsers, nil)
	default:
		err = fmt.Errorf("unknown directory %v", conf.Cfg.Core.Directory)
	}
	must(err)
	imsDB := store.MariaDB(conf.Cfg)

	mux := http.NewServeMux()
	api.AddToMux(mux, conf.Cfg, &store.DB{DB: imsDB}, userStore)
	web.AddToMux(mux, conf.Cfg)

	addr := fmt.Sprintf("%v:%v", conf.Cfg.Core.Host, conf.Cfg.Core.Port)
	s := &http.Server{
		Addr:        addr,
		Handler:     mux,
		ReadTimeout: 30 * time.Second,
		// This needs to be long to support long-lived EventSource calls.
		// After this duration, a client will be disconnected and forced
		// to reconnect.
		WriteTimeout:   30 * time.Minute,
		MaxHeaderBytes: 1 << 20,
	}
	slog.Info("IMS server up-and-running", "address", addr)
	log.Fatal(s.ListenAndServe())
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
