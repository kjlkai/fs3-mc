package cmd

import "github.com/minio/cli"

var listCmdSubcommands = []cli.Command{
	listMinerCmd,
}

var listCmd = cli.Command{
	Name:         "list",
	Usage:        "list swan info",
	Action:       mainSwanList,
	OnUsageError: onUsageError,
	Before:       setGlobalsFromContext,
	Subcommands:  listCmdSubcommands,
}

// mainSwanList is the handle for "mc list" command.
func mainSwanList(ctx *cli.Context) error {
	commandNotFound(ctx, adminCmdSubcommands)
	return nil
}
