package terminal

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/Easy-Infra-Ltd/assert"
)

type TextAlignment int

const (
	AlignLeft TextAlignment = iota
	AlignCenter
	AlignRight
)

func moveTo(line int, column int) string {
	return fmt.Sprintf("\033[%d;%dH", line, column)
}

func moveCursorUp(n int) string {
	return fmt.Sprintf("\033[%dA", n)
}

func moveCursorDown(n int) string {
	return fmt.Sprintf("\033[%dB", n)
}

func clearLine() string {
	return "\033[K"
}

func clearScreen() string {
	return "\033[2J"
}

func getPaddedBytes(count int) []byte {
	result := make([]byte, count)
	for i := 0; i < count; i++ {
		result[i] = ' '
	}
	return result
}

func getAlignmentPadding(alignment TextAlignment) ([]byte, int) {
	thirdWidth := termWidth / 3
	switch alignment {
	case AlignLeft:
		return []byte{}, thirdWidth
	case AlignCenter:
		return getPaddedBytes(thirdWidth), thirdWidth * 2
	case AlignRight:
		return getPaddedBytes(thirdWidth * 2), termWidth
	default:
		return []byte{}, thirdWidth
	}
}

type Writer struct {
	Out       io.Writer
	Buf       bytes.Buffer
	lineCount int
	mtx       sync.Mutex
}

var termWidth int

func New(out io.Writer) *Writer {
	writer := &Writer{Out: out}

	if termWidth == 0 {
		termWidth, _ = writer.GetTerminalDimensions()
	}

	return writer
}

func (w *Writer) Reset() {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	w.Buf.Reset()
	w.lineCount = 0
}

func (w *Writer) AddLine() {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	w.lineCount++
}

func (w *Writer) Print(alignment TextAlignment) error {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	if len(w.Buf.Bytes()) == 0 {
		return nil
	}

	var currentLine bytes.Buffer
	for _, b := range w.Buf.Bytes() {
		lineWidth := termWidth
		if len(currentLine.Bytes()) == 0 {
			padding, alignWidth := getAlignmentPadding(alignment)
			lineWidth = alignWidth
			currentLine.Write(padding)
		}

		// TODO: Consider re-writing this so it wraps better, currently it doesn't consider words
		currentLine.Write([]byte{b})
		if b == '\n' {
			_, err := w.Out.Write(currentLine.Bytes())
			if err != nil {
				return err
			}

			w.lineCount++
			currentLine.Reset()
		} else if currentLine.Len() == lineWidth {
			currentLine.Write([]byte{'\n'})
			_, err := w.Out.Write(currentLine.Bytes())
			if err != nil {
				return err
			}

			w.lineCount++
			currentLine.Reset()
		}
	}

	currentLine.Write([]byte{'\n'})
	if _, err := w.Out.Write(currentLine.Bytes()); err != nil {
		return err
	}

	w.Buf.Reset()

	return nil
}

func (w *Writer) Write(b []byte) (int, error) {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	return w.Buf.Write(b)
}

func (w *Writer) Clear() {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	for i := 0; i < w.lineCount+1; i++ {
		if _, err := fmt.Fprint(w.Out, moveCursorUp(0)); err != nil {
			assert.NoError(err, "Should not cause an error when trying to move to previous line in buffer")
		}

		if _, err := fmt.Fprint(w.Out, clearLine()); err != nil {
			assert.NoError(err, "Should not cause an error when trying to clear text from buffer")
		}
	}
}

func (w *Writer) GetTerminalDimensions() (int, int) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 80, 25
	}

	splits := strings.Split(strings.Trim(string(out), "\n"), " ")
	height, errH := strconv.ParseInt(splits[0], 0, 0)
	width, errW := strconv.ParseInt(splits[1], 0, 0)
	if errH != nil || errW != nil {
		return 80, 25
	}

	return int(width), int(height)
}
