package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const testdataDir = "testdata"

// helper: read file bytes
func readFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("readFile(%s): %v", path, err)
	}
	return data
}

// helper: temp dest file
func tempDest(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp("", "go-cp-test-*")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

func TestCopy(t *testing.T) {
	src := filepath.Join(testdataDir, "input.txt")

	t.Run("full file", func(t *testing.T) {
		dst := tempDest(t)
		if err := Copy(src, dst, 0, 0); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := readFile(t, src)
		got := readFile(t, dst)

		require.Equal(t, string(want), string(got))
	})

	t.Run("with offset", func(t *testing.T) {
		dst := tempDest(t)
		const offset = 10
		if err := Copy(src, dst, offset, 0); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := readFile(t, src)[offset:]
		got := readFile(t, dst)

		require.Equal(t, string(want), string(got))
	})

	t.Run("with limit", func(t *testing.T) {
		dst := tempDest(t)
		const limit = 15

		if err := Copy(src, dst, 0, limit); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := readFile(t, src)[:limit]
		got := readFile(t, dst)

		require.Equal(t, string(want), string(got))
	})

	t.Run("with offset and limit", func(t *testing.T) {
		dst := tempDest(t)
		const offset, limit = 5, 10

		if err := Copy(src, dst, offset, limit); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := readFile(t, src)[offset : offset+limit]
		got := readFile(t, dst)

		require.Equal(t, string(want), string(got))
	})

	t.Run("limit EOF", func(t *testing.T) {
		dst := tempDest(t)

		srcData := readFile(t, src)
		hugeLimit := int64(len(srcData)) * 100

		if err := Copy(src, dst, 0, hugeLimit); err != nil {
			t.Fatalf("unexpected error (limit > size must be valid): %v", err)
		}

		got := readFile(t, dst)

		require.Equal(t, string(srcData), string(got))
	})
}

func TestCopy_OffsetExceedsFileSize(t *testing.T) {
	src := filepath.Join(testdataDir, "input.txt")
	dst := tempDest(t)

	info, err := os.Stat(src)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	bigOffset := info.Size() + 1

	err = Copy(src, dst, bigOffset, 0)

	if !errors.Is(err, ErrOffsetExceedsFileSize) {
		t.Errorf("expected ErrOffsetExceedsFileSize, got: %v", err)
	}
}

func TestCopy_UnsupportedFile(t *testing.T) {
	dst := tempDest(t)

	err := Copy("/dev/urandom", dst, 0, 0)
	if !errors.Is(err, ErrUnsupportedFile) {
		t.Errorf("expected ErrUnsupportedFile, got: %v", err)
	}
}

func TestCopy_SourceNotFound(t *testing.T) {
	dst := tempDest(t)

	err := Copy("trestdata/nonexistent_file.txt", dst, 0, 0)
	if err == nil {
		t.Error("expected error for missing source file, got nil")
	}
}
