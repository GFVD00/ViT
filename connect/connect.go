package connect

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"Voice/mic"
	"Voice/edit"
	"Voice/jsonjob"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	vosk "github.com/alphacep/vosk-api/go"
)

var TextReceive = errors.New("Text received")

func getIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		fmt.Println("Error getting IP:", err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func checkTCP(addr string) bool {
	timeout := 500 * time.Millisecond
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return true
	}
	conn.Close()
	return false
}

func Start(ctx context.Context, model *vosk.VoskModel, cancelFunc context.CancelFunc, f fyne.Storage, textArea *widget.Entry) string {
	ip := getIP()
	addr := ip + ":8080"
	check := checkTCP(addr)
	if check == false {
		return ""
	}
	go startTCP(ctx, model, cancelFunc, f, textArea)
	return addr
}

func startTCP(ctx context.Context, model *vosk.VoskModel, cancelFunc context.CancelFunc, f fyne.Storage, textArea *widget.Entry) {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting TCP listener:", err)
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}

		go handle(conn, ctx, model, listener, f, textArea)
	}
}

func handle(conn net.Conn, ctx context.Context, model *vosk.VoskModel, listener net.Listener, f fyne.Storage, textArea *widget.Entry) {
	for {
		select {
		case <-ctx.Done():
			listener.Close()
			return
		default:
			err := receive(conn, ctx, model, f, textArea)
			if errors.Is(err, TextReceive) {
				continue
			}
			if err != nil {
				return
			}
			err = send(conn, ctx, textArea)
			if err != nil {
				return
			}
		}
	}
}

func send(conn net.Conn, ctx context.Context, textArea *widget.Entry) error {
	go func() {
		<-ctx.Done()
		return
	}()
	content := mic.Rtxt()
	fyne.Do(func() {
		textArea.SetText(content)
	})
	size := int64(len(content))
	if err := binary.Write(conn, binary.BigEndian, size); err != nil {
		return err
	}
	_, err := conn.Write([]byte(content))
	return err
}

func readJS(JS io.ReadCloser, f fyne.Storage) (int64, string, error) {
	defer JS.Close()
	err := jsonjob.CreateJson(f, "File.json")
	if err != nil {
		return 0, "", err
	}

	var d jsonjob.MetaData
	err = json.NewDecoder(JS).Decode(&d)
	if err != nil {
		return 0, "", err
	}
	err = jsonjob.WriteJson(f, "File.json", d)
	if err != nil {
		return 0, "", err
	}

	d, err = jsonjob.ReadJson[jsonjob.MetaData](f, "File.json")
	if err != nil {
		return 0, "", err
	}

	size := d.Size
	tp := d.Type
	return size, tp, nil
}

func receive(conn net.Conn, ctx context.Context, model *vosk.VoskModel, f fyne.Storage, textArea *widget.Entry) error {
	var size int64
	if err := binary.Read(conn, binary.BigEndian, &size); err != nil {
		return err
	}

	buf := make([]byte, size)

	if _, err := io.ReadFull(conn, buf); err != nil {
		return err
	}
	rc := io.NopCloser(bytes.NewReader(buf))
	finalSize, tp, err := readJS(rc, f)
	if err != nil {
		return err
	}

	if err := binary.Read(conn, binary.BigEndian, &finalSize); err != nil {
		fmt.Println("Error reading final size:", err)
		return err
	}

	Filebuf := make([]byte, finalSize)

	if _, err := io.ReadFull(conn, Filebuf); err != nil {
		return err
	}
	if tp == "Text" {
		finalText := edit.Start(string(Filebuf))
		mic.Wtxt(finalText, true)
		fyne.Do(func() {
			textArea.SetText(finalText)
		})
		return TextReceive
	}
	Voicerc := io.NopCloser(bytes.NewReader(Filebuf))
	mic.Rec(ctx, model, Voicerc)
	return nil
}
