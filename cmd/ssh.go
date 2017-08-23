/*
Open Source Initiative OSI - The MIT License (MIT):Licensing
The MIT License (MIT)
Copyright (c) 2017 Ralph Caraveo (deckarep@gmail.com)
Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

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
