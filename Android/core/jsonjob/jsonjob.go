package jsonjob

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
)

type Device struct {
	Name string `json:"name"`
	Ip   string `json:"ip"`
}

type MetaData struct {
	Type string `json:"type"`
	Size int64  `json:"size"`
}

// "Device.json"
func CreateJson(store fyne.Storage, name string) error {
	file, err := store.Create(name)
	if err != nil {
		if errors.Is(err, storage.ErrAlreadyExists) {
			return nil
		}
		return err
	}
	defer file.Close()

	if _, err := io.WriteString(file, "{}\n"); err != nil {
		file.Close()
		store.Remove(name)
		return err
	}
	return nil
}

func WriteJson(store fyne.Storage, name string, T any) error {
	jsonData, err := json.MarshalIndent(T, "", " ")
	if err != nil {
		return err
	}

	jsonData = append(jsonData, '\n')
	file, err := store.Save(name)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		return err
	}
	return nil
}

func ReadJson[T any](store fyne.Storage, name string) (T, error) {
	var dev T

	file, err := store.Open(name)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return dev, err
	}
	defer file.Close()

	jsonData, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return dev, err
	}
	err = json.Unmarshal(jsonData, &dev)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return dev, err
	}
	return dev, nil
}
