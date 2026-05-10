package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	src, err := os.Open(fromPath)
	if err != nil {
		return fmt.Errorf("opening source file: %w", err)
	}
	defer src.Close()

	info, err := src.Stat()
	if err != nil {
		return fmt.Errorf("stating source file: %w", err)
	}

	fileSize := info.Size()

	if fileSize == 0 && !info.Mode().IsRegular() {
		return ErrUnsupportedFile
	}

	if offset > fileSize {
		return ErrOffsetExceedsFileSize
	}

	if offset > 0 {
		if _, err = src.Seek(offset, io.SeekStart); err != nil {
			return fmt.Errorf("seeking in source file: %w", err)
		}
	}

	bytesToCopy := fileSize - offset
	if limit > 0 && limit < bytesToCopy {
		bytesToCopy = limit
	}

	dst, err := os.Create(toPath)
	if err != nil {
		return fmt.Errorf("creating destination file: %w", err)
	}
	defer dst.Close()

	pb := ProgressWriter{Total: bytesToCopy, Current: 0}

	w := io.MultiWriter(dst, &pb)

	if _, err := io.CopyN(w, src, bytesToCopy); err != nil {
		return err
	}

	return nil
}
