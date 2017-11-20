package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jamowei/senv"
	"github.com/spf13/cobra"
)

const hostDefault, portDefault, nameDefault, labelDefault string = "127.0.0.1", "8888", "application", "master"

var profile_default []string = []string{"default"}

var (
	host, port, name, label string
	profiles                []string
	override, verbose       bool
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "senv",
	Short: "Senv is a very fast config client for the spring cloud config server",
	Long: `A fast spring config client written in Go for receiving properties
from a spring cloud config server and
make them available via system environment variables`,
	Args: cobra.NoArgs,
	PreRun: func(cmd *cobra.Command, args []string) {
		if name == nameDefault {
			fmt.Fprintln(os.Stderr, "Warning: no application name given, using default 'application'")
		}
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := senv.NewConfig(host, port, name, profiles, label, formatKey, formatVal)
		if err := cfg.Fetch(); err != nil {
			return err
		}
		if err := cfg.Process(verbose); err != nil {
			return err
		}
		return setEnvVars(cfg.Properties, override)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&host, "host", hostDefault, "config-server host")
	rootCmd.PersistentFlags().StringVar(&port, "port", portDefault, "config-server port")
	rootCmd.PersistentFlags().StringVarP(&name, "name", "n", nameDefault, "spring.application.name")
	rootCmd.PersistentFlags().StringSliceVarP(&profiles, "profiles", "p", profile_default, "spring.active.profiles")
	rootCmd.PersistentFlags().StringVarP(&label, "label", "l", labelDefault, "config-repo label to be used")
	rootCmd.PersistentFlags().BoolVarP(&override, "override", "o", false, "overrides existing environment variables")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "show all received properties")
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
	out = strings.Replace(s, "\r\n", "", -1)
	out = strings.Replace(out, "\n", "", -1)
	return
}
