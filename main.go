package main

import (
	"context"
	"fmt"
	"log"
    "os"
	"strings"

	"Voice/mic"
	"Voice/connect"
	"Voice/doc"
	"Voice/edit"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	vosk "github.com/alphacep/vosk-api/go"
)

func FindVoskModel() (string, error) {
	entries, err := os.ReadDir(".")
	if err != nil {
		return "", fmt.Errorf("failed to read current directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "vosk-model") {
			return entry.Name(), nil
		}
	}

	return "", fmt.Errorf("no folder starting with 'vosk-model' found in the current directory")
}

func main() {
	vosk.SetLogLevel(-1)
    modelName,err:=FindVoskModel()
    if err !=nil{
      fmt.Println(err)
      return
    }
	// 1. Model is loaded ONCE when the application starts
	model, err := vosk.NewModel(modelName)
	if err != nil {
		log.Fatalf("Error loading Vosk model: %v", err)
	}
	defer model.Free()

	var (
		cancelFunc    context.CancelFunc
		cancelFuncNet context.CancelFunc
		filePath      string
	)

	myApp := app.NewWithID("myapp")
	window := myApp.NewWindow("Voice")
	window.Resize(fyne.NewSize(300, 200))

	textArea := widget.NewMultiLineEntry()
	textArea.SetPlaceHolder("Voice text...")
	textArea.Wrapping = fyne.TextWrapWord

	icon, err := fyne.LoadResourceFromPath("mic.png")
	if err != nil {
		fmt.Println("Error loading icon:", err)
	}

	active := false
	var btn *widget.Button
	var openCon *widget.Button

	btn = widget.NewButtonWithIcon("Off", icon, func() {
		active = !active
		if active {
			var ctx context.Context
			ctx, cancelFunc = context.WithCancel(context.Background())
			btn.SetText("On")

			// Start recording and close channel upon complete termination
			go func() {
				mic.Rec(ctx, model, nil)
			}()
		} else {
			if cancelFunc != nil {
				cancelFunc()
			}
			btn.SetText("Off")

			// Wait for recording to finish into a temporary file before outputting to the interface
			go func() {
				text := mic.Rtxt()
				fyne.Do(func() {
					textArea.SetText(text)
				})
			}()
		}
		btn.Refresh()
	})

	activeCon := false
	openCon = widget.NewButtonWithIcon("Connect Server", icon, func() {
		activeCon = !activeCon
		if activeCon {
			var netctx context.Context
			netctx, cancelFuncNet = context.WithCancel(context.Background())
			addr := connect.Start(netctx, model, cancelFuncNet, myApp.Storage(), textArea)
			openCon.SetText(addr)
		} else {
			if cancelFuncNet != nil {
				cancelFuncNet()
			}
			openCon.SetText("Connect Server")
		}
		openCon.Refresh()
	})

	check, err := fyne.LoadResourceFromPath("check.png")
	if err != nil {
		fmt.Println("Error loading icon:", err)
	}
	checkBtn := widget.NewButtonWithIcon("", check, func() {
		finalText := edit.Start(textArea.Text)
		mic.Wtxt(finalText, true)
	})

	pathLabel := widget.NewLabel("File not selected")
	browseBtn := widget.NewButton("Browse", func() {
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, window)
			}
			if reader == nil {
				return
			}
			defer reader.Close()

			filePath = reader.URI().Path()
			pathLabel.SetText("Selected file: " + filePath)
		}, window)

		extFilter := storage.NewExtensionFileFilter([]string{".docx"})
		fileDialog.SetFilter(extFilter)

		fileDialog.Show()
	})

	saveLabel := widget.NewLabel("")
	saveBtn := widget.NewButton("Save", func() {
		text := mic.Rtxt()
		doc.Create(text, filePath)
		saveLabel.SetText("File saved")
	})

	panel := container.NewHBox(btn, browseBtn, saveBtn, openCon)
	content := container.NewBorder(panel, checkBtn, nil, nil, textArea)
	window.SetContent(content)

	window.SetCloseIntercept(func() {
		mic.Dtxt()
		window.Close()
	})

	window.ShowAndRun()
}
