package utils

import (
	"context"
	"fmt"
	"github.com/mholt/archiver/v4"
	"io"
	"os"
	"path/filepath"
)

func ExtractTarGz(archive *os.File, targetDir string) error {
	format := archiver.CompressedArchive{
		Compression: archiver.Gz{},
		Archival:    archiver.Tar{},
	}

	var symlinks []archiver.File // Collect symlinks for second pass

	handler := func(ctx context.Context, f archiver.File) error {
		path := filepath.Join(targetDir, f.NameInArchive)

		switch f.IsDir() {
		case true:
			// Create directory if it doesn't exist
			err := os.MkdirAll(path, 0777)
			if err != nil {
				return err
			}

		case false:
			// If it's a symlink, defer its creation
			if f.LinkTarget != "" {
				symlinks = append(symlinks, f)
				return nil // Skip creating the symlink for now
			}

			// Ensure the parent directory exists
			parentDir := filepath.Dir(path)
			err := os.MkdirAll(parentDir, 0777)
			if err != nil {
				return fmt.Errorf("failed to create parent directory: %v", err)
			}

			// Create file with original permissions
			w, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return fmt.Errorf("failed to create file: %v", err)
			}
			defer w.Close()

			stream, _ := f.Open()
			defer stream.Close()

			_, err = io.Copy(w, stream)
			if err != nil {
				return err
			}
		}

		return nil
	}

	// First Pass: Extract all files and directories (no symlinks yet)
	err := format.Extract(context.Background(), archive, nil, handler)
	if err != nil {
		return err
	}

	// Second Pass: Create symlinks
	for _, f := range symlinks {
		path := filepath.Join(targetDir, f.NameInArchive)

		// Remove existing file if any
		os.Remove(path)

		// Create the symlink
		err := os.Symlink(f.LinkTarget, path)
		if err != nil {
			fmt.Printf("Failed to create symlink: %s -> %s, Error: %v\n", path, f.LinkTarget, err)
		}
	}

	return nil
}
