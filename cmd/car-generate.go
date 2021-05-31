package cmd

import (
	"github.com/minio/cli"
	"github.com/minio/mc/pkg/graphsplit"
)

var carGenerateCmd = cli.Command{
	Name:         "generate",
	Usage:        "Generate CAR files of the specified size",
	Action:       graphsplit.MainCarGenerate,
	OnUsageError: onUsageError,
	Before:       setGlobalsFromContext,
	Flags:        append(carGenerateFlags, globalFlags...),
}

var carGenerateFlags = []cli.Flag{
	cli.Int64Flag{
		Name:  "slice-size",
		Value: 17179869184, // 16G
		Usage: "specify chunk piece size",
	},
	cli.IntFlag{
		Name:  "parallel",
		Value: 4,
		Usage: "specify how many number of goroutines runs when generate file node",
	},
	cli.StringFlag{
		Name:  "graph-name",
		Usage: "specify graph name",
	},
	cli.StringFlag{
		Name:  "parent-path",
		Value: "",
		Usage: "specify graph parent path",
	},
	cli.StringFlag{
		Name:  "car-dir",
		Usage: "specify output CAR directory",
	},
}
