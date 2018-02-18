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
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/deckarep/blade/lib/recipe"
	bladessh "github.com/deckarep/blade/lib/ssh"
	"github.com/spf13/cobra"
)

// Flag variables
var (
	retries     int
	concurrency int
	hosts       string
	port        int
	user        string
	quiet       bool
	verbose     bool
)

// blade ssh deploy-cloud-server-a // matches a recipe and therefore will follow the recipe guidelines against servers defined in recipe
// blade ssh deploy 127.0.0.1,128.0.0.1,129.0.0.1 //matches a recipe but uses the servers noted instead
// A recipe can be defined as the following
// 1. a single and simple command, with no servers specified, it's on you to provide a list.
// 2. a complex command, your recipe defines the server list.
// 3. a timed, tuned sequence of commands...perhaps it denotes a deploy order of multiple clusters, some sleep time, and perhaps even prompting to continue.
// 4. also, blade can cache server fetches...and has a default timeout
// 5. recipes are meant to be shared in some repository, this means that you can share with the team
// 6. recipes can be overriden per user, usually not a good idea...but possible still.
// 7. a recipe can be either:
//    a. a single command no servers specified/servers specified
//    b. a series of steps no servers specified/servers specified
//    c. a series of steps with sleep intervals baked in/continue prompt baked in
//    d. an aggregate, intended to execute on all servers but returning aggregates
//    e. banner prompts baked in, when dealing with production services (think: are you sure you want to continue?)
//    f. validation commands, intended on revealing some truthyness about infastructure
//    g. all recipes should emit proper log reports, and can tie into checklist as a service
//    h. integration into hipchat/slack/etc. when key events occur.
//    i. ability to view a simplified output of progress, notify end-user with gosx-notifier
//    j. ability to tie into ANY query infrastructure by allowing you to either pass in comma delimited servers
//          - or add something like this to your recipe: `ips prod redis-hosts -c`
// 8. blade generate where generate a database of all your recipes as Cobra Commands.
// 9. A seperate bolt database stores hosts cache for speed.

func init() {
	generateCommandLine()
	runCmd.PersistentFlags().StringVarP(&hosts,
		"servers", "s", "", "servers flag is one or more comma-delimited servers.")
	runCmd.PersistentFlags().IntVarP(&concurrency,
		"concurrency", "c", 0, "Max concurrency when running ssh commands")
	runCmd.PersistentFlags().IntVarP(&retries,
		"retries", "r", 3, "Number of times to retry until a successful command returns")
	runCmd.PersistentFlags().IntVarP(&port,
		"port", "p", 22, "The ssh port to use")
	runCmd.PersistentFlags().StringVarP(&user,
		"user", "u", "root", "user for ssh host login.")
	runCmd.PersistentFlags().BoolVarP(&quiet,
		"quiet", "q", false, "quiet mode will keep Blade as silent as possible.")
	runCmd.PersistentFlags().BoolVarP(&verbose,
		"verbose", "v", false, "verbose mode will keep Blade as verbose as possible.")
	RootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run [command]",
}

func validateFlags() {
	if concurrency < 0 {
		log.Fatal("The specified --concurrency flag must not be a negative number.")
	}

	if port < 22 {
		log.Fatal("The specified --port flag must be 22 or greater.")
	}

	if quiet && verbose {
		log.Fatal("You must specify either --quiet or --verbose but not both.")
	}

	if retries < 0 {
		log.Fatal("The specified --retries flag must not be a negative number.")
	}
}

func applyRecipeFlagOverrides(currentRecipe *recipe.BladeRecipeYaml, cobraCommand *cobra.Command) {
	if len(currentRecipe.Args) > 0 {
		for _, arg := range currentRecipe.Args {
			arg.AttachFlag(cobraCommand)
		}
	}
}

func applyFlagOverrides(recipe *recipe.BladeRecipeYaml, modifier *bladessh.SessionModifier) {
	if hosts != "" {
		modifier.FlagOverrides.Hosts = strings.Split(strings.TrimSpace(hosts), ",")
	}

	if concurrency > 0 {
		modifier.FlagOverrides.Concurrency = concurrency
	}

	if port > 0 {
		modifier.FlagOverrides.Port = port
	}
}

func searchFolders(folders ...string) []string {
	var results []string

	for _, folder := range folders {
		if _, err := os.Stat(folder); err == nil {
			results = append(results, walkFolder(folder)...)
		}
	}

	return results
}

func walkFolder(rootFolder string) []string {
	fileList := []string{}
	err := filepath.Walk(rootFolder, func(path string, f os.FileInfo, err error) error {
		fileList = append(fileList, path)
		return nil
	})

	if err != nil {
		log.Fatal("Failed to generate recipe data")
	}

	return fileList
}

func userHomeDir() string {
	const bladeFolder = ".blade"

	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return path.Join(home, bladeFolder)
	}
	return path.Join(os.Getenv("HOME"), bladeFolder)
}

func generateCommandLine() {
	fileList := searchFolders(userHomeDir(), "recipes")
	commands := make(map[string]*cobra.Command)

	// For now let's skip the global.blade.yaml file.
	for _, file := range fileList {
		if strings.HasSuffix(file, ".blade.yaml") &&
			!strings.Contains(file, "global.blade.yaml") {
			parts := strings.Split(file, "/")
			var lastCommand *cobra.Command
			lastCommand = nil

			currentRecipe, err := recipe.LoadRecipeYaml(file)
			if err != nil {
				log.Println("Found a broken recipe...skipping: ", err.Error())
				continue
			}

			// Find and drop all /recipes folders including all parent dirs.
			var recipeIndex = 0
			for index, part := range parts {
				if part == "recipes" {
					recipeIndex = index
					break
				}
			}

			remainingParts := parts[recipeIndex+1:]
			currentRecipe.Name = strings.TrimSuffix(strings.Join(remainingParts, "."), ".blade.toml")
			currentRecipe.Filename = file

			for _, p := range remainingParts {
				var currentCommand *cobra.Command

				// Known bug, we need to dedup these, but add them to a map based on their full path.
				// Reason is: if you have the same folder name in different hiearchies you'll collide.
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
				if strings.HasSuffix(p, "blade.yaml") {
					// Set the Use to just {recipe-name} of {recipe-name}.blade.toml.
					currentCommand.Use = strings.TrimSuffix(p, ".blade.yaml")
					applyRecipeFlagOverrides(currentRecipe, currentCommand)
					currentCommand.Run = func(cmd *cobra.Command, args []string) {
						// Apply validation of flags if used.
						validateFlags()

						modifier := bladessh.NewSessionModifier()

						// Apply flag overrides to the recipe here.
						applyFlagOverrides(currentRecipe, modifier)

						// Finally kick off session of requests.
						bladessh.StartSession(currentRecipe, modifier)
					}
				}

				// Only add recipe nodes we haven't already found.
				if !recipeAlreadyFound {
					if lastCommand == nil {
						runCmd.AddCommand(currentCommand)
					} else {
						lastCommand.AddCommand(currentCommand)
					}
				}

				lastCommand = currentCommand
			}
		}
	}
}
