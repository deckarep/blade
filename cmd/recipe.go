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

	"os"

	"github.com/deckarep/blade/lib/recipe"

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
	recipeCmd.AddCommand(recipeListCmd, recipeShowCmd, recipeValidateCmd, recipeTestCmd, recipeDumpCmd, recipeCreateCmd)
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

var recipeCreateCmd = &cobra.Command{
	Use:   "create [path/to/recipe]",
	Short: "creates a new recipe template file",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("This command will create a new recipe in the specified location.")
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

var recipeDumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "dumps a recipe",
	Run: func(cmd *cobra.Command, args []string) {
		s := &recipe.BladeRecipe{
			Required: &recipe.RequiredRecipe{
				Command: "hostname",
				Hosts:   []string{"blade.local", "blade.prod.local", "blade.integ.local"},
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

// End block for prototype.

func loadRecipe(path string) (*recipe.BladeRecipe, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		// TODO: errors.Wrap
		return nil, err
	}

	rec := recipe.NewRecipe()
	_, err = toml.Decode(string(b), rec)
	if err != nil {
		// TODO: errors.Wrap
		return nil, err
	}

	// The first _ in toml.Decode is a meta property with more interesting details.
	//fmt.Println(meta.IsDefined("PromptBanner"))

	return rec, nil
}
