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
	"os"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"

	"github.boschdevcloud.com/fsil/fsil.go/command"
	"github.boschdevcloud.com/fsil/fsil.go/command/log"
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
	slog.SetDefault(log.NewLogger(c.logLevel))

	fmt.Fprintf(flag.CommandLine.Output(), "Reading file: %s\n", c.inputFile)
	c.loadYamlAST(c.inputFile)
	fmt.Fprintf(flag.CommandLine.Output(), "Updating file: %s\n", c.inputFile)
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
		fmt.Println("Error creating directories:", err)
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

func (c *ResolveCommand) loadYamlAST(file string) error {
	usesMap := make(map[string]interface{})

	data, err := os.ReadFile(file)
	if err != nil {
		fmt.Println("Error reading YAML file:", err)
		return fmt.Errorf("Error reading YAML AST file: %v", err)
	}

	if err := yaml.Unmarshal(data, &c.yamlAst); err != nil {
		fmt.Println("Error parsing YAML:", err)
		return fmt.Errorf("Error parsing YAML file: %v", err)
	}

	//used for E2E tests, eg: bin/ast resolve -input ast.yml -uses dse.fmi -file md_dse.fmi.yml
	if c.repoName != "" && c.metadataFile != "" {
		data, err := os.ReadFile(c.metadataFile)
		if err != nil {
			fmt.Println("Error reading Metadata YAML file:", err)
			return fmt.Errorf("Error reading Metadata YAML AST file: %v", err)
		}

		if err := yaml.Unmarshal(data, &c.yamlMetadata); err != nil {
			fmt.Println("Error parsing Metadata YAML:", err)
			return fmt.Errorf("Error parsing Metadata YAML file: %v", err)
		}
		usesMap[c.repoName] = c.yamlMetadata
		updateModelMd(usesMap, c, file)
		updateUsesMd(usesMap, c, file)
	} else {
		if spec, ok := c.yamlAst["spec"].(map[string]interface{}); ok {
			if uses, ok := spec["uses"].([]interface{}); ok {
				for _, use := range uses {
					if useMap, ok := use.(map[string]interface{}); ok {
						var rawUrl = genGitRawURL(useMap)
						if rawUrl != "" {
							var sha = calculateSha256(rawUrl)
							var cacheFilepath = appendFileName(c.cacheDir, sha)
							var yamlData = make(map[string]interface{})
							if c.cacheDir != "" {
								if !dirExists(c.cacheDir) {
									createCacheDir(c.cacheDir)
								}
								if FileExists(cacheFilepath) {
									data, err := os.ReadFile(cacheFilepath)
									if err != nil {
										fmt.Println("Error reading chache file:", err)
										return fmt.Errorf("Error reading chache file: %v", err)
									}
									if err := yaml.Unmarshal(data, &yamlData); err != nil {
										fmt.Println("Error parsing chache YAML:", err)
										return fmt.Errorf("Error parsing cache YAML file: %v", err)
									}
								} else {
									yamlData = fetchMetadata(rawUrl)
									saveCacheFile(cacheFilepath, yamlData)
								}
							} else {
								yamlData = fetchMetadata(rawUrl)
							}
							usesMap[useMap["name"].(string)] = yamlData
						}
					}
				}
			}
		}
		updateModelMd(usesMap, c, file)
		updateUsesMd(usesMap, c, file)
	}
	return nil
}

func genGitRawURL(useMap map[string]interface{}) string {
	pattern := `https:\/\/github\.com\/(\w+)\/(\w+(?:\.\w+))(\/.*)?`
	re := regexp.MustCompile(pattern)

	gitLink, ok := useMap["url"].(string)
	if !ok {
		fmt.Println("Invalid repo link")
		return ""
	}

	matchResult := re.FindStringSubmatch(gitLink)
	owner, repoName, path := "", "", ""

	if len(matchResult) > 0 {
		owner = matchResult[1]
		repoName = matchResult[2]
		if len(matchResult) > 3 {
			path = matchResult[3]
		}
	}

	var rawURL string
	if path != "" {
		return ""
	} else {
		version, _ := useMap["version"].(string)
		rawURL = fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", owner, repoName, version, "Metadata.yml")
	}

	return rawURL
}

func fetchMetadata(url string) map[string]interface{} {
	var yamlData = map[string]interface{}{}
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching the URL:", err)
		return yamlData
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return yamlData
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading the YAML file:", err)
		return yamlData
	}

	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		fmt.Println("Error parsing YAML:", err)
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

func filterMetadataYaml(data map[string]interface{}, searchModelVal string) map[string]interface{} {
	metadata, ok := data["metadata"].(map[string]interface{})
	if !ok {
		return nil
	}

	models, ok := metadata["models"].(map[string]interface{})
	if !ok {
		return nil
	}

	var selectedModel map[string]interface{}
	var relatedWorkflows []interface{}
	// find the model node with the searchModelVal
	for key, model := range models {
		modelMap, ok := model.(map[string]interface{})
		if !ok {
			continue
		}
		if key == searchModelVal {
			selectedModel = modelMap
			relatedWorkflows, _ = modelMap["workflows"].([]interface{})
			metadata["models"] = map[string]interface{}{key: modelMap}
			break
		}
	}

	if selectedModel == nil {
		return nil // No matching model found
	}

	tasks, ok := metadata["tasks"].(map[string]interface{})
	if !ok {
		return data
	}

	filteredTasks := make(map[string]interface{})

	// Keep only the tasks that are associated with related workflows
	for _, workflow := range relatedWorkflows {
		if task, exists := tasks[workflow.(string)]; exists {
			if taskMap, ok := task.(map[string]interface{}); ok {
				_, exists := taskMap["generates"]
				if exists {
					delete(taskMap, "vars")
					filteredTasks[workflow.(string)] = taskMap
				} else {
					delete(filteredTasks, workflow.(string))
				}
			}
		}
	}
	metadata["tasks"] = filteredTasks
	return data
}

func updateUsesMd(usesMap map[string]interface{}, c *ResolveCommand, file string) {
	if spec, exists := c.yamlAst["spec"].(map[string]interface{}); exists {
		if uses, exists := spec["uses"].([]interface{}); exists {
			for i, item := range uses {
				if useMap, ok := item.(map[string]interface{}); ok {
					if name, nameExists := useMap["name"].(string); nameExists {
						value, exists := usesMap[name]
						if exists {
							if valueMap, ok := value.(map[string]interface{}); ok {
								if containerMap, ok := valueMap["metadata"].(map[string]interface{}); ok {
									if container, ok := containerMap["container"].(map[string]interface{}); ok {
										if value, exists := container["repository"]; exists {
											if _, ok := useMap["metadata"].(map[string]interface{}); !ok {
												useMap["metadata"] = make(map[string]interface{})
												useMap["metadata"].(map[string]interface{})["container"] = map[string]interface{}{
													"repository": value,
												}
												uses[i] = useMap
											}
										}
									}
								}
							}
						} else {
							useMap["metadata"] = make(map[string]interface{})
							uses[i] = useMap
						}
					}
				}
			}
		}
	}

	err := updateFile(c.yamlAst, file)
	if err != nil {
		fmt.Println("Error:", err)
	}
}

func updateModelMd(usesMap map[string]interface{}, c *ResolveCommand, file string) {
	if spec, ok := c.yamlAst["spec"].(map[string]interface{}); ok { //looping through the input yaml
		if stacks, ok := spec["stacks"].([]interface{}); ok {
			for _, stack := range stacks {
				if stackMap, ok := stack.(map[string]interface{}); ok {
					if models, ok := stackMap["models"].([]interface{}); ok {
						for _, model := range models {
							if modelMap, ok := model.(map[string]interface{}); ok {
								if model_name_to_search, ok := modelMap["model"].(string); ok {
									for key, usesItem := range usesMap { //looping through the cached uses items to find if 'model_name_to_search' is present in uses items model displayname
										if modelObj, ok := usesItem.(map[string]interface{}); ok {
											filteredMD := filterMetadataYaml(modelObj, model_name_to_search)
											if filteredMD != nil {
												metadata := filteredMD["metadata"].(map[string]interface{})
												//delete(metadata, "container")
												modelMap["metadata"] = metadata
												modelMap["uses"] = key
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	err := updateFile(c.yamlAst, file)
	if err != nil {
		fmt.Println("Error:", err)
	}
}
