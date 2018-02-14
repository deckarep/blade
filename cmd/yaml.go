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
	"log"

	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v1"
)

func init() {
	RootCmd.AddCommand(yamlCmd)
}

var yamlCmd = &cobra.Command{
	Use:   "yaml",
	Short: "yaml tests some yaml prototypes",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Here's some yaml")

		var data = `
hosts: ['blade-prod-a', 'blade-prod-b']
exec:
  - echo "how are you?"
  - echo "Dood"
overrides:
  concurrency: 5
  args:
    username:
      value: Ralph
      help: username is the user you want to use
`

		var yc YamlConfig
		err := yaml.Unmarshal([]byte(data), &yc)
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		fmt.Println(yc.HasOverrides())
	},
}

type ArgumentDetails struct {
	Value string
	Help  string
}

type Arguments map[string]ArgumentDetails

type YamlConfig struct {
	Hosts     []string
	Exec      []string
	Overrides *struct {
		Concurrency *int
		Args        *Arguments
	}
}

func (yc *YamlConfig) HasOverrides() bool {
	return yc.Overrides != nil
}
