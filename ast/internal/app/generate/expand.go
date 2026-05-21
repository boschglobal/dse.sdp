// Copyright 2025 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

package generate

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/tools/txtar"
)

func ExpandTxtar(archivePath string, outputDir string, overwrite bool) error {
	a, err := txtar.ParseFile(archivePath)
	if err != nil {
		return err
	}

	for _, f := range a.Files {
		relPath := f.Name
		if path.IsAbs(relPath) || filepath.IsAbs(relPath) {
			continue
		}
		cleanRel := filepath.Clean(filepath.FromSlash(relPath))
		if strings.HasPrefix(cleanRel, "..") {
			continue
		}
		if strings.HasSuffix(relPath, ".dse") {
			continue
		}
		absPath := filepath.Join(outputDir, cleanRel)
		if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
			return err
		}
		if !overwrite {
			if _, err := os.Stat(absPath); err == nil {
				continue
			}
		}
		if err := os.WriteFile(absPath, f.Data, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", absPath, err)
		}
	}
	return nil
}
