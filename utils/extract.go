package utils

import (
	"codacy/cli-v2/constants"
	"codacy/cli-v2/utils/logger"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mholt/archiver/v4"
	"github.com/sirupsen/logrus"
)

func ExtractTarGz(archive *os.File, targetDir string) error {
	format := archiver.CompressedArchive{
		Compression: archiver.Gz{},
		Archival:    archiver.Tar{},
	}

	if err := os.MkdirAll(targetDir, constants.DefaultDirPerms); err != nil {
		logger.Error("Failed to create target directory", logrus.Fields{
			"directory": targetDir,
			"error":     err,
		})
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	if err := os.Chmod(targetDir, constants.DefaultDirPerms); err != nil {
		return fmt.Errorf("failed to set target directory permissions: %w", err)
	}

	handler := func(ctx context.Context, f archiver.File) error {
		path := filepath.Join(targetDir, f.NameInArchive)

		switch f.IsDir() {
		case true:
			if err := os.MkdirAll(path, constants.DefaultDirPerms); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", path, err)
			}
			if err := os.Chmod(path, constants.DefaultDirPerms); err != nil {
				return fmt.Errorf("failed to set directory permissions for %s: %w", path, err)
			}

		case false:
			parentDir := filepath.Dir(path)
			if err := os.MkdirAll(parentDir, constants.DefaultDirPerms); err != nil {
				return fmt.Errorf("failed to create parent directory %s: %w", parentDir, err)
			}
			if err := os.Chmod(parentDir, constants.DefaultDirPerms); err != nil {
				return fmt.Errorf("failed to set parent directory permissions for %s: %w", parentDir, err)
			}

			fileMode := os.FileMode(constants.DefaultFilePerms)
			if f.Mode()&0111 != 0 {
				fileMode = os.FileMode(constants.DefaultDirPerms)
			}
			w, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileMode)
			if err != nil {
				logger.Error("Failed to create file", logrus.Fields{
					"file":  path,
					"error": err,
				})
				return fmt.Errorf("failed to create file %s: %w", path, err)
			}

			stream, err := f.Open()
			if err != nil {
				w.Close()
				return fmt.Errorf("failed to open file stream for %s: %w", path, err)
			}

			_, err = io.Copy(w, stream)
			stream.Close()
			w.Close()
			if err != nil {
				return fmt.Errorf("failed to copy file contents for %s: %w", path, err)
			}

			if err := os.Chmod(path, fileMode); err != nil {
				return fmt.Errorf("failed to set file permissions for %s: %w", path, err)
			}
		}

		return nil
	}

	err := format.Extract(context.Background(), archive, nil, handler)
	if err != nil {
		logger.Error("Failed to extract archive", logrus.Fields{"error": err})
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	return nil
}

// ExtractZip extracts a ZIP archive to the target directory
func ExtractZip(zipPath string, targetDir string) error {
	format := archiver.Zip{}

	if err := os.MkdirAll(targetDir, constants.DefaultDirPerms); err != nil {
		logger.Error("Failed to create target directory", logrus.Fields{
			"directory": targetDir,
			"error":     err,
		})
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	if err := os.Chmod(targetDir, constants.DefaultDirPerms); err != nil {
		return fmt.Errorf("failed to set target directory permissions: %w", err)
	}

	handler := func(ctx context.Context, f archiver.File) error {
		path := filepath.Join(targetDir, f.NameInArchive)

		switch f.IsDir() {
		case true:
			if err := os.MkdirAll(path, constants.DefaultDirPerms); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", path, err)
			}
			if err := os.Chmod(path, constants.DefaultDirPerms); err != nil {
				return fmt.Errorf("failed to set directory permissions for %s: %w", path, err)
			}

		case false:
			parentDir := filepath.Dir(path)
			if err := os.MkdirAll(parentDir, constants.DefaultDirPerms); err != nil {
				return fmt.Errorf("failed to create parent directory %s: %w", parentDir, err)
			}
			if err := os.Chmod(parentDir, constants.DefaultDirPerms); err != nil {
				return fmt.Errorf("failed to set parent directory permissions for %s: %w", parentDir, err)
			}

			fileMode := os.FileMode(constants.DefaultFilePerms)
			if f.Mode()&0111 != 0 {
				fileMode = os.FileMode(constants.DefaultDirPerms)
			}
			w, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileMode)
			if err != nil {
				logger.Error("Failed to create file", logrus.Fields{
					"file":  path,
					"error": err,
				})
				return fmt.Errorf("failed to create file %s: %w", path, err)
			}

			stream, err := f.Open()
			if err != nil {
				w.Close()
				return fmt.Errorf("failed to open file stream for %s: %w", path, err)
			}

			_, err = io.Copy(w, stream)
			stream.Close()
			w.Close()
			if err != nil {
				return fmt.Errorf("failed to copy file contents for %s: %w", path, err)
			}

			if err := os.Chmod(path, fileMode); err != nil {
				return fmt.Errorf("failed to set file permissions for %s: %w", path, err)
			}
		}

		return nil
	}

	file, err := os.Open(zipPath)
	if err != nil {
		logger.Error("Failed to open zip file", logrus.Fields{
			"file":  zipPath,
			"error": err,
		})
		return fmt.Errorf("failed to open zip file %s: %w", zipPath, err)
	}
	defer file.Close()

	err = format.Extract(context.Background(), file, nil, handler)
	if err != nil {
		logger.Error("Failed to extract zip archive", logrus.Fields{
			"file":  zipPath,
			"error": err,
		})
		return fmt.Errorf("failed to extract zip archive %s: %w", zipPath, err)
	}
	return nil
}
