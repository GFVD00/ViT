package main

import (
	"context"
	_ "embed"
	"fmt"

	"Android/core/chanjob"
	"Android/core/connect"
	"Android/core/file"
	"Android/core/jsonjob"
	"Android/core/mic"
	"Android/core/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

//go:embed assets/mic.png
var micPNG []byte

//go:embed assets/check.png
var checkPNG []byte

var App fyne.App

func main() {
	myApp := app.NewWithID("myapp")
	files := file.New(myApp.Storage())
	window := myApp.NewWindow("Voice")
	window.Resize(fyne.NewSize(300, 200))

	err := jsonjob.CreateJson(files.Storage, "Device.json")
	if err != nil {
		fmt.Println("Error creating JSON:", err)
	}

	textArea := widget.NewMultiLineEntry()
	textArea.SetPlaceHolder("Voice text...")
	textArea.Wrapping = fyne.TextWrapWord
	icon := fyne.NewStaticResource("mic.png", micPNG)

	active := false
	var (
		btn        *widget.Button
		cancelFunc context.CancelFunc
		ch         chan *chanjob.Flags
	)

	status := chanjob.CreateStatus()
	check := fyne.NewStaticResource("check.png", checkPNG)
	checkBtn := widget.NewButtonWithIcon("", check, func() {
		newText, err := files.Rewrite(textArea.Text)
		if err != nil {
			fmt.Println("Error rewriting file:", err)
		}
		fyne.Do(func() {
			textArea.SetText(newText)
		})
		chanjob.ChangeChan(true, false, false, ch)
	})

	ch = chanjob.CreateChan()
	textButt := "Connect"
	openCon := widget.NewButton(textButt, func() {
		entryIp := widget.NewEntry()
		entryIp.SetPlaceHolder("Enter IP")
		entryName := widget.NewEntry()
		entryName.SetPlaceHolder("Enter name")
		d := dialog.NewCustomConfirm(
			"Connection",
			"Send",
			"Cancel",
			container.NewVBox(entryIp, entryName),
			func(sub bool) {
				if sub {
					go func() {
						ip := entryIp.Text
						name := entryName.Text
						connect.Start(ip, files, status, ch)
						var d jsonjob.Device
						d.Name = name
						d.Ip = ip
						err = jsonjob.WriteJson(files.Storage, "Device.json", d)
						if err != nil {
							fmt.Println("Error writing JSON:", err)
						}
					}()
				}
			},
			window,
		)
		d.Show()
	})

	devBtn := widget.NewButton("Devices", func() {
		ui.ShowDevice(&window, files, status, ch)
	})

	btn = widget.NewButtonWithIcon("Off", icon, func() {
		active = !active
		if active {
			var ctx context.Context
			ctx, cancelFunc = context.WithCancel(context.Background())
			btn.SetText("On")

			// Start recording and close channel upon complete termination
			go func() {
				err := mic.Rec(ctx, files)
				if err != nil {
					fmt.Println("Recording error:", err)
				}
			}()
		} else {
			if cancelFunc != nil {
				cancelFunc()
			}
			btn.SetText("Off")

			chanjob.ChangeChan(true, false, true, ch)
		}
		btn.Refresh()
	})

	go func() {
		for {
			status.WaitStatus()
			if status.Get() {
				text, err := files.ReadFile()
				if err != nil {
					status.Set(false)
					continue
				}
				fyne.Do(func() {
					textArea.SetText(text)
				})
				status.Set(false)
			}
		}
	}()

	panel := container.NewHBox(btn, openCon, devBtn)
	content := container.NewBorder(
		panel, checkBtn, nil, nil, textArea,
	)

	window.SetContent(content)

	window.SetCloseIntercept(func() {
		myApp.Storage().Remove(".temporary file")
		window.Close()
	})

	window.ShowAndRun()
}
