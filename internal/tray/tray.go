package tray

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/anna-oake/systray"
)

const (
	title     = "Framework Privacy Bar"
	iconOn    = "framework-privacy-bar-mic-on"
	iconOff   = "framework-privacy-bar-mic-off"
	iconError = "framework-privacy-bar-error"
)

type tray struct {
	mu           sync.Mutex
	ready        bool
	hasPending   bool
	pendingState bool
	pendingErr   error
	status       *systray.MenuItem
}

func New() *tray {
	return &tray{}
}

func (t *tray) Run() {
	systray.Run(t.onReady, func() {})
}

func (t *tray) UpdateState(microphoneConnected bool, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.ready {
		t.pendingState = microphoneConnected
		t.pendingErr = err
		t.hasPending = true
		return
	}

	t.applyState(microphoneConnected, err)
}

func (t *tray) onReady() {
	systray.SetTitle(title)
	systray.SetIconThemePath(os.Getenv("ICON_THEME_PATH"))

	t.status = systray.AddMenuItem("Microphone is Blocked", "")
	t.status.Disable()
	systray.AddSeparator()

	quit := systray.AddMenuItem("Quit", "")
	go func() {
		<-quit.ClickedCh
		systray.Quit()
	}()

	t.mu.Lock()
	t.ready = true
	hasPending := t.hasPending
	pendingState := t.pendingState
	pendingErr := t.pendingErr
	t.hasPending = false
	t.pendingErr = nil
	t.mu.Unlock()

	if hasPending {
		t.UpdateState(pendingState, pendingErr)
	} else {
		t.UpdateState(false, nil)
	}
}

func (t *tray) applyState(microphoneConnected bool, err error) {
	var status string
	if err != nil {
		status = err.Error()
	} else if microphoneConnected {
		status = "Microphone is On"
	} else {
		status = "Microphone is Blocked"
	}

	systray.SetTitle(title)
	systray.SetIconName(iconPath(microphoneConnected, err))
	systray.SetTooltip(status)
	t.status.SetTitle(status)
}

func iconPath(microphoneConnected bool, err error) string {
	name := iconOn
	if err != nil {
		name = iconError
	} else if !microphoneConnected {
		name = iconOff
	}

	themePath := os.Getenv("ICON_THEME_PATH")
	if themePath == "" {
		return name
	}

	return filepath.Join(themePath, "hicolor", "scalable", "status", name+".svg")
}
