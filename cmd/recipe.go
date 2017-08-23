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

// // StepRecipe is an ordered series of recipes that will be attempted in the specified order.
// // The parameters specified in this recipe supercede the parameters in the individual recipe.
// type StepRecipe struct {
// 	// Hmmm...does a step recipe inherit the properties above? Or does it have it's own similar specialized properties.
// 	Recipe
// 	Steps             []*Recipe
// 	StepPauseDuration string
// }

// type AggregateRecipe struct {
// 	Recipe
// }

func init() {
	recipeCmd.AddCommand(recipeListCmd, recipeShowCmd, recipeValidateCmd, recipeTestCmd, recipeDumpCmd)
	generateCommandLine()
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
		rec, err := loadRecipe("recipes/arsenic/mail-server/restart.blade.toml")

		err = toml.NewEncoder(os.Stdout).Encode(rec)
		if err != nil {
			log.Fatal("Failed to encode your recipe sucka!")
		}
	},
}

func NewRecipe() *BladeRecipe {
	return &BladeRecipe{
		Required:    &RequiredRecipe{},
		Overrides:   &OverridesRecipe{},
		Help:        &HelpRecipe{},
		Interaction: &InteractionRecipe{},
		Resilience:  &ResilienceRecipe{},
	}
}

// BladeRecipe is the root recipe type.
type BladeRecipe struct {
	Required    *RequiredRecipe
	Overrides   *OverridesRecipe
	Help        *HelpRecipe
	Interaction *InteractionRecipe
	Resilience  *ResilienceRecipe
}

// This block is for prototyping a good design.
type RequiredRecipe struct {
	Command           string
	Hosts             []string `toml:"Hosts,omitempty"`             // Must specify one or the other.
	HostLookupCommand string   `toml:"HostLookupCommand,omitempty"` // What happens if both are?
}

type OverridesRecipe struct {
	Concurrency int
	User        string
	// HostLookupCacheDisabled indicates that you want HostLookupCommand's to never be cached based on global settings.
	HostLookupCacheDisabled bool
	// HostLookupCacheDuration specifies the amount of time to utilize cache before refreshing the host list.
	HostLookupCacheDuration string
}

type HelpRecipe struct {
	Short string
	Long  string
	Usage string
}

type InteractionRecipe struct {
	Banner       string
	PromptBanner bool
	PromptColor  string
}

type ResilienceRecipe struct {
	WaitDuration           string
	Retries                int
	RetryBackoffStrategy   string
	RetryBackoffMultiplier string // <-- this is a duration like 5s
	FailBatch              bool
}

var recipeDumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "dumps a recipe",
	Run: func(cmd *cobra.Command, args []string) {
		s := &BladeRecipe{
			Required: &RequiredRecipe{
				Command: "hostname",
				Hosts:   []string{"blade.local", "blade.prod.local", "blade.integ.local"},
			},
			Overrides: &OverridesRecipe{
				Concurrency: 7,
				User:        "john",
			},
			Resilience: &ResilienceRecipe{
				WaitDuration:           "5s", // <-- time to sleep after command exits.
				Retries:                3,
				RetryBackoffStrategy:   "Exponential",
				RetryBackoffMultiplier: "5s",
				FailBatch:              true,
			},
			Help: &HelpRecipe{
				Usage: "boom",
				Short: "Does something cool",
				Long: `This recipe does something cool and also makes sure to blah...blah...blah. Also it's 
				supposed to ensure that blah blah and so you can be assured that it works great.`,
			},
			Interaction: &InteractionRecipe{
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

// End block for prototype.

func loadRecipe(path string) (*BladeRecipe, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		// TODO: errors.Wrap
		return nil, err
	}

	rec := NewRecipe()
	_, err = toml.Decode(string(b), rec)
	if err != nil {
		// TODO: errors.Wrap
		return nil, err
	}

	// The first _ in toml.Decode is a meta property with more interesting details.
	//fmt.Println(meta.IsDefined("PromptBanner"))

	return rec, nil
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
				log.Println("Found a broken recipe...skipping: ", err.Error())
				continue
			}

			// parts[1:] drop the /recipe part.
			for _, p := range parts[1:] {
				var currentCommand *cobra.Command

				// Known bug, we need to dedup these, but add them to a map based on their full path.
				// Reason is: if you have the same folder name in different hiearchies you'll collide.
				// Idea: Let user drop a .blade.toml file in a folder with a Short/Long
				// This way we can add docs to describe command hiearchies when user uses the --help system.
				recipeAlreadyFound := false
				if _, ok := commands[p]; !ok {
					// If not found create it.
					currentCommand = &cobra.Command{
						Use:   p,
						Short: currentRecipe.Help.Short,
						Long:  currentRecipe.Help.Long,
					}
					commands[p] = currentCommand
				} else {
					// If found use it.
					currentCommand = commands[p]
					recipeAlreadyFound = true
				}

				// If we're not a dir but a blade.toml...set it up to Run.
				if strings.HasSuffix(p, "blade.toml") {
					currentCommand.Run = func(cmd *cobra.Command, args []string) {
						// It's probably better to not compile the properties like this
						// And instead just kick off the work pointing to the .blade.toml file
						// This way, we defer loading of the file until the last minute of execution.
						// This means you can regenerate your command hiearchy but since it loads the .blade.toml
						// on demand, should you change the file you don't have to regenerate.
						// Therefore regenerating speeds up the startup of the program but doesn't bake in the end result of each command.
						// The generate command is only going to be useful when a user starts having on the order of 100s of commands.
						// fmt.Println("Recipe Command:", currentRecipe.Command)
						// fmt.Println("Hosts: ", currentRecipe.Hosts)
						// fmt.Println("HostsLookupCommand:", currentRecipe.HostLookupCommand)
						//fmt.Printf("current recipe: %+v\n", currentRecipe)
						startSSHSession(currentRecipe)
					}
				}

				// Only add recipe nodes we haven't already found.
				if !recipeAlreadyFound {
					if lastCommand == nil {
						recipeCmd.AddCommand(currentCommand)
					} else {
						lastCommand.AddCommand(currentCommand)
					}
				}

				lastCommand = currentCommand
			}
		}
	}
}
