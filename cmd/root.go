/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "coolify-cli",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) {},
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

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.coolify-cli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
func Shellout(command string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), err
}

type RemoteCommandParams struct {
	command     string
	host        string
	exitOnError bool `default:"true"`
}

func remoteCommand(params RemoteCommandParams) string {
	if host == "localhost" {
		stdout, stderr, err := Shellout(params.command)
		if err != nil || stderr != "" {
			log.Fatalf("[ERROR](%v): %v with error status %v", params.command, stderr, err)
		}
		return stdout
	}
	var userAndHost = user + "@" + host
	var remoteCommand = "ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o PasswordAuthentication=no -o ServerAliveInterval=20 -o LogLevel=ERROR -o ControlMaster=auto -o ControlPersist=1m -o ConnectTimeout=" + strconv.Itoa(sshTimeout) + " " + userAndHost + " " + params.command
	if debug {
		log.Println(remoteCommand)
	}
	stdout, stderr, err := Shellout(remoteCommand)
	if (err != nil || stderr != "") && params.exitOnError {
		log.Fatalf("[ERROR](%v): %v with error status %v", params.command, stderr, err)
	}
	return stdout
}
