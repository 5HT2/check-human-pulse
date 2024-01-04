package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/schollz/progressbar/v3"
)

var (
	defaultSeconds = flag.Int("s", 60, "Default number of seconds to count for")
)

type textViewWriter struct {
	textView *tview.TextView
}

func (tw *textViewWriter) Write(p []byte) (n int, err error) {
	tw.textView.SetText(string(p))
	return len(p), nil
}

func main() {
	flag.Parse()

	keypressCount := 0
	seconds := *defaultSeconds
	started := false
	finished := false
	start := time.Now()

	app := tview.NewApplication()

	// Create a text view for the progress bar
	progressTextView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetChangedFunc(func() {
			app.Draw()
		})

	// Create a text view for the keypress count
	currentRateView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetChangedFunc(func() {
			app.Draw()
		})

	// Create a text view for the keypress count
	keypressTextView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetChangedFunc(func() {
			app.Draw()
		})

	// Create a flex for the progress bar and keypress count
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(keypressTextView, 0, 22, false).
		AddItem(currentRateView, 0, 1, false).
		AddItem(progressTextView, 0, 1, true)

	finishCounting := func() {
		if finished {
			return
		}

		finished = true
		seconds := float64(time.Now().UnixMilli()-start.UnixMilli()) / 1000.0
		fmt.Fprintf(keypressTextView, "%.1f beats per minute, ran for %.1fs! :3\n",
			float64(keypressCount)*(60.0/seconds), seconds,
		)
	}

	//bar := progressbar.NewOptions(keypressCount,
	//	progressbar.OptionSetWriter(&textViewWriter{progressTextView}),
	//	progressbar.OptionEnableColorCodes(true),
	//	progressbar.OptionSetDescription("[cyan]Keypresses this second:[reset]"),
	//	progressbar.OptionSetTheme(progressbar.Theme{
	//		Saucer:        "[green]=[reset]",
	//		SaucerHead:    "[green]>[reset]",
	//		SaucerPadding: " ",
	//		BarStart:      "[",
	//		BarEnd:        "]",
	//	}))
	var cur *progressbar.ProgressBar
	var bar *progressbar.ProgressBar

	cur = progressbar.NewOptions(
		200,
		progressbar.OptionSetWriter(&textViewWriter{currentRateView}),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionOnCompletion(func() {
			if bar != nil {
				if !bar.IsFinished() {
					bar.Finish()
				}
			}
		}),
		progressbar.OptionShowIts(),
		progressbar.OptionSetTheme(progressbar.Theme{Saucer: "█", SaucerPadding: " ", BarStart: "|", BarEnd: "|"}),
	)

	bar = progressbar.NewOptions(
		seconds-1,
		progressbar.OptionSetWriter(&textViewWriter{progressTextView}),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionOnCompletion(func() {
			if cur != nil && !cur.IsFinished() {
				cur.Finish()
			}
		}),
		progressbar.OptionShowIts(),
		progressbar.OptionSetTheme(progressbar.Theme{Saucer: "█", SaucerPadding: " ", BarStart: "|", BarEnd: "|"}),
	)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if bar.IsFinished() || cur.IsFinished() {
			return event
		}

		if !started {
			started = true
			start = time.Now()

			go func() {
				for i := 0; i < seconds; i++ {
					if cur.IsFinished() {
						break
					}

					if !started {
						time.Sleep(20 * time.Millisecond)
						continue
					}

					progressTextView.ScrollToBeginning()
					bar.Set(i)
					time.Sleep(time.Second)

					if i >= seconds {
						finishCounting()
					}
				}
			}()
		}

		if event.Key() == tcell.KeyEnter {
			keypressTextView.ScrollToBeginning()
			_ = bar.Finish()
			finishCounting()
		} else {
			keypressCount++
			cur.Set(keypressCount)
		}
		return event
	})

	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}
}
