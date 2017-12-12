package main

import (
	"fmt"
	"github.com/jamowei/senv"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

const hostDefault, portDefault, nameDefault, labelDefault string = "127.0.0.1", "8888", "application", "master"

var profileDefault = []string{"default"}

var version = "0.0.0"
var date = "2017"
//var showVersion bool

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
	Version: "v"+version,
	Long: fmt.Sprintf(
		`v%s                             © %s Jan Weidenhaupt

Senv is a fast native config-client for a
spring-cloud-config-server written in Go`, version, date[:4]),
	Args: cobra.NoArgs,
	//Run: func(cmd *cobra.Command, args []string) {
	//	if showVersion {
	//		fmt.Printf("Version: v%s            © %s Jan Weidenhaupt", version, date[:4])
	//	} else {
	//		cmd.Help()
	//	}
	//},
}

func warningDefault(cmd *cobra.Command, args []string) {
	if name == nameDefault {
		fmt.Fprintln(os.Stderr, "warning: no application name given, using default 'application'")
	}
}

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Fetches properties and sets them as environment variables",
	Long: `Fetches properties from the spring-cloud-config-server
and sets them as environment variables.
By that the property-key will be converted to uppercase and divided by underscore`,
	Example: `spring.application.name="Senv" => SPRING_APPLICATION_NAME="Senv"
In the same cmd env you can now get the corresponding values, e.g:
	- in Windows: echo %SPRING_APPLICATION_NAME%   //prints: Senv
	- in Linux:   echo $SPRING_APPLICATION_NAME    //prints: Senv`,
	PreRun: warningDefault,
	SilenceUsage: true,
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
	Use:   "file [filenames]",
	Short: "Receives static file(s)",
	Long: `Receives static file(s) from the spring-cloud-config-server`,
	Args: cobra.MinimumNArgs(1),
	PreRun: warningDefault,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := senv.NewConfig(host, port, name, profiles, label, formatKey, formatVal)
		if len(args) == 1 {
			return cfg.FetchFile(args[0], content)
		} else if len(args) > 1 {
			isErr := false
			for i, file := range args {
				if err := cfg.FetchFile(file, content); err != nil {
					isErr = true
					fmt.Fprintf(os.Stderr, "error receiving %v. file %s: %s\n", i+1, file, err)
				}
			}
			if isErr {
				return fmt.Errorf("error while receiving files")
			}
		}
		return nil
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
	rootCmd.PersistentFlags().StringSliceVarP(&profiles, "profiles", "p", profileDefault, "spring.active.profiles")
	rootCmd.PersistentFlags().StringVarP(&label, "label", "l", labelDefault, "config-repo label to be used")
	//rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "show version")
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
