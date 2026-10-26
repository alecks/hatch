package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// writeFileAtomic writes a file atomically by writing to a temp file then renaming.
//
// Adapted from https://github.com/tailscale/tailscale/blob/main/atomicfile/atomicfile.go, Windows fixes removed.
// Copyright (c) Tailscale Inc & contributors.
// SPDX-License-Identifier: BSD-3-Clause.
func writeFileAtomic(filename string, data []byte, perm os.FileMode) (err error) {
	fi, err := os.Stat(filename)
	if err == nil && !fi.Mode().IsRegular() {
		return fmt.Errorf("%s already exists and is not a regular file", filename)
	}

	f, err := os.CreateTemp(filepath.Dir(filename), filepath.Base(filename)+".tmp")
	if err != nil {
		return err
	}

	tmpName := f.Name()
	defer func() {
		if err != nil {
			f.Close()
			os.Remove(tmpName)
		}
	}()

	if _, err := f.Write(data); err != nil {
		return err
	}
	if err := f.Chmod(perm); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	// NOTE: using os.Rename here rather than tailscale's custom rename for Windows perms
	return os.Rename(tmpName, filename)
}
