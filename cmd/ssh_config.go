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

	"github.com/deckarep/blade/lib/ssh"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(sshConfigCmd)
}

var sshConfigCmd = &cobra.Command{
	Use:   "sshconfig",
	Short: "sshconfig",
	Long: `sshconfig will dump the user's ~/.ssh/config file. Currently used for debugging purposes.
		   Also since ssh config is not a standard no guarantee all ssh configs will parse`,
	Run: func(cmd *cobra.Command, args []string) {
		if ssh.ConfigMapping != nil {
			for _, host := range ssh.ConfigMapping {
				fmt.Printf("Host: %+v\n", host)
			}
			return
		}
		fmt.Println("User's ssh config was not parseable: Does the file ~/.ssh/config exist?")
	},
}
