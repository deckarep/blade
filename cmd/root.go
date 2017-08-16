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
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "blade",
	Short: "do blade stuff",
}

func init() {
	// TODO: scan through users recipe/ folder hierarchy looking for all *.blade.toml files
	// Upon finding the files, build a Cobra Command hierarchy and cache that data-structre in BoltDB
	// This way we don't have to do it each time...OR have an outside step that generates the hierarchy.

	// Blade is opinioned on how the recipe/ folder works
	// but not opinioned on how you structure it to match your infrastructure

	// Example:
	// recipe/infra-a/mail-server/deploy.blade.toml
	// recipe/infra-a/mail-server/restart.blade.toml
	// recipe/infra-a/mail-server/mail/check-service.blade.toml
	// recipe/infra-a/mail-server/mail/purge.blade.toml
	// recipe/infra-b/audit-services.blade.toml

	// In the example above, each folder is a subcommand that itself doesn't do anything but has more subcommands.
	// Each terminating blade.toml file is the actual command.

	// So from an API perspective:

	// To purge your mail services:
	// blade ssh infra-a mail-server mail purge

	// To deploy your mail-server
	// blade ssh infra-a mail-server deploy

	// To audit all your services under infrastructure b
	// blade ssh infra-b audit-services
}
