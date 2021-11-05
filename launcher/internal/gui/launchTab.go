package gui

import (
	"fmt"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	c "github.com/keinenclue/sasm-docker/launcher/internal/container"
)

var statusLabel *widget.Label = nil

func newLaunchTab(w fyne.Window) fyne.CanvasObject {
	hello := widget.NewLabel(`Welcome to the sasm docker launcher!
Before you start, make sure to have these tools installed:
- Docker
- XServer

Once you have them installed, just click Launch :D`)

	statusLabel = widget.NewLabel("")

	startButton := widget.NewButton("Launch!", nil)

	vBox := container.NewVBox(
		hello,
		statusLabel,
		startButton,
	)

	startButton.OnTapped = launchImage(startButton, vBox)
	return vBox
}

func launchAppendLog(level string, message string) {
	if level != "CONTAINER" {
		statusLabel.SetText(message)
	}
	appendLog(level, message)
}

func handleContainerEvent(layerProgress map[string]*widget.ProgressBar, vBox *fyne.Container, startButton *widget.Button) func(event c.ContainerEvent) {
	return func(event c.ContainerEvent) {
		switch event.Type {
		case c.ImagePullStatusChanged:
			s := event.Data.(c.ImagePullStatus)
			//fmt.Println(s)
			if s.Id == "" || s.ProgressDetail == nil {
				launchAppendLog("INFO", s.Status)
				return
			}

			if layerProgress[s.Id] == nil {
				pb := widget.NewProgressBar()
				layerProgress[s.Id] = pb
				pb.SetValue(0)
				vBox.AddObject(pb)
				vBox.Refresh()
			}
			//fmt.Printf("%s %s", s.Status, s.Progress)
			//
			pb := layerProgress[s.Id]
			currentMB := float64(s.ProgressDetail.Current) / 1000000
			totalMB := float64(s.ProgressDetail.Total) / 1000000
			value := currentMB / totalMB
			if math.IsNaN(value) {
				value = 1
			}
			pb.SetValue(value)
			pb.TextFormatter = func() string {
				if value == 1 {
					return fmt.Sprintf("%s: %s", s.Id, s.Status)
				} else {
					return fmt.Sprintf("%s: %s (%.2fMB/%.2fMB)", s.Id, s.Status, currentMB, totalMB)
				}
			}

			pb.Refresh()

		case c.ContainerStateChanged:
			state := event.Data.(c.ContainerState)

			launchAppendLog("INFO", fmt.Sprintf("Container state is now: %d", state))

			switch state {
			case c.ContainerOfflineState:
				endLogsSession()
				statusLabel.SetText("Sasm exited")
				startButton.Show()
			default:
				startButton.Hide()
			}
			if state != c.ContainerPullingState && len(layerProgress) > 0 {
				for _, bar := range layerProgress {
					vBox.Remove(bar)
				}
				layerProgress = map[string]*widget.ProgressBar{}
			}

		case c.ConsoleOutput:
			launchAppendLog("CONTAINER", event.Data.(string))
		}
	}
}

func launchImage(startButton *widget.Button, vBox *fyne.Container) func() {
	return func() {
		startButton.Hide()
		newLogSession()
		statusLabel.SetText("")

		launchAppendLog("INFO", "Starting ...")
		err, cont := c.NewSasmContainer()

		if err != nil {
			launchAppendLog("ERROR", err.Error())
			tabs.SelectTabIndex(1)
			return
		}

		layerProgress := make(map[string]*widget.ProgressBar)
		cont.OnContainerEvent(handleContainerEvent(layerProgress, vBox, startButton))
		err = cont.Launch()

		if err != nil {
			launchAppendLog("ERROR", err.Error())
			tabs.SelectTabIndex(1)
		}
	}
}