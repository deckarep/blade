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
	// blade infra-a mail-server mail purge

	// To deploy your mail-server
	// blade infra-a mail-server deploy

	// To audit all your services under infrastructure b
	// blade infra-b audit-services
}
