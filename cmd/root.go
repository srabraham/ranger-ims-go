/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"github.com/srabraham/ranger-ims-go/conf"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ranger-ims-go",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
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

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
		//} else {
		//	// Find home directory.
		//	home, err := os.UserHomeDir()
		//	cobra.CheckErr(err)
		//
		//	// Search config in home directory with name ".ranger-ims-go" (without extension).
		//	viper.AddConfigPath(home)
		viper.SetConfigType("toml")
		//	viper.SetConfigName(".ranger-ims-go")
	}

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
	// If a config file is found, read it in.
	must(viper.ReadInConfig())
	log.Printf("Using config file: %v", viper.ConfigFileUsed())

	newCfg := conf.DefaultIMS()
	must(viper.Unmarshal(&newCfg))
	conf.Cfg = newCfg
}

func must(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func must2(_ any, err error) {
	must(err)
}
