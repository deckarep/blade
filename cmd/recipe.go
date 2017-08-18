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
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"os"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

// Idea: scan all files, and build a simple tree strucutre this way we can
// traverse the tree and build the Cobra.Command matching tree.
// or maybe we just do this in one shot...and we don't need this other representation.
type SubCommand struct {
	// name is the name of the current command.
	name string
	// child a reference to a child subcommand.
	children []*SubCommand
	// terminating indicates no child is defined and this command is actionable.
	terminating bool
}

// TODO: what does a recipe look like?
// Recipe is the root structure that all recipes shall "inherit".
type Recipe struct {
	// Use string
	Use string
	// Short string
	Short string
	// Long string
	Long string
	// Prompt indicates whether Blade should always prompt before continuing.
	Prompt bool
	// PromptBanner allows you to display a message before the user continues after the prompt.
	PromptBanner string
	// PromptColor allows you to color the PromptBanner to make things obvious before continuing.
	PromptColor string
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
	generateCommandLine()
	RootCmd.AddCommand(recipeCmd)

	// Sample model of commands and subcommands
	// root := &SubCommand{
	// 	name: "root recipe folder",
	// 	children: []*SubCommand{
	// 		name: "arsenic",
	// 		children: []*SubCommand{
	// 			name: "mail-server",
	// 			children: []*SubCommand{
	// 				name:        "restart",
	// 				terminating: true,
	// 			},
	// 		},
	// 	},
	// }
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
		rec, err := loadRecipe("recipes/arsenic/mail-server/restart.blade.toml")

		err = toml.NewEncoder(os.Stdout).Encode(rec)
		if err != nil {
			log.Fatal("Failed to encode your recipe sucka!")
		}
	},
}

func loadRecipe(path string) (*Recipe, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		// TODO: errors.Wrap
		return nil, err
	}

	var basicRecipe Recipe
	_, err = toml.Decode(string(b), &basicRecipe)
	if err != nil {
		// TODO: errors.Wrap
		return nil, err
	}

	// The first _ in toml.Decode is a meta property with more interesting details.
	//fmt.Println(meta.IsDefined("PromptBanner"))

	return &basicRecipe, nil
}

func generateCommandLine() {
	rootRecipeFolder := "recipes/"

	fileList := []string{}
	err := filepath.Walk(rootRecipeFolder, func(path string, f os.FileInfo, err error) error {
		fileList = append(fileList, path)
		return nil
	})

	if err != nil {
		log.Fatal("Failed to generate recipe data")
	}

	commands := make(map[string]*cobra.Command)

	// For now let's skip the global.blade.toml file.
	for _, file := range fileList {
		if strings.HasSuffix(file, ".blade.toml") &&
			!strings.Contains(file, "global.blade.toml") {
			parts := strings.Split(file, "/")
			var lastCommand *cobra.Command
			lastCommand = nil

			currentRecipe, err := loadRecipe(file)
			if err != nil {
				// TODO: don't fatal but skip recipe or log.
				log.Fatal("Found a broken recipe that we can't load: ", err.Error())
			}

			// parts[1:] drop the /recipe part.
			for _, p := range parts[1:] {
				var currentCommand *cobra.Command

				if _, ok := commands[p]; !ok {
					// If not found create it.
					currentCommand = &cobra.Command{
						Use:   p,
						Short: currentRecipe.Short,
						Long:  currentRecipe.Long,
					}
					commands[p] = currentCommand
				} else {
					// If found use it.
					currentCommand = commands[p]
				}

				// If we're not a dir but a blade.toml...set it up to Run.
				if strings.HasSuffix(p, "blade.toml") {
					currentCommand.Run = func(cmd *cobra.Command, args []string) {
						fmt.Println("Recipe Command:", currentRecipe.Command)
						fmt.Println("Hosts: ", currentRecipe.Hosts)
						fmt.Println("HostsLookupCommand:", currentRecipe.HostLookupCommand)
					}
				}

				if lastCommand == nil {
					recipeCmd.AddCommand(currentCommand)
				} else {
					lastCommand.AddCommand(currentCommand)
				}

				lastCommand = currentCommand
			}
		}
	}
}
