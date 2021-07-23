package cmd

import (
	"fmt"
	"github.com/filswan/fs3-mc/logs"
	"path/filepath"

	"github.com/minio/cli"
	"strings"
)

type ImportOnlineDealData struct {
	Bucket string
	Object string
}

//func NewOnlineDealData() *ImportOnlineDealData {
//return &ImportOnlineDealData{}
//}

var importCmd = cli.Command{
	Name:         "import",
	Usage:        "import online deal data",
	Action:       mainImport,
	OnUsageError: onUsageError,
	Before:       setGlobalsFromContext,
	Subcommands:  nil,
	Flags:        importFlags,
}

func mainImport(ctx *cli.Context) error {
	bucket, object := checkImportArgs(ctx)

	importConfigs := ImportOnlineDealData{
		Bucket: bucket,
		Object: object,
	}
	proposeImportData(importConfigs)

	return nil
}

var importFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "bucket",
		Usage: "specify which bucket to import",
	},
	cli.StringFlag{
		Name:  "object",
		Usage: "specify which object to import",
	},
}

func checkImportArgs(ctx *cli.Context) (string, string) {
	args := ctx.Args()
	for _, arg := range args {
		if strings.TrimSpace(arg) == "" {
			fatalIf(errInvalidArgument().Trace(args...), "Unable to validate empty argument.")
			return "", ""
		}
	}
	bucket := strings.TrimSpace(ctx.String("bucket"))
	object := strings.TrimSpace(ctx.String("object"))
	if len(bucket) == 0 {
		fatalIf(errInvalidArgument(), "please provide a valid bucket name")
		return "", ""
	}
	if len(object) == 0 {
		fatalIf(errInvalidArgument(), "please provide a valid objective name")
		return "", ""
	}

	return bucket, object
}

func proposeImportData(config ImportOnlineDealData) {
	fs3VolumeAddress := defaultVolumeAddress
	objectPath := filepath.Join(fs3VolumeAddress, config.Bucket, config.Object)
	//commandArgs = []string{"client", "import", objectPath}
	//dataCID, err := exec.Command("lotus",commandArgs...).Output()
	commandLine := "lotus " + "client " + "import " + objectPath
	dataCID, err := ExecCommand(commandLine)
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}
	outStr := strings.Fields(string(dataCID))
	dataCIDStr := outStr[len(outStr)-1]
	fmt.Println(fmt.Sprintf("Bucket: %s, Object: %s, DataCid: %s", config.Bucket, config.Object, dataCIDStr))
	return
}
