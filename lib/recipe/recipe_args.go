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

package recipe

import "github.com/spf13/cobra"

// ArgumentsRecipe allows you to specify arguments for your commands.
type ArgumentsRecipe struct {
	Set []*Arg
}

// Arg is argument that can be applied to commands in the form of `{{arg-name}}.`
// Args can additionally be overriden via command-line flags.
type Arg struct {
	Arg       string
	flagValue string
	Value     string
	Help      string
}

// AttachFlag allows you to pass in a Cobra command if you'd like to attach an override
// flag in relation to this argument.
func (a *Arg) AttachFlag(cobraCommand *cobra.Command) {
	cobraCommand.Flags().StringVarP(&a.flagValue, a.Arg, "", "", a.Help+" (recipe flag)")
}

// FlagValue returns the applied flag which either came from a command-line override or
// from the Recipe itself.
func (a *Arg) FlagValue() string {
	if a.flagValue != "" {
		return a.flagValue
	}
	return a.Value
}
