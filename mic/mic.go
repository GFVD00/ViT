package mic

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"sync/atomic"

	"Voice/edit"

	vosk "github.com/alphacep/vosk-api/go"
	"github.com/gen2brain/malgo"
)

type VoskResult struct {
	Text string `json:"text"`
}

func Dtxt() {
	_ = os.Remove(".temporary file")
}

func Rtxt() string {
	content, err := os.ReadFile(".temporary file")
	if err != nil {
		return ""
	}
	return string(content)
}

func Wtxt(text string, del bool) {
	mode := os.O_APPEND
	if del {
		mode = os.O_TRUNC
	}

	file, err := os.OpenFile(".temporary file", os.O_CREATE|os.O_WRONLY|mode, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	_, _ = file.WriteString(text)
}

func getAudioStream(ctx context.Context) (io.ReadCloser, func(), error) {
	if runtime.GOOS == "windows" {
		pr, pw := io.Pipe()
		samplesChan := make(chan []byte, 100)

		var active atomic.Bool
		active.Store(true)

		mCtx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
		if err != nil {
			return nil, nil, err
		}

		deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
		deviceConfig.Capture.Format = malgo.FormatS16
		deviceConfig.Capture.Channels = 1
		deviceConfig.SampleRate = 16000

		onRecvFrames := func(pOutput, pSamples []byte, frameCount uint32) {
			if !active.Load() || len(pSamples) == 0 {
				return
			}
			buf := make([]byte, len(pSamples))
			copy(buf, pSamples)

			select {
			case samplesChan <- buf:
			default:
			}
		}

		device, err := malgo.InitDevice(mCtx.Context, deviceConfig, malgo.DeviceCallbacks{Data: onRecvFrames})
		if err != nil {
			mCtx.Uninit()
			mCtx.Free()
			return nil, nil, err
		}

		_ = device.Start()

		go func() {
			for samples := range samplesChan {
				if !active.Load() {
					break
				}
				_, err := pw.Write(samples)
				if err != nil {
					break
				}
			}
		}()

		go func() {
			<-ctx.Done()
			active.Store(false)
			_ = pw.Close()
		}()

		cleanup := func() {
			active.Store(false)
			device.Uninit()
			mCtx.Uninit()
			mCtx.Free()
			_ = pw.Close()
		}

		return pr, cleanup, nil
	}

	cmd := exec.CommandContext(ctx, "arecord", "-f", "S16_LE", "-r", "16000", "-c", "1", "-t", "raw")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	}

	return stdout, cleanup, nil
}

func saveVoskResult(jsonStr string) {
	var res VoskResult
	if err := json.Unmarshal([]byte(jsonStr), &res); err == nil && res.Text != "" {
		finalText := edit.Start(res.Text)
		Wtxt(finalText, false)
	}
}

func Rec(ctx context.Context, model *vosk.VoskModel, audioStream io.ReadCloser) {
	rec, err := vosk.NewRecognizer(model, 16000.0)
	if err != nil {
		fmt.Println("Error creating recognizer:", err)
		return
	}
	defer rec.Free()

	if audioStream == nil {
		audio, cleanup, err := getAudioStream(ctx)
		if err != nil {
			fmt.Println("Audio stream error:", err)
			return
		}
		defer cleanup()
		audioStream = audio
	}

	buf := make([]byte, 4096)

	for {
		select {
		case <-ctx.Done():
			// Save the unparsed remainder of the phrase before exiting
			saveVoskResult(rec.FinalResult())
			return
		default:
		}

		n, err := audioStream.Read(buf)
		if err != nil {
			saveVoskResult(rec.FinalResult())
			return
		}

		if n > 0 {
			if rec.AcceptWaveform(buf[:n]) != 0 {
				saveVoskResult(rec.Result())
			}
		}
	}
}
