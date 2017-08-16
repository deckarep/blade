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

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
	"os"
)

// TODO: what does a recipe look like?
// Recipe is the root structure that all recipes shall "inherit".
type Recipe struct {
	// Name is the name of this recipe.
	Name string
	// FilePath is the exact file location of this recipe.
	FilePath string
	// Concurrency is the bounded limit on number of parallel executions of a recipe.
	Concurrency int
	// Command is the actual command to be executed on the remote server.
	Command string
	// Retries is the number of allowed retries before we stop trying.
	Retries int
	// RetryBackOffStrategy is the strategy used for retrying and waiting such as: Constant, Exponential, etc.
	RetryBackStrategy string
	// FailBatch indicates whether a single failure should prevent the entire batch from executing.
	FailBatch bool
	// PauseDuration is a time.Duration string indicating how long to wait between executed commands.
	PauseDuration string
	// Hosts is a hardcoded list of hosts to execute on.
	Hosts []string
	// HostLookupCommand is an external command that you can provide to have Blade dynamically lookup before execution of this recipe.
	HostLookupCommand string
}

// StepRecipe is an ordered series of recipes that will be attempted in the specified order.
// The parameters specified in this recipe supercede the parameters in the individual recipe.
type StepRecipe struct {
	// Hmmm...does a step recipe inherit the properties above? Or does it have it's own similar specialized properties.
	Recipe
	Steps             []*Recipe
	StepPauseDuration string
}

type AggregateRecipe struct {
	Recipe
}

func init() {
	recipeCmd.AddCommand(recipeListCmd, recipeShowCmd, recipeValidateCmd, recipeTestCmd)
	RootCmd.AddCommand(recipeCmd)
}

var recipeValidateCmd = &cobra.Command{
	Use:   "validate [recipe-name]",
	Short: "validates that your recipe is correctly defined",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("I am validating your recipe")
	},
}

var recipeCmd = &cobra.Command{
	Use:   "recipe",
	Short: "recipe manages recipes including showing, validating, etc.",
}

var recipeListCmd = &cobra.Command{
	Use:   "list [search-constraint]",
	Short: "list lists all your recipes or those by a search constraint",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("list yo recipes")
	},
}

var recipeShowCmd = &cobra.Command{
	Use:   "show [recipe-name]",
	Short: "show shows a recipe by name",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("I am going to show that recipe if it exists or error out.")
	},
}

var recipeTestCmd = &cobra.Command{
	Use:   "test [recipe-name]",
	Short: "test is internally used for testing this code",
	Run: func(cmd *cobra.Command, args []string) {
		basicRecipe := &Recipe{
			FilePath:    "recipes/arsenic/mail-server/restart.blade.toml",
			Name:        "arsenic.mail-server.restart",
			Concurrency: 3,
			Command:     "hostname",
			Retries:     3,
			FailBatch:   false,
		}
		err := toml.NewEncoder(os.Stdout).Encode(basicRecipe)
		if err != nil {
			log.Fatal("Failed to encode your recipe sucka!")
		}
	},
}
