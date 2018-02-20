/*
Open Source Initiative OSI - The MIT License (MIT):Licensing
The MIT License (MIT)
Copyright (c) 2018 Ralph Caraveo (deckarep@gmail.com)
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
	cobraCommand.Flags().StringVarP(&a.flagValue, a.argName, "", "", a.Help+" (recipe flag)")
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

// func (yc *BladeRecipeYaml) OverridesDefined() bool {
// 	return yc.Overrides != nil
// }
