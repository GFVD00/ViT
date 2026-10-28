package mic

import (
	"context"
	"fmt"
	"io"
	"sync/atomic"

	"Android/core/file"

	"github.com/gen2brain/malgo"
)

func getVoice(ctx context.Context) (io.ReadCloser, func(), error) {
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

func Rec(ctx context.Context, f *file.FileManager) error {
	audioStream, cleanup, err := getVoice(ctx)
	if err != nil {
		fmt.Println("Audio stream error:", err)
		return err
	}
	defer cleanup()

	err = f.SaveVoice(audioStream)
	if err != nil {
		return err
	}
	return nil
}
