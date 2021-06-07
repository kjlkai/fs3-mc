package cmd

import (
	"fmt"
	"github.com/minio/cli"
	csv "github.com/minio/minio/pkg/csvparser"
	"math/big"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var (
	defaultStart    = 7
	defaultDuration = 365
	epochPerHour    = 120
	GiBToByte       = 1024 * 1024 * 1024
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

func NewDealConfig() *OfflineDeal {
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

	for _, dealConfig := range dealConfigs {
		dealConfig.MinerId = miner
		dealConfig.SenderWallet = wallet
		dealConfig.StartEpoch = strconv.FormatUint(uint64(calculateStartEpoch(start)), 10)
		dealConfig.Duration = strconv.FormatUint(uint64(calculateDuration(duration)), 10)
		dealConfig.Cost = calculateCost(price, dealConfig.PieceSize).String()
		proposeOfflineDeal(dealConfig)
		writeCsv(fmt.Sprintf("%s.out", inputPath), *dealConfig)
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
	fmt.Println(pieceSizeInGiB.String())
	fmt.Println(price.String())

	cost := pieceSizeInGiB.Mul(pieceSizeInGiB, price)
	fmt.Println(cost)
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
		config.DealCid = string(stdout)
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
		dealConfig := &OfflineDeal{
			DataCid:   line[0],
			Filename:  line[1],
			PieceCid:  line[2],
			PieceSize: line[3],
		}
		dealConfigs = append(dealConfigs, dealConfig)
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
		file, err = os.Create(filePath)
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
		errorIf(errDummy(), err.Error())
	}
}
