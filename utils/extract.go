package utils

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mholt/archiver/v4"
)

func ExtractTarGz(archive *os.File, targetDir string) error {
	format := archiver.CompressedArchive{
		Compression: archiver.Gz{},
		Archival:    archiver.Tar{},
	}

	// Create a map to store symlinks for later creation
	symlinks := make(map[string]string)

	handler := func(ctx context.Context, f archiver.File) error {
		path := filepath.Join(targetDir, f.NameInArchive)

		switch f.IsDir() {
		case true:
			// create a directory
			err := os.MkdirAll(path, DefaultDirPerms)
			if err != nil {
				return err
			}

		case false:
			// if is a symlink, store it for later
			if f.LinkTarget != "" {
				symlinks[path] = f.LinkTarget
				return nil
			}

			// Ensure parent directory exists
			err := os.MkdirAll(filepath.Dir(path), DefaultDirPerms)
			if err != nil {
				return err
			}

			// write a file
			w, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}

			stream, _ := f.Open()
			defer stream.Close()

			_, err = io.Copy(w, stream)
			if err != nil {
				return err
			}
			w.Close()
		}

		return nil
	}

	err := format.Extract(context.Background(), archive, nil, handler)
	if err != nil {
		return err
	}

	// Create symlinks after all files have been extracted
	for path, target := range symlinks {
		// Remove any existing file/symlink
		os.Remove(path)

		// Ensure parent directory exists
		err := os.MkdirAll(filepath.Dir(path), DefaultDirPerms)
		if err != nil {
			return err
		}

		// Create the symlink
		err = os.Symlink(target, path)
		if err != nil {
			return fmt.Errorf("failed to create symlink %s -> %s: %w", path, target, err)
		}
	}

	return nil
}

// ExtractZip extracts a ZIP archive to the target directory
func ExtractZip(zipPath string, targetDir string) error {
	format := archiver.Zip{}

	handler := func(ctx context.Context, f archiver.File) error {
		path := filepath.Join(targetDir, f.NameInArchive)

		switch f.IsDir() {
		case true:
			// create a directory
			err := os.MkdirAll(path, DefaultDirPerms)
			if err != nil {
				return err
			}

		case false:
			// if is a symlink
			if f.LinkTarget != "" {
				os.Remove(path)
				err := os.Symlink(f.LinkTarget, path)
				if err != nil {
					return err
				}
				return nil
			}

			// ensure parent directory exists
			err := os.MkdirAll(filepath.Dir(path), DefaultDirPerms)
			if err != nil {
				return err
			}

			// write a file
			w, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer w.Close()

			stream, err := f.Open()
			if err != nil {
				return err
			}
			defer stream.Close()

			_, err = io.Copy(w, stream)
			if err != nil {
				return err
			}
		}

		return nil
	}

	file, err := os.Open(zipPath)
	if err != nil {
		return err
	}
	defer file.Close()

	err = format.Extract(context.Background(), file, nil, handler)
	if err != nil {
		return err
	}
	return nil
}
