package cmd

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/uuid"
	"github.com/minio/cli"
	"github.com/minio/mc/pkg/probe"
	csv "github.com/minio/minio/pkg/csvparser"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	defaultStart    = 7
	defaultDuration = 365
	epochPerHour    = 120
	GiBToByte       = 1024 * 1024 * 1024
	aliasName       = "swanminio"
)

type OfflineDeal struct {
	MinerId       string
	SenderWallet  string
	Cost          string
	PieceCid      string
	PieceSize     string
	DataCid       string
	Duration      string
	StartEpoch    string
	FastRetrieval bool
	DealCid       string
	Filename      string
}

func NewOfflineDeal() *OfflineDeal {
	return &OfflineDeal{FastRetrieval: true}
}

var sendCmd = cli.Command{
	Name:         "send",
	Usage:        "send filecoin deal",
	Action:       mainSend,
	OnUsageError: onUsageError,
	Before:       setGlobalsFromContext,
	Subcommands:  nil,
	Flags:        sendFlags,
}

func mainSend(ctx *cli.Context) error {

	wallet, start, duration, miner, inputPath, price := checkSendArgs(ctx)
	dealConfigs := readCsv(inputPath)
	asbInputPath, err := filepath.Abs(inputPath)
	if err != nil {
		fatalIf(errInvalidArgument().Trace(inputPath), "please provide a valid input path")
	}
	uid := uuid.New()
	uidStr := strings.Split(uid.String(), "-")[0]
	csvParentPath := filepath.Dir(asbInputPath)
	dealCsvPath := filepath.Join(csvParentPath, fmt.Sprintf("dealMetadata-%s.csv", uidStr))

	for _, dealConfig := range dealConfigs {
		dealConfig.MinerId = miner
		dealConfig.SenderWallet = wallet
		dealConfig.StartEpoch = strconv.FormatUint(uint64(calculateStartEpoch(start)), 10)
		dealConfig.Duration = strconv.FormatUint(uint64(calculateDuration(duration)), 10)
		dealConfig.Cost = calculateCost(price, dealConfig.PieceSize).String()
		proposeOfflineDeal(dealConfig)
		writeCsv(dealCsvPath, *dealConfig)
	}
	upload := ctx.Bool("upload")
	if upload {
		uploadCsv(dealCsvPath, fmt.Sprintf("%s/swan", aliasName), ctx)
	}

	return nil
}

var sendFlags = []cli.Flag{
	cli.StringFlag{
		Name:   "from",
		EnvVar: "FIL_WALLET",
		Usage:  "specify filecoin wallet to use, default $FIL_WALLET",
	},
	cli.UintFlag{
		Name:  "start",
		Value: uint(defaultStart),
		Usage: "specify days for miner to process the file, default: 7",
	},
	cli.UintFlag{
		Name:  "duration",
		Value: uint(defaultDuration),
		Usage: "specify length in day to store the file, default: 365",
	},
	cli.StringFlag{
		Name:  "input",
		Usage: "specify the path of the csv file from car generation",
	},
	cli.StringFlag{
		Name:  "price",
		Value: "0",
		Usage: "specify the deal price for each GiB of file, default: 0",
	},
	cli.BoolFlag{
		Name:  "upload",
		Usage: "specify whether upload the generated csv to minio or not, default: false",
	},
}

func checkSendArgs(ctx *cli.Context) (string, uint, uint, string, string, *big.Float) {
	args := ctx.Args()
	for _, arg := range args {
		if strings.TrimSpace(arg) == "" {
			fatalIf(errInvalidArgument().Trace(args...), "Unable to validate empty argument.")
		}
	}
	if len(args) < 1 {

	}
	miner := args[0]

	wallet := strings.TrimSpace(ctx.String("from"))
	start := ctx.Uint("start")
	duration := ctx.Uint("duration")
	input := strings.TrimSpace(ctx.String("input"))
	price := ctx.String("price")
	upload := ctx.Bool("upload")

	if len(input) == 0 {
		fatalIf(errInvalidArgument().Trace(input), "please provide a input path")
	} else {
		if _, err := os.Stat(input); os.IsNotExist(err) {
			fatalIf(errInvalidArgument().Trace(input), "please provide a valid input path")
		}
	}
	if len(wallet) == 0 {
		fatalIf(errInvalidArgument().Trace(wallet), "please provide a valid wallet")
	}
	if start == 0 {
		fatalIf(errInvalidArgument(), "please provide a valid length of start time in day")
	}
	if duration == 0 {
		fatalIf(errInvalidArgument(), "please provide a valid length of duration in day")
	}
	priceDecimal, _, err := big.ParseFloat(price, 10, 256, big.ToNearestEven)
	if err != nil {
		fatalIf(errInvalidArgument(), "please provide a valid price")
	}
	if upload {
		AccessKey := os.Getenv("ACCESS_KEY")
		SecretKey := os.Getenv("SECRET_KEY")
		Endpoint := os.Getenv("ENDPOINT")
		if !(strings.HasPrefix(Endpoint, "http") || strings.HasPrefix(Endpoint, "https")) {
			fatalIf(errInvalidArgument().Trace(Endpoint), "endpoint should start with 'http' or 'https'")
		}
		if len(AccessKey) == 0 {
			fatalIf(errInvalidArgument().Trace(AccessKey), "$ACCESS_KEY not provided")
		}
		if len(SecretKey) == 0 {
			fatalIf(errInvalidArgument().Trace(SecretKey), "$SECRET_KEY not provided")
		}
	}

	return wallet, start, duration, miner, input, priceDecimal
}

func getCurrentEpoch() uint {
	sec := time.Now().Unix()
	currentEpoch := uint((sec - 1598306471) / 30)
	return currentEpoch
}

func calculateStartEpoch(start uint) uint {
	startEpoch := getCurrentEpoch() + (start*24)*uint(epochPerHour)
	return startEpoch
}

func calculateCost(price *big.Float, pieceSize string) *big.Float {
	pieceSizeInt := new(big.Float)
	pieceSizeInt.SetString(pieceSize)
	pieceSizeInGiB := pieceSizeInt.Quo(pieceSizeInt, big.NewFloat(float64(GiBToByte))).Quo(pieceSizeInt, big.NewFloat(127)).Mul(pieceSizeInt, big.NewFloat(128))

	cost := pieceSizeInGiB.Mul(pieceSizeInGiB, price)
	return cost
}

func calculateDuration(duration uint) uint {
	epoch := duration * 24 * 3600 / 30
	return epoch
}

func proposeOfflineDeal(config *OfflineDeal) {

	var commandArgs []string
	commandArgs = []string{"client", "deal", "--from", config.SenderWallet, "--start-epoch", config.StartEpoch,
		fmt.Sprintf("--fast-retrieval=%s", strconv.FormatBool(config.FastRetrieval)), "--manual-piece-cid",
		config.PieceCid, "--manual-piece-size", config.PieceSize, config.DataCid, config.MinerId, config.Cost,
		config.Duration}

	cmd := exec.Command("lotus", commandArgs...)
	stdout, err := cmd.Output()
	if err != nil {
		errorIf(errDummy(), err.Error())
	} else {
		config.DealCid = strings.TrimSuffix(string(stdout), "\n")
	}
}

func readCsv(filepath string) []*OfflineDeal {
	csvFile, err := os.Open(filepath)
	if err != nil {
		fmt.Println(err)
	}
	defer csvFile.Close()

	csvLines, err := csv.NewReader(csvFile).ReadAll()
	if err != nil {
		errorIf(errDummy(), err.Error())
	}
	var dealConfigs []*OfflineDeal
	// playload_cid,filename,piece_cid,piece_size
	for i, line := range csvLines {
		if i == 0 {
			// skip header line
			continue
		}

		offlineDeal := NewOfflineDeal()

		offlineDeal.DataCid = line[0]
		offlineDeal.Filename = line[1]
		offlineDeal.PieceCid = line[2]
		offlineDeal.PieceSize = line[3]

		dealConfigs = append(dealConfigs, offlineDeal)
	}
	return dealConfigs
}
func writeCsv(filePath string, deal OfflineDeal) {
	_, err := os.Stat(filePath)
	header := []string{"DataCid", "filename", "PieceCid", "PieceSize", "DealCid"}
	var records [][]string
	var file *os.File
	if os.IsNotExist(err) {
		file, err = os.Create(filePath)
		records = append(records, header)
	} else {
		file, err = os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, 0644)
	}
	if err != nil {
		errorIf(errDummy(), err.Error())
	}
	defer file.Close()

	dealRecord := []string{deal.DataCid, deal.Filename, deal.PieceCid, deal.PieceSize, deal.DealCid}
	records = append(records, dealRecord)

	writer := csv.NewWriter(file)
	err = writer.WriteAll(records)
	if err != nil {
		errorIf(probe.NewError(err).Untrace(), err.Error())
	}
}

func makeFinalEnv(accessKey, secretKey, endpoint string) string {
	protocol := "http"
	if strings.HasPrefix(endpoint, "https") {
		protocol = "https"
	}
	url := strings.Split(endpoint, "://")[1]
	return fmt.Sprintf("%s://%s:%s@%s", protocol, accessKey, secretKey, url)
}

func uploadCsv(csvPath string, targetFolder string, cliCtx *cli.Context) {

	AccessKey := os.Getenv("ACCESS_KEY")
	SecretKey := os.Getenv("SECRET_KEY")
	Endpoint := os.Getenv("ENDPOINT")

	os.Setenv(fmt.Sprintf("MC_HOST_%s", aliasName), makeFinalEnv(AccessKey, SecretKey, Endpoint))

	ctx, cancelCopy := context.WithCancel(globalContext)
	defer cancelCopy()

	encKeyDB, err := getEncKeys(cliCtx)
	fatalIf(err, "Unable to parse encryption keys.")

	flagSet := flag.NewFlagSet("copy", flag.ExitOnError)
	flagSet.Parse([]string{csvPath, targetFolder})
	context := cli.NewContext(cliCtx.App, flagSet, cliCtx)

	var session *sessionV8

	e := doCopySession(ctx, cancelCopy, context, session, encKeyDB, false)
	if session != nil {
		session.Delete()
	}
	if e != nil {
		fatalIf(probe.NewError(e).Untrace(), e.Error())
	}
}
