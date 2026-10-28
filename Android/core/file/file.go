package file

import (
	"fmt"
	"io"
	"sync"

	"fyne.io/fyne/v2"
)

const (
	textFileName  = "temporary.txt"
	voiceFileName = "voice.wav"
)

type FileManager struct {
	Storage fyne.Storage
	mu      sync.Mutex
}

func New(st fyne.Storage) *FileManager {
	return &FileManager{
		Storage: st,
	}
}

// Creates the file if it does not exist yet.
func (f *FileManager) ensureExists(name string) error {
	reader, err := f.Storage.Open(name)
	if err == nil {
		return reader.Close()
	}

	writer, createErr := f.Storage.Create(name)
	if createErr != nil {
		return fmt.Errorf(
			"failed to open %q (%v) or create it: %w",
			name, err, createErr,
		)
	}

	return writer.Close()
}

// Reads the file. Called under mu lock.
func (f *FileManager) readLocked(name string) (string, error) {
	if err := f.ensureExists(name); err != nil {
		return "File not found", err
	}

	reader, err := f.Storage.Open(name)
	if err != nil {
		return "Cannot open file", err
	}

	data, readErr := io.ReadAll(reader)
	closeErr := reader.Close()

	if readErr != nil {
		return "Cannot read file", readErr
	}
	if closeErr != nil {
		return "Cannot close file", closeErr
	}

	return string(data), nil
}

func (f *FileManager) replaceWriter(
	name string,
) (fyne.URIWriteCloser, error) {

	// If the file exists, close and remove it.
	r, err := f.Storage.Open(name)
	if err == nil {
		if err := r.Close(); err != nil {
			return nil, fmt.Errorf(
				"error closing %q: %w", name, err,
			)
		}

		if err := f.Storage.Remove(name); err != nil {
			return nil, fmt.Errorf(
				"error removing %q: %w", name, err,
			)
		}
	}

	// Create a new file.
	w, err := f.Storage.Create(name)
	if err != nil {
		return nil, fmt.Errorf(
			"error creating %q: %w", name, err,
		)
	}

	return w, nil
}

func (f *FileManager) writeLocked(
	name, text string,
) error {
	if err := f.ensureExists(name); err != nil {
		return err
	}

	writer, err := f.Storage.Save(name)
	if err != nil {
		return fmt.Errorf(
			"Storage.Save(%q): %w", name, err,
		)
	}

	n, writeErr := io.WriteString(writer, text)
	closeErr := writer.Close()

	if writeErr != nil {
		return writeErr
	}
	if n != len(text) {
		return io.ErrShortWrite
	}
	return closeErr
}

func (f *FileManager) Rewrite(text string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.writeLocked(textFileName, text); err != nil {
		return "", err
	}

	return text, nil
}

func (f *FileManager) AddToFile(newText string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	content, err := f.readLocked(textFileName)
	if err != nil {
		return err
	}

	return f.writeLocked(textFileName, content+newText)
}

func (f *FileManager) ReadFile() (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.readLocked(textFileName)
}

func (f *FileManager) SaveVoice(rc io.ReadCloser) error {
	if rc == nil {
		return fmt.Errorf("empty audio stream received")
	}
	defer rc.Close()

	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.ensureExists(voiceFileName); err != nil {
		return err
	}

	writer, err := f.Storage.Save(voiceFileName)
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(writer, rc)
	closeErr := writer.Close()

	if copyErr != nil {
		return copyErr
	}

	return closeErr
}

func (f *FileManager) ReadVoice() io.ReadCloser {
	reader, err := f.Storage.Open(voiceFileName)
	if err != nil {
		fmt.Println("Error reading voice.wav:", err)
		return nil
	}

	return reader
}
