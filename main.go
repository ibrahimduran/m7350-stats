package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"os"
	"strings"
	"time"

	ico "github.com/Kodeworks/golang-image-ico"
	"github.com/flopp/go-findfont"
	"github.com/getlantern/systray"
	"github.com/golang/freetype"
	"github.com/sqweek/dialog"
)

var (
	STATUS_COLOR_CONNECTING = color.RGBA{255, 255, 0, 255}
	STATUS_COLOR_CONNECTED  = color.RGBA{0, 255, 0, 255}
	STATUS_COLOR_FAILED     = color.RGBA{255, 0, 0, 255}
	STATUS_DEFAULT_TEXT     = "•••"

	ftc = initializeFreetype()
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	dev := NewM7350()
	period := time.Second * 5

	systray.SetIcon(generateTrayIcon(STATUS_DEFAULT_TEXT, STATUS_COLOR_CONNECTING))
	systray.SetTitle("M7350 Stats")
	systray.SetTooltip("M7350 Stats")

	mOperator := systray.AddMenuItem(STATUS_DEFAULT_TEXT, "Operator name")
	mOperator.Disable()

	mQuitOrig := systray.AddMenuItem("Quit", "Quit the whole app")
	go func() {
		<-mQuitOrig.ClickedCh
		systray.Quit()
	}()

	go func() {
		formatStatusText := func() string {
			return strings.Split(PrettyFormatDataSize(dev.Stats.Wan.DailyStatisticsBytes, 1, 0), " ")[0]
		}

		for {
			failed := false

			for range time.Tick(period) {
				statusColor := STATUS_COLOR_CONNECTING
				statusText := STATUS_DEFAULT_TEXT

				if dev.Stats.Wan.DailyStatisticsBytes != 0 {
					statusText = formatStatusText()
				}

				if failed {
					systray.SetIcon(generateTrayIcon(statusText, statusColor))
					time.Sleep(time.Second * 3)
					failed = false
				}

				err := dev.FetchStats()

				if err == nil {
					mOperator.SetTitle(dev.Stats.Wan.OperatorName)

					tooltip := fmt.Sprintf(
						"↑ %s  ↓ %s\nDaily: %s\nTotal: %s",
						PrettyFormatDataSize(dev.Stats.Wan.TxSpeedBytes, 2, 1),
						PrettyFormatDataSize(dev.Stats.Wan.RxSpeedBytes, 2, 1),
						PrettyFormatDataSize(dev.Stats.Wan.TotalStatisticsBytes, 3, 0),
						PrettyFormatDataSize(dev.Stats.Wan.DailyStatisticsBytes, 3, 0),
					)

					statusColor = STATUS_COLOR_CONNECTED
					if dev.Stats.Wan.DailyStatisticsBytes != 0 {
						statusText = formatStatusText()
					}

					systray.SetTooltip(tooltip)
				} else {
					statusColor = STATUS_COLOR_FAILED
					failed = true
				}

				systray.SetIcon(generateTrayIcon(statusText, statusColor))
			}
		}
	}()
}

func onExit() {
	os.Exit(0)
}

func initializeFreetype() *freetype.Context {
	ftc := freetype.NewContext()
	fontPath, err := findfont.Find("seguisb.ttf")

	if err != nil {
		dialog.Message("Font not found: %s", err).Error()
		os.Exit(1)
	}

	fontData, err := ioutil.ReadFile(fontPath)

	if err != nil {
		dialog.Message("Font read error: %s", err).Error()
		os.Exit(1)
	}

	font, err := freetype.ParseFont(fontData)

	if err != nil {
		dialog.Message("Font parse error: %s", err).Error()
		os.Exit(1)
	}

	ftc.SetDPI(72)
	ftc.SetFont(font)
	ftc.SetSrc(image.NewUniform(color.White))

	return ftc
}

func generateTrayIcon(text string, statusColor color.RGBA) []byte {
	size := 128
	start := image.Point{0, 0}
	end := image.Point{size, size}
	im := image.NewRGBA(image.Rectangle{start, end})

	margin := 40
	marginy := 0
	for x := margin; x < size-margin; x++ {
		for y := size - marginy; y > size-marginy-10; y-- {
			im.SetRGBA(x, y, statusColor)
		}
	}

	fs := 100.0
	pt := freetype.Pt(1, -5+int(ftc.PointToFixed(fs)>>6))

	ftc.SetDst(im)
	ftc.SetClip(im.Bounds())
	ftc.SetFontSize(fs)
	ftc.DrawString(text, pt)

	buf := new(bytes.Buffer)
	err := ico.Encode(buf, im)

	if err != nil {
		return nil
	}

	return buf.Bytes()
}
