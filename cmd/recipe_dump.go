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
	"log"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/deckarep/blade/lib/recipe"
	"github.com/spf13/cobra"
)

func init() {
	recipeCmd.AddCommand(recipeDumpCmd)
}

var recipeDumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "dumps a recipe",
	Run: func(cmd *cobra.Command, args []string) {
		s := &recipe.BladeRecipe{
			Required: &recipe.RequiredRecipe{
				Commands: []string{"hostname"},
				Hosts:    []string{"blade.local", "blade.prod.local", "blade.integ.local"},
			},
			Overrides: &recipe.OverridesRecipe{
				Concurrency: 7,
				User:        "john",
			},
			Resilience: &recipe.ResilienceRecipe{
				WaitDuration:           "5s", // <-- time to sleep after command exits.
				Retries:                3,
				RetryBackoffStrategy:   "Exponential",
				RetryBackoffMultiplier: "5s",
				FailBatch:              true,
			},
			Help: &recipe.HelpRecipe{
				Usage: "boom",
				Short: "Does something cool",
				Long: `This recipe does something cool and also makes sure to blah...blah...blah. Also it's 
				supposed to ensure that blah blah and so you can be assured that it works great.`,
			},
			Interaction: &recipe.InteractionRecipe{
				Banner:       "Are you sure you want to continue?",
				PromptBanner: true,
				PromptColor:  "red",
			},
		}

		err := toml.NewEncoder(os.Stdout).Encode(&s)
		if err != nil {
			log.Fatal("Failed to encode your recipe sucka!")
		}
	},
}
