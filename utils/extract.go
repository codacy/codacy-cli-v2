package utils

import (
	"context"
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

func ExtractZip(archive *os.File, targetDir string) error {
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
