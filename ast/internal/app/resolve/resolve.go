// Copyright 2025 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

package resolve

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/boschglobal/dse.clib/extra/go/command"
)

type ResolveCommand struct {
	command.Command

	inputFile    string
	logLevel     int
	repoName     string
	metadataFile string
	cacheDir     string

	yamlAst      map[string]interface{}
	yamlMetadata map[string]interface{}
}

var luaModels []string

func NewResolveCommand(name string) *ResolveCommand {
	c := &ResolveCommand{
		Command: command.Command{
			Name:    name,
			FlagSet: flag.NewFlagSet(name, flag.ExitOnError),
		},
	}
	c.FlagSet().StringVar(&c.inputFile, "input", "", "path to YAML AST file")
	c.FlagSet().IntVar(&c.logLevel, "log", 4, "Loglevel")
	c.FlagSet().StringVar(&c.repoName, "uses", "", "repository name (hidden)")
	c.FlagSet().StringVar(&c.metadataFile, "file", "", "path to metadata file")
	c.FlagSet().StringVar(&c.cacheDir, "cache", "out/cache", "cache directory")
	return c
}

func (c ResolveCommand) Name() string {
	return c.Command.Name
}

func (c ResolveCommand) FlagSet() *flag.FlagSet {
	return c.Command.FlagSet
}

func (c *ResolveCommand) Parse(args []string) error {
	return c.FlagSet().Parse(args)
}

func (c *ResolveCommand) Run() error {
	//slog.SetDefault(log.NewLogger(c.logLevel))
	c.yamlMetadata = make(map[string]interface{})

	slog.Info("Reading AST file", "file", c.inputFile)
	if err := c.loadYamlAST(); err != nil {
		return err
	}
	slog.Info("Load metadata files")
	if err := c.loadMetadata(); err != nil {
		return err
	}
	slog.Info("Updating AST file", "file", c.inputFile)
	if err := c.updateMetadata(); err != nil {
		return err
	}
	return nil
}

func calculateSha256(url string) string {
	hash := sha256.Sum256([]byte(url))
	hashString := hex.EncodeToString(hash[:])
	return hashString
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func FileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil && !info.IsDir()
}

func createCacheDir(path string) {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		slog.Error("Unable to create cache dir", "path", path, "err", err)
		return
	}
}

func saveCacheFile(filePath string, data map[string]interface{}) error {
	yamlData, err := yaml.Marshal(data)
	if len(yamlData) != 3 {
		if err != nil {
			return fmt.Errorf("failed to marshal YAML: %v", err)
		}

		if err := os.WriteFile(filePath, yamlData, 0644); err != nil {
			return fmt.Errorf("failed to write cache file: %v", err)
		}
	}
	return nil
}

func appendFileName(path, filename string) string {
	return filepath.Join(path, filename)
}

func (c *ResolveCommand) loadYamlAST() error {
	data, err := os.ReadFile(c.inputFile)
	if err != nil {
		return fmt.Errorf("Error reading YAML AST file: %v", err)
	}
	if err := yaml.Unmarshal(data, &c.yamlAst); err != nil {
		return fmt.Errorf("Error parsing YAML file: %v", err)
	}
	return nil
}

func getYamlPath(root interface{}, keys ...string) interface{} {
	node := root
	var exists bool
	for _, key := range keys {
		if node, exists = node.(map[string]interface{})[key]; !exists {
			return nil
		}
	}
	return node
}

func isLuaReference(useUrl string, useMap map[string]interface{}) bool {
	var pathPtr *string
	if p, ok := useMap["path"].(string); ok {
		pathPtr = &p
	}
	// inner path (e.g. inside zip)
	if pathPtr != nil {
		if strings.EqualFold(filepath.Ext(*pathPtr), ".lua") {
			return true
		}
	}
	// direct file URL
	return strings.EqualFold(filepath.Ext(useUrl), ".lua")
}

func parseFilePath(useUrl string, useMap map[string]interface{}) (bool, string) {
	u, err := url.Parse(useUrl)
	if err != nil {
		return false, ""
	}

	if u.Scheme == "file" {
		return true, u.Path
	}

	return false, ""
}

func resolvePath(pathOrURL string) (abs string, isDir bool, err error) {
	// Handle file:// URLs
	if strings.HasPrefix(pathOrURL, "file://") {
		u, err := url.Parse(pathOrURL)
		if err != nil {
			return "", false, err
		}
		pathOrURL = u.Path
	}

	// Resolve to absolute path
	abs, err = filepath.Abs(pathOrURL)
	if err != nil {
		return "", false, err
	}

	info, err := os.Stat(abs)
	if err != nil {
		return "", false, err
	}

	isDir = info.IsDir() // if info.IsDir() is true then the uses item is pointing to local repo folder(Taskfile.yaml is loaded from that path). eg:dse.sdp file:///mnt/c/Users/dse.sdp

	return abs, isDir, nil
}

func loadTaskfile(filePath string) (map[string]interface{}, error) {
	var data []byte
	taskfiles := []string{"Taskfile.yml", "Taskfile.yaml"} // for supporting '.yml' or '.yaml' Taskfile extension
	for _, name := range taskfiles {
		abs, _, err := resolvePath(filepath.Join(filePath, name))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}

		data, err = os.ReadFile(abs)
		if err == nil {
			break
		}

		if !os.IsNotExist(err) {
			slog.Error("failed to read taskfile", "path", abs, "error", err)
			return nil, err
		}
	}

	if data == nil {
		slog.Error("taskfile not found", "path", filePath)
		return nil, fmt.Errorf("no Taskfile.yml or Taskfile.yaml found in %s", filePath)
	}

	yamlData := map[string]interface{}{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return nil, err
	}

	return yamlData, nil
}

func (c *ResolveCommand) loadMetadata() error {
	if c.repoName != "" && c.metadataFile != "" {
		// Supports E2E tests.
		// eg: bin/ast resolve -input ast.yml -uses dse.fmi -file md_dse.fmi.yml
		slog.Info("Load metadata from local file", "repo", c.repoName, "file", c.metadataFile)
		data, err := os.ReadFile(c.metadataFile)
		if err != nil {
			return fmt.Errorf("Error reading Metadata YAML AST file: %v", err)
		}
		var yamlData = map[string]interface{}{}
		if err := yaml.Unmarshal(data, &yamlData); err != nil {
			return fmt.Errorf("Error parsing Metadata YAML file: %v", err)
		}
		c.yamlMetadata[c.repoName] = yamlData
		return nil
	}

	uses := getYamlPath(c.yamlAst, "spec", "uses")
	if uses == nil {
		slog.Error("Path spec/users not found in AST file")
		return nil
	}
	for _, _use := range uses.([]interface{}) {
		use := _use.(map[string]interface{})
		// Fetch metadata.
		var yamlData = map[string]interface{}{}
		slog.Debug("Fetch metadata for uses", "name", use["name"].(string))
		useUrl, ok := use["url"].(string)
		if !ok {
			slog.Error("Invalid or missing URL in uses map")
			return nil
		}
		ok_fileRef, path := parseFilePath(useUrl, use)
		isLua := isLuaReference(useUrl, use)
		if ok_fileRef && !isLua {
			// file urls can be either pointing to local repo folder having taskfile.yaml or another static file in repo
			// eg uses block, dse.sdp file:///mnt/c/Users/NUZ2KOR/Desktop/dse.sdp/
			//				  input file:///mnt/c/Users/files.zip path=data/filename.txt
			abs, isDir, err := resolvePath(path)
			if err != nil || isDir == false {
				return nil
			}
			slog.Info("Loading Taskfile from local path", "path", abs)
			yamlData, err = loadTaskfile(abs)

		} else {
			// eg uses block, dse.sdp https://github.com/boschglobal/dse.sdp v0.8.26
			rawUrls := genGitRawURL(use)
			if len(rawUrls) == 0 {
				continue
			}
			var loaded bool
			for _, rawUrl := range rawUrls {
				slog.Info("Metadata download", "url", rawUrl)
				// Search the cache.
				if c.cacheDir != "" {
					var cacheFilepath = appendFileName(c.cacheDir, calculateSha256(rawUrl))
					if !dirExists(c.cacheDir) {
						createCacheDir(c.cacheDir)
					}

					if FileExists(cacheFilepath) {
						data, err := os.ReadFile(cacheFilepath)
						if err != nil {
							continue
						}
						if err := yaml.Unmarshal(data, &yamlData); err == nil {
							loaded = true
							break
						}
					} else {
						yamlData = fetchMetadata(rawUrl, use)
						if len(yamlData) != 0 {
							saveCacheFile(cacheFilepath, yamlData)
							loaded = true
							break
						}
					}
				} else {
					yamlData = fetchMetadata(rawUrl, use)
					if len(yamlData) != 0 {
						loaded = true
						break
					}
				}
			}

			if !loaded {
				slog.Warn(
					"404 Not Found: Taskfile.yml / Taskfile.yaml not found",
					"use", use["name"],
					"url", use["url"],
				)
				continue
			}
		}

		slog.Info("Update metadata for repo", "name", use["name"].(string))
		c.yamlMetadata[use["name"].(string)] = yamlData
	}
	return nil
}

func (c *ResolveCommand) updateMetadata() error {
	c.updateAstUsesMetadata()
	c.updateAstModelMetadata()
	err := updateFile(c.yamlAst, c.inputFile)
	if err != nil {
		return err
	}
	return nil
}

func genGitRawURL(useMap map[string]interface{}) []string {
	useUrl, ok := useMap["url"].(string)
	if !ok {
		slog.Error("Invalid or missing URL in uses map")
		return nil
	}
	version, _ := useMap["version"].(string)

	// Encode the URL, especially for https://{{.GHE_TOKEN}}@github ....
	u, _ := func() (*url.URL, error) {
		_u := useUrl
		_u = strings.ReplaceAll(_u, `{`, `%7B`)
		_u = strings.ReplaceAll(_u, `}`, `%7D`)
		return url.Parse(_u)
	}()
	if strings.HasPrefix(u.Host, "github.") == false {
		slog.Debug("Unsupported metadata url", "url", useUrl)
		return nil
	}
	pathParts := strings.Split(u.Path, string(os.PathSeparator))
	if len(pathParts) > 3 {
		// Not a repo path, more likely an asset link (for download).
		return nil
	}
	useUrl = func() string {
		var finalUrl *url.URL
		switch u.Host {
		case "github.com":
			u.Host = "raw.githubusercontent.com"
			finalUrl, _ = u.Parse(fmt.Sprintf("/%s/%s/refs/tags/%s/Taskfile.yml", pathParts[1], pathParts[2], version))
		case "github.boschdevcloud.com":
			u.Host = "raw.github.boschdevcloud.com"
			finalUrl, _ = u.Parse(fmt.Sprintf("/%s/%s/%s/Taskfile.yml", pathParts[1], pathParts[2], version))
		default:
			slog.Error("Unsupported URL hostname")
		}
		return func() string {
			// Remove encoding.
			_u := finalUrl.String()
			_u = strings.ReplaceAll(_u, `%7B`, `{`)
			_u = strings.ReplaceAll(_u, `%7D`, `}`)
			return _u
		}()
	}()

	return []string{
		strings.Replace(useUrl, "Taskfile.yml", "Taskfile.yml", 1),
		strings.Replace(useUrl, "Taskfile.yml", "Taskfile.yaml", 1),
	}
}

func fetchMetadata(url string, use map[string]interface{}) map[string]interface{} {
	var yamlData = map[string]interface{}{}
	url = strings.ReplaceAll(url, `{{.GHE_TOKEN}}`, os.Getenv("GHE_TOKEN"))
	resp, err := http.Get(url)
	if err != nil {
		slog.Error("Error fetching the URL", "err", err)
		return yamlData
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return yamlData
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("Bad return code", "code", resp.StatusCode, "url", url)
		return yamlData
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Error reading the YAML file", "err", err)
		return yamlData
	}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		slog.Error("Error parsing YAML", "err", err)
		return yamlData
	}
	return yamlData
}

func updateFile(data interface{}, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()
	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	defer encoder.Close()
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("error encoding YAML: %v", err)
	}
	return nil
}

func (c *ResolveCommand) updateAstUsesMetadata() {
	// AST : spec/uses/use[name=repo]/metadata <= repo metadata from Taskfile.
	uses := getYamlPath(c.yamlAst, "spec", "uses")
	if uses == nil {
		slog.Error("Path spec/users not found in AST file")
		return
	}
	for _, _use := range uses.([]interface{}) {
		use := _use.(map[string]interface{})
		urlStr, urlOk := use["url"].(string)
		// Locate the metadata.
		re := regexp.MustCompile(`\{\{.*?\}\}@`) // for resolving Parse issue with urls like https://{{.GHE_TOKEN}}@github...
		parsedUrl, err := url.Parse(re.ReplaceAllString(urlStr, ""))
		if isLuaReference(urlStr, use) {
			luaModels = append(luaModels, use["name"].(string))
			continue
		}

		if parsedUrl.Scheme == "file" {
			// file urls can be either pointing to local repo folder having taskfile.yaml or another static file in repo
			if raw, ok := use["path"]; ok && raw != nil {
				if path, ok := raw.(string); ok && path != "" {
					_, isDir, err := resolvePath(path)
					if err != nil {
						continue
					}
					if !isDir {
						continue
					}
				}
			}
		}
		metadata := getYamlPath(c.yamlMetadata, use["name"].(string), "metadata")
		if parsedUrl.Scheme != "file" {
			versionStr, versionOk := use["version"].(string)
			if !urlOk || !versionOk || !strings.HasPrefix(versionStr, "v") {
				continue
			}

			if err != nil || strings.HasPrefix(parsedUrl.Host, "github.") == false {
				continue
			}

			slog.Info("Uses item", "name", use["name"].(string))
			if metadata == nil {
				if strings.Contains(use["url"].(string), "blob") || strings.Contains(use["url"].(string), "releases") {
					continue
				}
				taskfileURL := fmt.Sprintf("%s/blob/%s/Taskfile.yml", use["url"], use["version"].(string))
				slog.Error("Repo does not have associated metadata", "name", use["name"].(string))
				slog.Info(fmt.Sprintf("Include Metadata in %s to resolve the issue", taskfileURL))
				os.Exit(1)
			}
		}

		// Update (merge) to the underlying map/slice.
		slog.Info("Merge metadata to spec/uses[]", "name", use["name"].(string))
		if _, ok := use["metadata"]; !ok {
			use["metadata"] = map[string]interface{}{}
		}
		mergeKeys := []string{"container", "package", "models"}
		metaMap, ok := metadata.(map[string]interface{})
		if !ok || metaMap == nil {
			continue
		}
		for k, v := range metaMap {
			if slices.Contains(mergeKeys, k) {
				use["metadata"].(map[string]interface{})[k] = v
			}
		}
	}
}

func (c *ResolveCommand) updateAstModelMetadata() {
	// AST : spec/uses/use[*]/metadata/models[name]/model <= model metadata from Taskfile.
	stacks := getYamlPath(c.yamlAst, "spec", "stacks")
	if stacks == nil {
		slog.Error("Path spec/stacks not found in AST file")
		return
	}
	for _, _stack := range stacks.([]interface{}) {
		stack := _stack.(map[string]interface{})
		models := getYamlPath(stack, "models")
		for _, _model := range models.([]interface{}) {
			model := _model.(map[string]interface{})
			if _, ok := model["metadata"]; !ok {
				model["metadata"] = map[string]interface{}{}
			}
			slog.Info("Updating model metadata", "model", model["model"].(string), "name", model["name"].(string))

			// Locate the related Repo Metadata (for this model).
			repos := getYamlPath(c.yamlMetadata)
			if repos == nil {
				slog.Info("Repos metadata not present")
				continue
			}
			if slices.Contains(luaModels, model["model"].(string)) { // if the model is Lua the uses should be assigned with the Model name value
				model["uses"] = model["model"]
				continue
			}

			for repoName, _repo := range repos.(map[string]interface{}) {
				repo := _repo.(map[string]interface{})
				models := getYamlPath(repo, "metadata", "models")
				if models == nil {
					slog.Debug("Repo does not have metadata", "repoName", repoName)
					continue
				}
				for modelName, _ := range models.(map[string]interface{}) {
					if modelName == model["model"].(string) {
						slog.Info("Repo metadata located", "repo", repoName, "model", modelName)

						// Merge in the repo metadata.
						repoMetadata := getYamlPath(repo, "metadata").(map[string]interface{})
						// [repo]/metadata/package => [model]/metadata/package
						if v := getYamlPath(repoMetadata, "package"); v != nil {
							model["metadata"].(map[string]interface{})["package"] = v
						}
						// [repo]/metadata/container => [model]/metadata/container
						if v := getYamlPath(repoMetadata, "container"); v != nil {
							model["metadata"].(map[string]interface{})["container"] = v

						}
						// [repo]/metadata/models/[model] => [model]/metadata/models/[model]
						if v := getYamlPath(repoMetadata, "models", model["model"].(string)); v != nil {
							model["metadata"].(map[string]interface{})["models"] = map[string]interface{}{}
							model["metadata"].(map[string]interface{})["models"].(map[string]interface{})[model["model"].(string)] = v
						}

						// Locate and merge in the workflow metadata.
						model["metadata"].(map[string]interface{})["tasks"] = map[string]interface{}{}
						tasks := getYamlPath(repo, "tasks")
						if tasks == nil {
							slog.Debug("Repo does not have tasks", "repoName", repoName)
							continue
						}
						for taskName, _task := range tasks.(map[string]interface{}) {
							task := _task.(map[string]interface{})
							if v := getYamlPath(task, "metadata", "generates"); v != nil {
								slog.Info("Task metadata located", "repo", repoName, "taskName", taskName)
								g := map[string]interface{}{}
								g["generates"] = v
								model["metadata"].(map[string]interface{})["tasks"].(map[string]interface{})[taskName] = g
							}
						}

						slog.Info("Updating model uses", "model", model["model"].(string), "name", model["name"].(string))
						model["uses"] = repoName

						//vars from repo Taskfile - "TAG", "PACKAGE_VERSION"
						//Collect existing model vars defined in dse (map or []{name,value})
						existingVars := map[string]interface{}{}
						if mv, ok := model["vars"]; ok {
							switch v := mv.(type) {
							case map[string]interface{}:
								for k, val := range v {
									existingVars[k] = val
								}
							case []interface{}:
								for _, item := range v {
									if m, ok := item.(map[string]interface{}); ok {
										if name, ok := m["name"].(string); ok {
											existingVars[name] = m["value"]
										}
									}
								}
							}
						}

						//Merge ONLY missing TAG & PACKAGE_VERSION from repo vars
						if repoVars := getYamlPath(repo, "vars"); repoVars != nil {
							if varsMap, ok := repoVars.(map[string]interface{}); ok {

								for _, k := range []string{"TAG", "PACKAGE_VERSION"} {
									// Do not override if already present
									if _, exists := existingVars[k]; exists {
										continue
									}

									v, ok := varsMap[k]
									if !ok {
										continue
									}

									// Resolve simple self-reference: {{.PACKAGE_VERSION}}
									if s, ok := v.(string); ok {
										re := regexp.MustCompile(`^\{\{\.(\w+)\}\}$`)
										if m := re.FindStringSubmatch(s); len(m) == 2 {
											if refVal, ok := varsMap[m[1]]; ok {
												v = refVal
											}
										}
									}

									existingVars[k] = v
								}
							}
						}

						// Step 3: Normalize ALL vars back to []{name,value}
						var out []interface{}
						for k, v := range existingVars {
							out = append(out, map[string]interface{}{
								"name":  k,
								"value": v,
							})
						}
						model["vars"] = out
					}
				}

			}
		}
	}
}
