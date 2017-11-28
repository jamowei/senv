package main

import (
	"fmt"
	"github.com/jamowei/senv"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

const hostDefault, portDefault, nameDefault, labelDefault string = "127.0.0.1", "8888", "application", "master"

var profile_default []string = []string{"default"}

var version = "0.0.0"
var date = "2017"

var (
	host, port, name, label       string
	profiles                      []string
	override, json, envs, content bool
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "senv [command]",
	Short: "A native config-client for the spring-cloud-config-server",
	Long: fmt.Sprintf(
		`v%s                             © %s Jan Weidenhaupt

Senv is a fast native config-client for a
spring-cloud-config-server written in Go`, version, date[:4]),
	Args: cobra.NoArgs,
	PreRun: func(cmd *cobra.Command, args []string) {
		if name == nameDefault {
			fmt.Fprintln(os.Stderr, "warning: no application name given, using default 'application'")
		}
	},
}

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Sets all fetched properties as environment variables",
	Long: `Fetches desired properties from the spring-cloud-config-server
and sets them as local environment variables.

By that the property key must converted to uppercase and divided by underscore, e.g:
    spring.application.name="Senv" => SPRING_APPLICATION_NAME="Senv"
In on the same commandline you can now get the corresponding values, e.g:
	- in Windows:  echo %SPRING_APPLICATION_NAME%   //prints: Senv
	- in Linux:	   echo $SPRING_APPLICATION_NAME    //prints: Senv`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := senv.NewConfig(host, port, name, profiles, label, formatKey, formatVal)
		if err := cfg.Fetch(json); err != nil {
			return err
		}
		if err := cfg.Process(envs); err != nil {
			return err
		}
		return setEnvVars(cfg.Properties, override)
	},
}

var fileCmd = &cobra.Command{
	Use:   "file [file name]",
	Short: "Receives a file to the current working directory",
	Long: `Receives a desired file from the spring-cloud-config-server
and put it in the current directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := senv.NewConfig(host, port, name, profiles, label, formatKey, formatVal)
		return cfg.FetchFile(args[0], content)
	},
}

func init() {
	envCmd.PersistentFlags().BoolVarP(&override, "override", "o", false, "overrides existing environment variables")
	envCmd.PersistentFlags().BoolVarP(&json, "json", "j", false, "print json")
	envCmd.PersistentFlags().BoolVarP(&envs, "envs", "e", false, "print environment variables")
	fileCmd.PersistentFlags().BoolVarP(&content, "content", "c", false, "print file content")
	rootCmd.PersistentFlags().StringVar(&host, "host", hostDefault, "configserver host")
	rootCmd.PersistentFlags().StringVar(&port, "port", portDefault, "configserver port")
	rootCmd.PersistentFlags().StringVarP(&name, "name", "n", nameDefault, "spring.application.name")
	rootCmd.PersistentFlags().StringSliceVarP(&profiles, "profiles", "p", profile_default, "spring.active.profiles")
	rootCmd.PersistentFlags().StringVarP(&label, "label", "l", labelDefault, "config-repo label to be used")
	rootCmd.AddCommand(envCmd)
	rootCmd.AddCommand(fileCmd)
}

func setEnvVars(props map[string]string, override bool) error {
	for key, nVal := range props {
		if oVal, exists := os.LookupEnv(key); exists && !override {
			return fmt.Errorf("environment variable already exists: %s=%s", key, oVal)
		}
		if err := os.Setenv(key, nVal); err != nil {
			return err
		}
	}
	return nil
}

func formatKey(in string) (out string) {
	out = strings.Replace(in, ".", "_", -1)
	out = strings.ToUpper(out)
	return
}

func formatVal(s string) (out string) {
	out = strings.Replace(s, "\r\n", " ", -1)
	out = strings.Replace(out, "\n", " ", -1)
	return
}
