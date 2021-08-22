package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type M7350 struct {
	Stats M7350Stats
}

type M7350Stats struct {
	Wan M7350StatsWan `json:"wan"`
}

type M7350StatsWan struct {
	ConnectStatus        uint8  `json:"connectStatus"`
	Ipv4                 string `json:"ipv4"`
	TotalStatistics      string `json:"totalStatistics"`
	TotalStatisticsBytes int64
	DailyStatistics      string `json:"dailyStatistics"`
	DailyStatisticsBytes int64
	OperatorName         string `json:"operatorName"`
	RxSpeed              string `json:"rxSpeed"`
	RxSpeedBytes         int64
	TxSpeed              string `json:"txSpeed"`
	TxSpeedBytes         int64
}

func NewM7350() M7350 {
	return M7350{
		Stats: M7350Stats{
			Wan: M7350StatsWan{},
		},
	}
}

func (x *M7350) FetchStats() error {
	payload := strings.NewReader("{\"module\":\"status\",\"action\":0}")
	resp, err := http.Post("http://tplinkmifi.net/cgi-bin/web_cgi", "application/x-www-form-urlencoded", payload)

	if err != nil {
		fmt.Printf("Error sending request: %s", err)
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Printf("Error reading body: %s", err)
		return err
	}

	err = json.Unmarshal(body, &x.Stats)

	if err != nil {
		fmt.Printf("Error decoding json body: %s", err)
		return err
	}

	totalStatisticsFloat, _ := strconv.ParseFloat(x.Stats.Wan.TotalStatistics, 64)
	dailyStatisticsFloat, _ := strconv.ParseFloat(x.Stats.Wan.DailyStatistics, 64)

	x.Stats.Wan.TotalStatisticsBytes = int64(totalStatisticsFloat)
	x.Stats.Wan.DailyStatisticsBytes = int64(dailyStatisticsFloat)

	x.Stats.Wan.RxSpeedBytes, _ = strconv.ParseInt(x.Stats.Wan.RxSpeed, 10, 64)
	x.Stats.Wan.TxSpeedBytes, _ = strconv.ParseInt(x.Stats.Wan.TxSpeed, 10, 64)

	return nil
}

func PrettyFormatDataSize(amount int64, fractions, minExp int) string {
	const base = int64(1024)

	if amount < base && minExp == 0 {
		return fmt.Sprintf("%d B", amount)
	}

	div, exp := base, 0
	minExp -= 1

	for x := amount / base; x >= base || exp < minExp; x /= base {
		div *= base
		exp++
	}

	return fmt.Sprintf("%."+fmt.Sprint(fractions)+"f %ciB", float64(amount)/float64(div), "KMGTPE"[exp])
}
