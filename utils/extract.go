package utils

import (
	"context"
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

	handler := func(ctx context.Context, f archiver.File) error {
		path := filepath.Join(targetDir, f.NameInArchive)

		switch f.IsDir() {
		case true:
			// create a directory
			//fmt.Println("creating:   " + f.NameInArchive)
			err := os.MkdirAll(path, 0777)
			if err != nil {
				return err
			}

		case false:
			//log.Print("extracting: " + f.NameInArchive)

			// if is a symlink
			if f.LinkTarget != "" {
				os.Remove(path)
				err := os.Symlink(f.LinkTarget, path)
				if err != nil {
					return err
				}
				return nil
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
			err := os.MkdirAll(path, 0777)
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
			err := os.MkdirAll(filepath.Dir(path), 0777)
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
