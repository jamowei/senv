package main

import (
	"errors"
	"fmt"
	"github.com/jamowei/senv"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"syscall"
)

const hostDefault, portDefault, nameDefault, labelDefault = "127.0.0.1", "8888", "application", "master"

var profileDefault = []string{"default"}

var version = "0.0.0"
var date = "2018"

var (
	host, port, name, label, command string
	profiles, files                  []string
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

func init() {
	rootCmd.PersistentFlags().BoolVarP(&noSysEnv, "nosysenv", "s", false, "start without system-environment variables")
	rootCmd.PersistentFlags().BoolVarP(&json, "json", "j", false, "print json to stdout")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose")
	rootCmd.PersistentFlags().StringVar(&host, "host", hostDefault, "config-server host")
	rootCmd.PersistentFlags().StringVar(&port, "port", portDefault, "config-server port")
	rootCmd.PersistentFlags().StringVarP(&name, "name", "n", nameDefault, "spring.application.name")
	rootCmd.PersistentFlags().StringSliceVarP(&profiles, "profiles", "p", profileDefault, "spring.active.profiles")
	rootCmd.PersistentFlags().StringSliceVarP(&files, "files", "f", []string{}, "files to fetch from config-server")
	rootCmd.PersistentFlags().StringVarP(&label, "label", "l", labelDefault, "config-repo label to be used")
	rootCmd.PersistentFlags().StringVarP(&command, "command", "c", "", "command line to execute")
}

var rootCmd = &cobra.Command{
	Use:     "senv",
	Short:   "A native config-client for the spring-cloud-config-server",
	Version: "v" + version,
	Long: fmt.Sprintf(
		`v%s                             © %s Jan Weidenhaupt

Senv is a fast native config-client for a
spring-cloud-config-server written in Go`, version, date[:4]),
	Args:    cobra.NoArgs,
	PreRun:  warningDefault,
	PreRunE: validateArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := senv.NewConfig(host, port, name, profiles, label)
		if len(files) > 0 {
			if err := fetchFiles(files); err != nil {
				return err
			}
		}
		if command != "" {
			if err := cfg.Fetch(json, verbose); err != nil {
				return err
			}
			if err := cfg.Process(); err != nil {
				return err
			}
			return runCommand(command, cfg.Properties, verbose, noSysEnv)
		}
		return nil
	},
}

func validateArgs(_ *cobra.Command, _ []string) error {
	if command == "" && len(files) == 0 {
		return errors.New("no files or command given")
	}
	return nil
}

func warningDefault(_ *cobra.Command, _ []string) {
	if name == nameDefault {
		fmt.Println("warning: no application name given, using default 'application'")
	}
}

//var envCmd = &cobra.Command{
//	Use:   "env [command]",
//	Short: "Fetches properties and sets them as environment variables",
//	Long: `Fetches properties from the spring-cloud-config-server
//and replaces the placeholder in the specified command.`,
//	Example: `on config-server:
//  spring.application.name="Senv"
//Example call:
//  senv env echo ${spring.application.name:default}  //prints 'Senv' or when config-server not reachable 'default'`,
//	PreRun:       warningDefault,
//	SilenceUsage: true,
//	RunE: func(cmd *cobra.Command, args []string) error {
//		cfg := senv.NewConfig(host, port, name, profiles, label)
//		if err := cfg.Fetch(json, verbose); err != nil {
//			return err
//		}
//		if err := cfg.Process(); err != nil {
//			return err
//		}
//		return runCommand(args, cfg.Properties, noSysEnv)
//	},
//}

func runCommand(command string, props map[string]string, verbose bool, noSysEnv bool) error {
	repl := senv.SpringReplacer{Opener: "${", Closer: "}", Default: ":"}
	commandRepl, err := repl.Replace(command, props)
	if err != nil {
		return err
	}

	if verbose {
		fmt.Println("executing:", commandRepl)
	}

	cmd := exec.Command(commandRepl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if dir, err := os.Getwd(); err != nil {
		cmd.Dir = dir
	}

	if !noSysEnv {
		cmd.Env = os.Environ()
	}
	//var buffSout, buffSerr bytes.Buffer
	//cmd.Stdout = &buffSout
	//cmd.Stderr = &buffSerr
	//
	//serr := buffSerr.String()
	//sout := buffSout.String()

	err = cmd.Run()
	if err != nil {
		// try to get the exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			errExitCode = ws.ExitStatus()
		}
	} else {
		// success, errExitCode should be 0 if go is ok
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		errExitCode = ws.ExitStatus()
	}

	//fmt.Fprintln(os.Stderr, serr)
	//fmt.Fprintln(os.Stdout, sout)

	return err
}

func fetchFiles(files []string) error {
	cfg := senv.NewConfig(host, port, name, profiles, label)
	if len(files) == 1 {
		return cfg.FetchFile(files[0], content, verbose)
	} else if len(files) > 1 {
		isErr := false
		for i, file := range files {
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
}
