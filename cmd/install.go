/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var debug bool
var host string
var user string
var sshTimeout int

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Coolify anywhere",
	Long:  `You can install Coolify anywhere you want. All you need is SSH access to the server.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Installing Coolify @", host)
		log.Println("Checking SSH connection.")

		whoAmI := remoteCommand(RemoteCommandParams{"whoami", host, true})
		log.Printf("Logged in as: '%v'", whoAmI)
		if whoAmI != "root" {
			log.Println("You need to be root to install Coolify.")
			return
		}

		log.Println("Checking operating system.")
		osType := remoteCommand(RemoteCommandParams{"cat /etc/os-release | grep -w 'ID' | cut -d '=' -f 2 | tr -d '\"'", host, true})
		log.Printf("OS type: %v", osType)

		osVersion := remoteCommand(RemoteCommandParams{"cat /etc/os-release | grep -w 'VERSION_ID' | cut -d '=' -f 2 | tr -d '\"'", host, true})
		log.Printf("OS version: %v", osVersion)

		log.Println("Checking if Docker is installed.")
		dockerVersion := remoteCommand(RemoteCommandParams{"docker -v", host, false})
		if dockerVersion == "" {
			log.Println("Docker is not installed. Installing Docker.")
			installDocker := remoteCommand(RemoteCommandParams{"curl -fsSL https://get.docker.com | sh ", host, false})
			log.Println(installDocker)
		}
		dockerVersion = remoteCommand(RemoteCommandParams{"docker -v", host, false})
		if dockerVersion == "" {
			log.Println("Docker installation failed.")
			return
		}
		log.Printf("Docker version: %v", dockerVersion)
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.PersistentFlags().BoolVar(&debug, "debug", false, "SSH connection timeout in seconds.")

	installCmd.PersistentFlags().IntVar(&sshTimeout, "connection-timeout", 5, "SSH connection timeout in seconds.")
	installCmd.PersistentFlags().StringVar(&host, "host", "", "Server IP address or DNS.")
	installCmd.PersistentFlags().StringVar(&user, "user", "root", "Username to use for SSH connection.")
	installCmd.MarkPersistentFlagRequired("host")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// installCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
