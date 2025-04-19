/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"database/sql"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/srabraham/ranger-ims-go/api"
	"github.com/srabraham/ranger-ims-go/conf"
	"github.com/srabraham/ranger-ims-go/directory"
	"github.com/srabraham/ranger-ims-go/store"
	"github.com/srabraham/ranger-ims-go/web"
	"log"
	"net/http"
	"time"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: runServer,
}

func runServer(cmd *cobra.Command, args []string) {
	fmt.Println("serve called")

	db := store.MariaDB()

	clubhouseDB := directory.MariaDB()

	mux := http.NewServeMux()
	api.AddToMux(mux, conf.Cfg, db, clubhouseDB)
	web.AddToMux(mux, conf.Cfg)
	//mux := api.CreateMux(context.Background(), db, clubhouseDB)

	addr := fmt.Sprintf("%v:%v", conf.Cfg.Core.Host, conf.Cfg.Core.Port)
	s := &http.Server{
		Addr:           addr,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Printf("I'm listening %v", addr)
	log.Fatal(s.ListenAndServe())
}

func valOrNil(v sql.NullInt16) *int16 {
	if v.Valid {
		return &v.Int16
	}
	return nil
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
