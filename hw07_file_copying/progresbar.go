package main

import (
	"fmt"
	"strings"
)

const (
	ProgrssWhidth = 50
)

type ProgressWriter struct {
	Total   int64 // Общий размер файла в байтах
	Current int64 // Сколько байт уже скопировано
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.Current += int64(n)
	pw.printProgress()
	return n, nil
}

func (pw *ProgressWriter) printProgress() {
	if pw.Total == 0 {
		return
	}

	percentage := float64(pw.Current) / float64(pw.Total) * 100

	bar := strings.Repeat("=", int(percentage*ProgrssWhidth/100))
	bar += strings.Repeat("-", ProgrssWhidth-len(bar))

	fmt.Printf("\rКопирование: [%s] %.2f%%", bar, percentage)
}
