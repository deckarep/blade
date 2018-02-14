package recipe

import "github.com/spf13/cobra"

type BladeArgumentDetails struct {
	Value string
	Help  string

	flagValue string
	argName   string
}

// AttachFlag allows you to pass in a Cobra command if you'd like to attach an override
// flag in relation to this argument.
func (a *BladeArgumentDetails) AttachFlag(cobraCommand *cobra.Command) {
	cobraCommand.Flags().StringVarP(&a.flagValue, a.Value, "", "", a.Help+" (recipe flag)")
}

// FlagValue returns the applied flag which either came from a command-line override or
// from the Recipe itself.
func (a *BladeArgumentDetails) FlagValue() string {
	if a.flagValue != "" {
		return a.flagValue
	}
	return a.Value
}

// Name returns the argument name or what string is in the {{}} construct.
func (a *BladeArgumentDetails) Name() string {
	return a.argName
}

type BladeRecipeArguments map[string]*BladeArgumentDetails

type BladeRecipeHelp struct {
	Short string
	Long  string
	Usage string
}

type BladeRecipeOverrides struct {
	Concurrency int
	Port        int
	User        string
}

type BladeRecipeResilience struct {
	WaitDuration           string
	Retries                int
	RetryBackoffStrategy   string
	RetryBackoffMultiplier string // <-- this is a duration like 5s
	FailBatch              bool
}

type BladeRecipeYaml struct {
	Args BladeRecipeArguments

	Name     string
	Filename string

	Hosts      []string
	HostLookup string
	Exec       []string

	Help       *BladeRecipeHelp
	Overrides  *BladeRecipeOverrides
	Resilience *BladeRecipeResilience
}

func (yc *BladeRecipeYaml) HasOverrides() bool {
	return yc.Overrides != nil
}
