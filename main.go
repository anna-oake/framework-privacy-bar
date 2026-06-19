package main

import (
	"time"

	"github.com/anna-oake/framework-privacy-bar/internal/ec"
	"github.com/anna-oake/framework-privacy-bar/internal/tray"
)

func main() {
	ui := tray.New()

	privacy, err := ec.PollMicrophonePrivacy(200 * time.Millisecond)
	if err != nil {
		ui.UpdateState(false, err)
	} else {
		go func() {
			for microphoneConnected := range privacy {
				ui.UpdateState(microphoneConnected, nil)
			}
		}()
	}

	ui.Run()
}
