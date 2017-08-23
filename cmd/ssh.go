package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/exec"
)

var (
	sshKey  string
	sshPort int
	sshUser string
)

func init() {
	sshCmd.Flags().StringVarP(&sshUser, "user", "u", "root", "--user flag is the ssh user.")
	sshCmd.Flags().IntVarP(&sshPort, "port", "p", 22, "--port to utilize")
	sshCmd.Flags().StringVarP(&sshKey, "key", "i", "", "--key is the ssh key.")

	RootCmd.AddCommand(sshCmd)
}

var sshCmd = &cobra.Command{
	Use:   "ssh [host]",
	Short: "ssh [host]",
	Long:  "ssh [host] will log into specified host using ssh. If the command matches a recipe name first, the recipe will be used in place of the command.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Fatal("A single ssh [host] must be specified")
		}

		sshArgs := []string{
			"-p", fmt.Sprintf("%d", sshPort),
			"-o", "UserKnownHostsFile=/dev/null",
			"-o", "StrictHostKeyChecking=no",
			"-o", "LogLevel=quiet",
			fmt.Sprintf("%s@%s", sshUser, args[0]),
		}

		if sshKey != "" {
			sshArgs = append(sshArgs, "-i")
			sshArgs = append(sshArgs, sshKey)
		}

		currentCmd := exec.Command("ssh", sshArgs...)

		currentCmd.Stdin = os.Stdin
		currentCmd.Stderr = os.Stderr
		currentCmd.Stdout = os.Stdout

		if err := currentCmd.Run(); err != nil {
			log.Fatal(err)
		}
	},
}
