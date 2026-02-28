// Copyright 2025 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/boschglobal/dse.schemas/code/go/dse/ast"
)

func metadataQuery(keys ...string) string {
	return "metadata/" + strings.Join(keys, "/")
}

func logMissingMetadata(required bool, query, uri, uses string, defaultValue interface{}) {
	if required {
		fmt.Fprintf(os.Stderr, "[Error] Item not found! query=%s, uri=%s, uses=%s\n", query, uri, uses)
		return
	}
	fmt.Fprintf(os.Stdout, "[Info] Item not found, using default (%v). query=%s, uri=%s, uses=%s\n", defaultValue, query, uri, uses)
}

func lookupMetadataValue(md map[string]interface{}, keys ...string) (interface{}, bool) {
	var node interface{} = md
	for _, key := range keys {
		m, ok := node.(map[string]interface{})
		if !ok {
			return nil, false
		}
		v, ok := m[key]
		if !ok {
			return nil, false
		}
		node = v
	}
	return node, true
}

func requiredMetadataString(md map[string]interface{}, uses ast.Uses, keys ...string) (string, error) {
	query := metadataQuery(keys...)
	uri := taskfileURIForUses(uses)
	value, ok := lookupMetadataValue(md, keys...)
	if !ok {
		logMissingMetadata(true, query, uri, uses.Name, "")
		return "", fmt.Errorf("missing metadata: %s", query)
	}
	s, ok := value.(string)
	if !ok || s == "" {
		logMissingMetadata(true, query, uri, uses.Name, "")
		return "", fmt.Errorf("missing metadata: %s", query)
	}
	return s, nil
}

func optionalMetadataBool(md map[string]interface{}, uses ast.Uses, defaultValue bool, keys ...string) bool {
	query := metadataQuery(keys...)
	uri := taskfileURIForUses(uses)
	value, ok := lookupMetadataValue(md, keys...)
	if !ok {
		logMissingMetadata(false, query, uri, uses.Name, defaultValue)
		return defaultValue
	}
	b, ok := value.(bool)
	if !ok {
		logMissingMetadata(false, query, uri, uses.Name, defaultValue)
		return defaultValue
	}
	return b
}

func optionalMetadataSlice(md map[string]interface{}, uses ast.Uses, defaultValue []interface{}, keys ...string) []interface{} {
	query := metadataQuery(keys...)
	uri := taskfileURIForUses(uses)
	value, ok := lookupMetadataValue(md, keys...)
	if !ok {
		logMissingMetadata(false, query, uri, uses.Name, defaultValue)
		return defaultValue
	}
	slice, ok := value.([]interface{})
	if !ok {
		logMissingMetadata(false, query, uri, uses.Name, defaultValue)
		return defaultValue
	}
	return slice
}

func taskfileURIForUses(uses ast.Uses) string {
	if uses.Url == "" {
		return ""
	}
	u, err := urlEscapedParse(uses.Url)
	if err != nil {
		return uses.Url
	}
	switch u.Scheme {
	case "file":
		path := u.Path
		if path == "" {
			return uses.Url
		}
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			for _, name := range []string{"Taskfile.yml", "Taskfile.yaml"} {
				candidate := filepath.Join(path, name)
				if _, err := os.Stat(candidate); err == nil {
					return "file://" + candidate
				}
			}
			return "file://" + filepath.Join(path, "Taskfile.yml")
		}
		if strings.HasSuffix(path, "Taskfile.yml") || strings.HasSuffix(path, "Taskfile.yaml") {
			return "file://" + path
		}
		return uses.Url
	case "https":
		if uses.Version == nil {
			return uses.Url
		}
		if !strings.HasPrefix(u.Host, "github.") {
			return uses.Url
		}
		pathParts := strings.Split(u.Path, string(os.PathSeparator))
		if len(pathParts) < 3 {
			return uses.Url
		}
		owner := pathParts[1]
		repo := pathParts[2]
		switch u.Host {
		case "github.com":
			return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/refs/tags/%s/Taskfile.yml", owner, repo, *uses.Version)
		case "github.boschdevcloud.com":
			return fmt.Sprintf("https://raw.github.boschdevcloud.com/%s/%s/%s/Taskfile.yml", owner, repo, *uses.Version)
		default:
			return uses.Url
		}
	default:
		return uses.Url
	}
}
