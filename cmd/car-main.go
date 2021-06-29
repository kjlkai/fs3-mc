package cmd

import "github.com/minio/cli"

var carCmdSubcommands = []cli.Command{
	carGenerateCmd,
}

var carCmd = cli.Command{
	Name:         "car",
	Usage:        "generate car file for offline deal",
	Action:       mainCar,
	OnUsageError: onUsageError,
	Before:       setGlobalsFromContext,
	Subcommands:  carCmdSubcommands,
}

// mainSwanList is the handle for "mc list" command.
func mainCar(ctx *cli.Context) error {
	commandNotFound(ctx, adminCmdSubcommands)
	return nil
}
