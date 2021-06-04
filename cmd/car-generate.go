package cmd

import (
	"context"
	"github.com/filedrive-team/go-graphsplit"
	"github.com/minio/cli"
	"golang.org/x/xerrors"
)

var carGenerateCmd = cli.Command{
	Name:  "generate",
	Usage: "Generate CAR files of the specified size",
	Action: func(c *cli.Context) error {
		ctx := context.Background()
		parallel := c.Uint("parallel")
		sliceSize := c.Uint64("slice-size")
		parentPath := c.String("parent-path")
		carDir := c.String("car-dir")
		graphName := c.String("graph-name")
		if sliceSize == 0 {
			return xerrors.Errorf("Unexpected! Slice size has been set as 0")
		}

		targetPath := c.Args().First()
		var cb graphsplit.GraphBuildCallback

		cb = graphsplit.CommPCallback(carDir)

		return graphsplit.Chunk(ctx, int64(sliceSize), parentPath, targetPath, carDir, graphName, int(parallel), cb)
	},
	OnUsageError: onUsageError,
	Before:       setGlobalsFromContext,
	Flags:        append(carGenerateFlags, globalFlags...),
}

var carGenerateFlags = []cli.Flag{
	cli.Uint64Flag{
		Name:  "slice-size",
		Value: 17179869184, // 16G
		Usage: "specify chunk piece size",
	},
	cli.UintFlag{
		Name:  "parallel",
		Value: 2,
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
	cli.BoolTFlag{
		Name:  "save-manifest",
		Usage: "create a mainfest.csv in car-dir to save mapping of data-cids and slice names",
	},
}
