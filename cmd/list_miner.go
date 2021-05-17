package cmd

import (
	"context"
	"fmt"
	"github.com/minio/cli"
	json "github.com/minio/colorjson"
	"github.com/minio/mc/pkg/probe"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type minerResponse struct {
	Data struct {
		Miner []MinerMessage `json:"miner"`
	} `json:"data"`
	Status string `json:"status"`
}

type MinerMessage struct {
	AdjustedPower        string      `json:"adjusted_power"`
	Location             string      `json:"location"`
	MaxPieceSize         string      `json:"max_piece_size"`
	MinPieceSize         string      `json:"min_piece_size"`
	MinerID              string      `json:"miner_id"`
	OfflineDealAvailable bool        `json:"offline_deal_available"`
	Price                string      `json:"price"`
	Score                interface{} `json:"score"`
	Status               string      `json:"status"`
	UpdateTimeStr        string      `json:"update_time_str"`
	VerifiedPrice        string      `json:"verified_price"`
}

func (m MinerMessage) JSON() string {
	m.Status = "success"
	jsonMessageBytes, e := json.MarshalIndent(m, "", " ")
	fatalIf(probe.NewError(e), "Unable to marshal into JSON.")

	return string(jsonMessageBytes)
}

func (m MinerMessage) String() string {
	message := fmt.Sprintf("%*s %*s %*s %*s %*s %*s %*s %s",
		-10, m.MinerID,
		-8, m.Status,
		-6, fmt.Sprintf("%v", m.Score),
		-15, m.AdjustedPower,
		-30, m.Price,
		-20, m.VerifiedPrice,
		-15, m.MinPieceSize,
		m.MaxPieceSize)
	return message
}

var listMinerCmd = cli.Command{
	Name:            "miner",
	Usage:           "get miner info from swan",
	Action:          mainSwanListMiner,
	Before:          setGlobalsFromContext,
	Flags:           globalFlags,
	HideHelpCommand: true,
	Subcommands:     adminServiceSubcommands,
}

const (
	swanMinerListUrl = "https://api.filswan.com/miners?limit=100&offset=0&status=Active&sort_by=score&order=ascending"
)

func closeResponse(resp *http.Response) {
	// Callers should close resp.Body when done reading from it.
	// If resp.Body is not closed, the Client's underlying RoundTripper
	// (typically Transport) may not be able to re-use a persistent TCP
	// connection to the server for a subsequent "keep-alive" request.
	if resp != nil && resp.Body != nil {
		// Drain any remaining Body and then close the connection.
		// Without this closing connection would disallow re-using
		// the same connection for future uses.
		//  - http://stackoverflow.com/a/17961593/4465767
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}
}

// mainSwanList is the handle for "mc list" command.
func mainSwanListMiner(cliCtx *cli.Context) error {
	ctx, cancelList := context.WithCancel(globalContext)
	defer cancelList()

	client := http.Client{Timeout: 5 * time.Second}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, swanMinerListUrl, nil)
	if err != nil {
		return nil
	}
	resp, err := client.Do(request)
	defer closeResponse(resp)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	response := minerResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	messageTitle := MinerMessage{
		MinerID:       "Miner",
		Status:        "Status",
		Score:         "Score",
		AdjustedPower: "Adjusted Power",
		Price:         "Price",
		VerifiedPrice: "VerifiedPrice",
		MinPieceSize:  "Min Piece Size",
		MaxPieceSize:  "Max Piece Size",
	}
	printMsg(messageTitle)

	for _, message := range response.Data.Miner {
		printMsg(message)
	}

	return nil
}
