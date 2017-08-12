package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	retries     int
	concurrency int
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

func init() {
	sshCmd.Flags().IntVarP(&concurrency, "concurrency", "c", 1, "Max concurrency when running ssh commands")
	sshCmd.Flags().IntVarP(&retries, "retries", "r", 3, "Number of times to retry until a successful command returns")
}

var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "ssh [command]",
	Long:  "ssh [command] will execute the [command] on all servers. If the command matches a recipe name first, the recipe will be used in place of the command.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("ssh command ran")
	},
}
