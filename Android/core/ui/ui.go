package ui

import (
	"fmt"

	"Android/core/chanjob"
	"Android/core/connect"
	"Android/core/file"
	"Android/core/jsonjob"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func ShowDevice(Window *fyne.Window, f *file.FileManager, status *chanjob.Status, ch chan *chanjob.Flags) {
	dev, err := jsonjob.ReadJson[jsonjob.Device](f.Storage, "Device.json")
	if err != nil {
		fmt.Println("Error reading device JSON:", err)
		return
	}

	deviceList := container.NewVBox()
	var btn *widget.Button
	btn = widget.NewButton(dev.Name, func() {
		go func() {
			connect.Start(dev.Ip, f, status, ch)
		}()
	})
	deviceList.Add(btn)

	newD := dialog.NewCustomConfirm(
		"Devices",
		"Connect",
		"Exit",
		deviceList,
		func(sub bool) {
			if sub && ch != nil {
			}
		},
		*Window,
	)
	newD.Show()
}
