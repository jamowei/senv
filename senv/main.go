package main

import (
	"fmt"
	"github.com/jamowei/senv"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"bytes"
)

const hostDefault, portDefault, nameDefault, labelDefault = "127.0.0.1", "8888", "application", "master"

var profileDefault = []string{"default"}

var version = "0.0.0"
var date = "2018"

var (
	host, port, name, label          string
	profiles                         []string
	noSysEnv, json, verbose, content bool
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:     "senv [command]",
	Short:   "A native config-client for the spring-cloud-config-server",
	Version: "v" + version,
	Long: fmt.Sprintf(
		`v%s                             © %s Jan Weidenhaupt

Senv is a fast native config-client for a
spring-cloud-config-server written in Go`, version, date[:4]),
	Args: cobra.NoArgs,
}

func warningDefault(_ *cobra.Command, _ []string) {
	if name == nameDefault {
		fmt.Fprintln(os.Stderr, "warning: no application name given, using default 'application'")
	}
}

var envCmd = &cobra.Command{
	Use:   "env [command]",
	Short: "Fetches properties and sets them as environment variables",
	Long: `Fetches properties from the spring-cloud-config-server
and replaces the placeholder in the specified command.`,
	Example: `on config-server:
  spring.application.name="Senv"
Example call:
  senv env echo ${spring.application.name:default}  //prints 'Senv' or when config-server not reachable 'default'`,
	PreRun:       warningDefault,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := senv.NewConfig(host, port, name, profiles, label)
		if err := cfg.Fetch(json, verbose); err != nil {
			return err
		}
		if err := cfg.Process(); err != nil {
			return err
		}
		return runCommand(args, cfg.Properties, noSysEnv)
	},
}

func runCommand(args []string, props map[string]string, noSysEnv bool) error {
	if len(args) < 2 {
		return fmt.Errorf("not enough args passed")
	}
	repl := senv.SpringReplacer{Opener: "${", Closer: "}", Default: ":"}
	for i, arg := range args {
		if val, err := repl.Replace(arg, props); err == nil {
			args[i] = val
		}
	}
	cmd := exec.Command(args[0], args[1:]...)
	if !noSysEnv {
		cmd.Env = os.Environ()
	}
	var sout, serr bytes.Buffer
	cmd.Stdout = &sout
	cmd.Stderr = &serr
	err := cmd.Run()
	if err != nil {
		fmt.Println(serr.String())
	} else {
		fmt.Println(sout.String())
	}
	return err
}

var fileCmd = &cobra.Command{
	Use:          "file [filenames]",
	Short:        "Receives static file(s)",
	Long:         `Receives static file(s) from the spring-cloud-config-server`,
	Args:         cobra.MinimumNArgs(1),
	PreRun:       warningDefault,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := senv.NewConfig(host, port, name, profiles, label)
		if len(args) == 1 {
			return cfg.FetchFile(args[0], content, verbose)
		} else if len(args) > 1 {
			isErr := false
			for i, file := range args {
				if err := cfg.FetchFile(file, content, verbose); err != nil {
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
	envCmd.PersistentFlags().BoolVarP(&noSysEnv, "nosysenv", "s", false, "start without system-environment variables")
	envCmd.PersistentFlags().BoolVarP(&json, "json", "j", false, "print json to stdout")
	envCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose")
	fileCmd.PersistentFlags().BoolVarP(&content, "content", "c", false, "print file to stdout")
	rootCmd.PersistentFlags().StringVar(&host, "host", hostDefault, "configserver host")
	rootCmd.PersistentFlags().StringVar(&port, "port", portDefault, "configserver port")
	rootCmd.PersistentFlags().StringVarP(&name, "name", "n", nameDefault, "spring.application.name")
	rootCmd.PersistentFlags().StringSliceVarP(&profiles, "profiles", "p", profileDefault, "spring.active.profiles")
	rootCmd.PersistentFlags().StringVarP(&label, "label", "l", labelDefault, "config-repo label to be used")
	rootCmd.AddCommand(envCmd)
	rootCmd.AddCommand(fileCmd)
}

//func setEnvVars(props map[string]string, override bool) error {
//	for key, nVal := range props {
//		if oVal, exists := os.LookupEnv(key); exists && !override {
//			return fmt.Errorf("environment variable already exists: %s=%s", key, oVal)
//		}
//		if err := os.Setenv(key, nVal); err != nil {
//			return err
//		}
//	}
//	return nil
//}

//func formatKey(in string) (out string) {
//	out = strings.Replace(in, ".", "_", -1)
//	out = strings.ToUpper(out)
//	return
//}
//
//func formatVal(s string) (out string) {
//	out = strings.Replace(s, "\r\n", " ", -1)
//	out = strings.Replace(out, "\n", " ", -1)
//	return
//}
