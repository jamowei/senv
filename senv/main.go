package main

import (
	"fmt"
	"github.com/jamowei/senv"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"bytes"
	"syscall"
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

var errExitCode = 0

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	} else if errExitCode > 0 {
		os.Exit(errExitCode)
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
	var buffSout, buffSerr bytes.Buffer
	cmd.Stdout = &buffSout
	cmd.Stderr = &buffSerr
	err := cmd.Run()

	serr := buffSerr.String()
	sout := buffSout.String()

	if err != nil {
		// try to get the exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			errExitCode = ws.ExitStatus()
		} else {
			// This will happen (in OSX) if `name` is not available in $PATH,
			// in this situation, exit code could not be get, and stderr will be
			// empty string very likely, so we use the default fail code, and format err
			// to string and set to stderr
			if serr == "" {
				serr = err.Error()
			}
		}
	} else {
		// success, errExitCode should be 0 if go is ok
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		errExitCode = ws.ExitStatus()
	}

	fmt.Fprintln(os.Stderr, serr)
	fmt.Fprintln(os.Stdout, sout)

	return nil
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
