package connect

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"

	"Android/core/chanjob"
	"Android/core/file"
	"Android/core/jsonjob"
)

var TextSend = errors.New("Text sent")

func Start(ip string, f *file.FileManager, status *chanjob.Status, ch chan *chanjob.Flags) {
	conn, err := net.Dial("tcp4", ip)
	if err != nil {
		fmt.Println("Connection error:", err)
		return
	}

	go handle(conn, ch, f, status)
}

func handle(conn net.Conn, ch chan *chanjob.Flags, f *file.FileManager, status *chanjob.Status) {
	defer conn.Close()

	for {
		select {
		case flag, ok := <-ch:
			if !ok {
				return
			}
			if flag.Send {
				err := send(conn, f, ch, flag)
				if errors.Is(err, TextSend) {
					continue
				}
				if err != nil {
					fmt.Println("Send error:", err)
				}
			}
			if flag.Receive {
				msg, err := receive(conn)
				if err != nil {
					fmt.Println("Receive error:", err)
				}
				_, err = f.Rewrite(msg)
				if err != nil {
					fmt.Println("Rewrite error:", err)
				}
				chanjob.ChangeChan(false, false, false, ch)
				status.Set(true)
			}
		}
	}
}

func prepareJson(conn net.Conn, f *file.FileManager, data []byte, t string) error {
	size := int64(len(data))
	err := jsonjob.CreateJson(f.Storage, "Data.json")
	if err != nil {
		return err
	}
	d := jsonjob.MetaData{
		Type: t,
		Size: size,
	}
	err = jsonjob.WriteJson(f.Storage, "Data.json", d)
	if err != nil {
		return err
	}
	file, err := f.Storage.Open("Data.json")
	if err != nil {
		return err
	}
	defer file.Close()

	byteJS, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	jsSize := int64(len(byteJS))
	memoryJS := io.NopCloser(bytes.NewReader(byteJS))
	defer memoryJS.Close()

	if err := binary.Write(conn, binary.BigEndian, jsSize); err != nil {
		return err
	}

	_, err = io.Copy(conn, memoryJS)
	f.Storage.Remove("Data.json")
	return nil
}

func sendContent(conn net.Conn, data []byte) error {
	size := int64(len(data))
	memoryRc := io.NopCloser(bytes.NewReader(data))
	defer memoryRc.Close()

	if err := binary.Write(conn, binary.BigEndian, size); err != nil {
		return err
	}

	_, err := io.Copy(conn, memoryRc)
	if err != nil {
		return err
	}
	return nil
}

func send(conn net.Conn, f *file.FileManager, ch chan *chanjob.Flags, flag *chanjob.Flags) error {
	if !flag.Voice {
		text, err := f.ReadFile()
		if err != nil {
			return err
		}
		r := strings.NewReader(text)
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}

		err = prepareJson(conn, f, data, "Text")
		if err != nil {
			return err
		}
		err = sendContent(conn, data)
		if err != nil {
			return err
		}
		chanjob.ChangeChan(false, false, false, ch)
		return TextSend
	}

	rc := f.ReadVoice()
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		return err
	}
	err = prepareJson(conn, f, data, "Voice")
	if err != nil {
		return err
	}
	err = sendContent(conn, data)
	if err != nil {
		return err
	}
	chanjob.ChangeChan(false, true, false, ch)
	return nil
}

func receive(conn net.Conn) (string, error) {
	var size int64

	if err := binary.Read(conn, binary.BigEndian, &size); err != nil {
		return "", err
	}

	buf := make([]byte, size)

	if _, err := io.ReadFull(conn, buf); err != nil {
		return "", err
	}
	return string(buf), nil
}
