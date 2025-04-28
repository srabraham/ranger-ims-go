package cmd

import (
	"fmt"
	"github.com/srabraham/ranger-ims-go/conf"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ranger-ims-go",
	Short: "The Ranger IMS server",
	Long:  "The Black Rock Ranger IMS server",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "conf/imsd.toml", "config file")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		_, err := os.Stat(cfgFile)
		// i.e. if the file does exist
		if !os.IsNotExist(err) {
			viper.SetConfigFile(cfgFile)
			viper.SetConfigType("toml")
		}
	}

	// e.g. we look for environment variables like IMS_CORE_LOGLEVEL,
	// which Viper sets as IMSConfig's Core.LogLevel value.
	const envPrefix = "IMS"
	viper.SetEnvPrefix(envPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	for _, e := range os.Environ() {
		split := strings.Split(e, "=")
		k := split[0]
		if strings.HasPrefix(k, envPrefix) {
			must(viper.BindEnv(strings.Join(strings.Split(k, "_")[1:], ".")))
		}
	}
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
		} else {
			panic(err)
		}
	} else {
		slog.Info("Using config file", "file", viper.ConfigFileUsed())
	}

	newCfg := conf.DefaultIMS()
	must(viper.Unmarshal(&newCfg))
	conf.Cfg = newCfg
	imsCfg := conf.Cfg

	must(imsCfg.Core.Directory.Validate())
	if imsCfg.Core.Deployment != "dev" {
		if imsCfg.Core.Directory == conf.DirectoryTypeTestUsers {
			must(fmt.Errorf("do not use TestUsers outside dev! A ClubhouseDB must be provided"))
		}
	}
}

func must(err error) {
	if err != nil {
		log.Panic(err)
	}
}
