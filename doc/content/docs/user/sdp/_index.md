---
title: "SDP - Simulation Development Platform"
linkTitle: "SDP"
weight: 40
tags:
- SDP
github_repo: "https://github.com/boschglobal/dse.sdp"
github_subdir: "doc"
---


## Synopsis
Simulation Development Platform (SDP) using Codespaces or DevContainers to code, build, and run DSE Simulations.


## Setup
### Codespacec

GitHub Codespaces provides a cloud-hosted development environment pre-configured for the SDP project.

#### Steps to use Codespaces:

1. Go to the repository on GitHub.
2. Click the <b>Code</b> button and choose <b>Open with Codespaces</b>.
3. If you don’t have an existing codespace, click <b>New codespace</b> to create one.

The workspace will:
- Automatically build the dev container.
- Install all required dependencies.
- Launch VS Code in the browser (or desktop, if configured).

### Dev Container

If you prefer a local setup with the same environment used in Codespaces, you can follow the below approach.

#### Steps to use the Dev Container locally:

1. Clone the repository:
   ```bash
   $ git clone https://github.com/boschglobal/dse.sdp.git
   $ cd dse.sdp
   $ make
   $ make install
   ```

2. Install the extension

	Install the DSE Language Extension in VS Code:

	After building the extension, a `dse.vsix` file will be generated in the `lsp/out/bin` folder.

	Using VS Code GUI:

	1. Open the Extensions view by pressing `Ctrl+Shift+X`.
	2. Click the `...` (More Actions) menu in the top-right corner of the Extensions panel.
	3. Select `Install from VSIX...`.
	4. Navigate to the `lsp/out/bin` folder and select the generated `dse.vsix` file.
	
	Using PowerShell:

	```powershell
	cd dse.sdp
	code --install-extension lsp\out\bin\dse.vsix
	```

## Usage

### VS Code

#### VS Code DSE Commands

The following commands are available via the Command Palette (`Ctrl+Shift+P`) when the SDP extension is installed:

| Command | Description |
|---------|-------------|
| Build (`DSE: Build`) | Generates `simulation.yaml` and `Taskfile.yml` from the active `.dse` file. This prepares the simulation environment. |
| Check (`DSE: Check`) | Analyzes the simulation graph and produces a report to help visualize and verify the structure of the simulation. |
| Run (`DSE: Run`) | Executes the simulation using the currently configured simulation definition. |
| Clean (`DSE: Clean`) | Performs a clean operation using `task: clean`, removing generated artifacts and build files. |
| Cleanall (`DSE: Cleanall`) | Performs a deep clean via `task: cleanall`, removing all outputs and intermediate data. |


#### Live AST View

The extension supports live viewing of the models and channels derived from `.dse` files

##### To view the AST preview

1. Open a supported `.dse` file in the VS Code editor.
2. Click the `Open Preview` button in the upper-right corner of the editor window.

###### Alternatively, you can use keyboard shortcuts

- Press **`Ctrl + K V`** to open preview in a side panel.
- Press **`Ctrl + Shift + V`** to open preview in the main panel.


#### Important: One-Time Setup for Local Environments

If you're working locally in VS Code (not in GitHub Codespaces), you may need to manually install required global dependencies the first time:

```powershell
cd dse.sdp/dsl
npm install -g --force
```
This will ensure that all necessary global dependencies are installed properly.


### Terminal

#### Build

The following steps outline build process using the `dse` CLI tools:

##### 1. Parse the Input File

Convert a DSE file into an intermediate JSON representation
```bash
$ dse-parse2ast <dse_file_path> <json_output_file_path>
```
##### 2. Convert JSON to YAML AST

Transform the parsed JSON into a YAML-based Abstract Syntax Tree (AST)
```bash
$ dse-ast convert -input <json_file_path> -output <yaml_ast_output_path>
```
##### 3. Resolve AST References

Resolve internal references within the AST to produce a fully linked version
```bash
$ dse-ast resolve -input <yaml_ast_path> -output <yaml_ast_output_path>
```
##### 4. Generate Output simulation files

Generate the final output simulation files based on the resolved AST
```bash
$ dse-ast generate -input <yaml_ast_path> -output <output_path>
```

> __Note:__ Replace the placeholders like <dse_file_path> and <output_path> with actual file paths specific to your project.

### GitHub Workflows

## DSL

### Keywords

**simulation**  
Defines the simulation setup including architecture, stepsize, and endtime.

**arch**  
Specifies the target architecture (e.g., `linux-amd64`, `linux-x86`).

**stepsize**  
Time increment for each simulation step.

**endtime**  
Total simulation duration.

**channel**  
Declares a communication channel — represents a grouping of signals which are exchanged between models.

**network**  
Defines a network interface with a detailed descriptor (e.g., for CAN).

**uses**  
Imports external dependencies such as modules, FMUs, or files.

**var**  
Declares variables, which may refer to other resources or contain static values.

**model**  
Defines a component in the simulation, such as an FMU or a gateway.

**uid**  
Assigns a unique ID to a model or component.

**envar**  
Declares an environment variable used at model or stack scope.

**workflow**  
Defines a processing or generation step applied to a model or stack.

**stack**  
Declares a group of models composed together for simulation.

**stacked**  
A boolean flag that indicates if the models in a stack should be layered.

**sequential**  
A boolean flag for stacks that ensures models are executed one after another, in a defined order.

**file**  
Maps or includes external input/configuration files in the simulation.

### Example DSL File

<details>
<summary>openloop.dse</summary>

```dse
simulation arch=linux-amd64
channel physical

uses
dse.modelc https://github.com/boschglobal/dse.modelc v2.1.23
dse.fmi https://github.com/boschglobal/dse.fmi v1.1.23
linear_fmu https://github.com/boschglobal/dse.fmi/releases/download/v1.1.23/Fmi-1.1.23-linux-amd64.zip path=examples/fmu/linear/fmi2/linear.fmu

model input dse.modelc.csv
channel physical signal_channel
envar CSV_FILE model/input/data/input.csv
file input.csv input/openloop.csv
file signalgroup.yaml input/signalgroup.yaml

model linear dse.fmi.mcl
channel physical signal_channel
envar MEASUREMENT_FILE /sim/measurement.mf4
workflow generate-fmimcl
var FMU_DIR uses linear_fmu
var OUT_DIR {{.PATH}}/data
var MCL_PATH {{.PATH}}/lib/libfmimcl.so
```
</details>

### AST

The Abstract Syntax Tree (AST) is a structured representation of the DSL input. It captures the full semantic meaning of a simulation configuration, breaking down the components, models, channels, workflows, and stacks into a hierarchical data structure.

The AST is generated after parsing the DSL and serves as the intermediate layer between the textual DSL input and the generated output artifacts (such as `simulation.yaml`, `Taskfile.yml`). It enables validation, transformation, and automation workflows.

#### Generate AST

###### 1. Convert a DSE file into an intermediate JSON representation

```bash
$ dse-parse2ast openloop.dse openloop.ast.json
```

<details>
<summary>openloop.ast.json</summary>

```json
{
"type": "Simulation",
"simulation": "simulation arch=linux-amd64",
"object": {
	"image": "simulation arch=linux-amd64",
	"startOffset": 0,
	"endOffset": 26,
	"startLine": 0,
	"endLine": 1,
	"startColumn": 0,
	"endColumn": 27,
	"tokenTypeIdx": 3,
	"payload": {
	"simulation_arch": {
		"value": "linux-amd64",
		"token_type": "simulation_arch",
		"start_offset": 11,
		"end_offset": 28
	},
	"stepsize": {
		"value": "0.0005",
		"token_type": "stepsize",
		"start_offset": 28,
		"end_offset": 34
	},
	"endtime": {
		"value": "0.005",
		"token_type": "endtime",
		"start_offset": 34,
		"end_offset": 39
	}
	}
},
"children": {
	"channels": [
	{
		"type": "Channel",
		"object": {
		"image": "channel physical",
		"startOffset": 0,
		"endOffset": 15,
		"startLine": 1,
		"endLine": 1,
		"startColumn": 0,
		"endColumn": 16,
		"tokenTypeIdx": 4,
		"payload": {
			"channel_name": {
			"value": "physical",
			"token_type": "channel_name",
			"start_offset": 8,
			"end_offset": 17
			},
			"channel_alias": {
			"value": "",
			"token_type": "channel_alias",
			"start_offset": null,
			"end_offset": null
			}
		}
		},
		"children": {
		"networks": [
			
		]
		}
	}
	],
	"uses": [
	{
		"type": "Uses",
		"object": {
		"image": "dse.modelc https://github.com/boschglobal/dse.modelc v2.1.23",
		"startOffset": 0,
		"endOffset": 59,
		"startLine": 4,
		"endLine": 1,
		"startColumn": 0,
		"endColumn": 60,
		"tokenTypeIdx": 8,
		"payload": {
			"use_item": {
			"value": "dse.modelc",
			"token_type": "use_item",
			"start_offset": 1,
			"end_offset": 11
			},
			"link": {
			"value": "https://github.com/boschglobal/dse.modelc",
			"token_type": "link",
			"start_offset": 11,
			"end_offset": 53
			},
			"version": {
			"value": "v2.1.23",
			"token_type": "version",
			"start_offset": 53,
			"end_offset": 61
			},
			"path": {
			"value": "",
			"token_type": "path",
			"start_offset": null,
			"end_offset": null
			},
			"user": {
			"value": "",
			"token_type": "user",
			"start_offset": null,
			"end_offset": null
			},
			"token": {
			"value": "",
			"token_type": "token",
			"start_offset": null,
			"end_offset": null
			}
		}
		}
	},
	{
		"type": "Uses",
		"object": {
		"image": "dse.fmi https://github.com/boschglobal/dse.fmi v1.1.23",
		"startOffset": 0,
		"endOffset": 53,
		"startLine": 5,
		"endLine": 1,
		"startColumn": 0,
		"endColumn": 54,
		"tokenTypeIdx": 8,
		"payload": {
			"use_item": {
			"value": "dse.fmi",
			"token_type": "use_item",
			"start_offset": 1,
			"end_offset": 8
			},
			"link": {
			"value": "https://github.com/boschglobal/dse.fmi",
			"token_type": "link",
			"start_offset": 8,
			"end_offset": 47
			},
			"version": {
			"value": "v1.1.23",
			"token_type": "version",
			"start_offset": 47,
			"end_offset": 55
			},
			"path": {
			"value": "",
			"token_type": "path",
			"start_offset": null,
			"end_offset": null
			},
			"user": {
			"value": "",
			"token_type": "user",
			"start_offset": null,
			"end_offset": null
			},
			"token": {
			"value": "",
			"token_type": "token",
			"start_offset": null,
			"end_offset": null
			}
		}
		}
	},
	{
		"type": "Uses",
		"object": {
		"image": "linear_fmu https://github.com/boschglobal/dse.fmi/releases/download/v1.1.23/Fmi-1.1.23-linux-amd64.zip path=examples/fmu/linear/fmi2/linear.fmu",
		"startOffset": 0,
		"endOffset": 142,
		"startLine": 6,
		"endLine": 1,
		"startColumn": 0,
		"endColumn": 143,
		"tokenTypeIdx": 8,
		"payload": {
			"use_item": {
			"value": "linear_fmu",
			"token_type": "use_item",
			"start_offset": 1,
			"end_offset": 11
			},
			"link": {
			"value": "https://github.com/boschglobal/dse.fmi/releases/download/v1.1.23/Fmi-1.1.23-linux-amd64.zip",
			"token_type": "link",
			"start_offset": 11,
			"end_offset": 103
			},
			"version": {
			"value": "",
			"token_type": "version",
			"start_offset": 103,
			"end_offset": 103
			},
			"path": {
			"value": "examples/fmu/linear/fmi2/linear.fmu",
			"token_type": "path",
			"start_offset": 103,
			"end_offset": 143
			},
			"user": {
			"value": "",
			"token_type": "user",
			"start_offset": null,
			"end_offset": null
			},
			"token": {
			"value": "",
			"token_type": "token",
			"start_offset": null,
			"end_offset": null
			}
		}
		}
	}
	],
	"vars": [
	
	],
	"stacks": [
	{
		"type": "Stack",
		"name": "default",
		"object": {
		
		},
		"env_vars": [
		
		],
		"children": {
		"models": [
			{
			"type": "Model",
			"object": {
				"image": "model input dse.modelc.csv",
				"startOffset": 0,
				"endOffset": 25,
				"startLine": 8,
				"endLine": 1,
				"startColumn": 0,
				"endColumn": 26,
				"tokenTypeIdx": 10,
				"payload": {
				"model_name": {
					"value": "input",
					"token_type": "model_name",
					"start_offset": 6,
					"end_offset": 11
				},
				"model_repo_name": {
					"value": "dse.modelc.csv",
					"token_type": "model_repo_name",
					"start_offset": 12,
					"end_offset": 26
				},
				"model_arch": {
					"value": "linux-amd64",
					"token_type": "model_arch",
					"start_offset": 26,
					"end_offset": 37
				},
				"model_uid": {
					"value": "",
					"token_type": "model_uid",
					"start_offset": null,
					"end_offset": null
				}
				}
			},
			"children": {
				"channels": [
				{
					"type": "Channel",
					"object": {
					"image": "channel physical signal_channel",
					"startOffset": 0,
					"endOffset": 30,
					"startLine": 9,
					"endLine": 1,
					"startColumn": 0,
					"endColumn": 31,
					"tokenTypeIdx": 4,
					"payload": {
						"channel_name": {
						"value": "physical",
						"token_type": "channel_name",
						"start_offset": 8,
						"end_offset": 17
						},
						"channel_alias": {
						"value": "signal_channel",
						"token_type": "channel_alias",
						"start_offset": 18,
						"end_offset": 33
						}
					}
					},
					"children": {
					"networks": [
						
					]
					}
				}
				],
				"files": [
				{
					"type": "File",
					"object": {
					"image": "file input.csv input/openloop.csv",
					"startOffset": 0,
					"endOffset": 32,
					"startLine": 11,
					"endLine": 1,
					"startColumn": 0,
					"endColumn": 33,
					"tokenTypeIdx": 5,
					"payload": {
						"file_name": {
						"value": "input.csv",
						"token_type": "file_name",
						"start_offset": 5,
						"end_offset": 15
						},
						"file_reference_type": {
						"value": "",
						"token_type": "file_reference_type",
						"start_offset": null,
						"end_offset": null
						},
						"file_value": {
						"value": "input/openloop.csv",
						"token_type": "file_value",
						"start_offset": 15,
						"end_offset": 34
						}
					}
					}
				},
				{
					"type": "File",
					"object": {
					"image": "file signalgroup.yaml input/signalgroup.yaml",
					"startOffset": 0,
					"endOffset": 43,
					"startLine": 12,
					"endLine": 1,
					"startColumn": 0,
					"endColumn": 44,
					"tokenTypeIdx": 5,
					"payload": {
						"file_name": {
						"value": "signalgroup.yaml",
						"token_type": "file_name",
						"start_offset": 5,
						"end_offset": 22
						},
						"file_reference_type": {
						"value": "",
						"token_type": "file_reference_type",
						"start_offset": null,
						"end_offset": null
						},
						"file_value": {
						"value": "input/signalgroup.yaml",
						"token_type": "file_value",
						"start_offset": 22,
						"end_offset": 45
						}
					}
					}
				}
				],
				"env_vars": [
				{
					"type": "EnvVar",
					"object": {
					"image": "envar CSV_FILE model/input/data/input.csv",
					"startOffset": 0,
					"endOffset": 40,
					"startLine": 10,
					"endLine": 1,
					"startColumn": 0,
					"endColumn": 41,
					"tokenTypeIdx": 11,
					"payload": {
						"env_var_name": {
						"value": "CSV_FILE",
						"token_type": "env_variable_name",
						"start_offset": 6,
						"end_offset": 15
						},
						"env_var_value": {
						"value": "model/input/data/input.csv",
						"token_type": "env_variable_value",
						"start_offset": 15,
						"end_offset": 42
						}
					}
					}
				}
				],
				"workflow": [
				
				]
			}
			},
			{
			"type": "Model",
			"object": {
				"image": "model linear dse.fmi.mcl",
				"startOffset": 0,
				"endOffset": 23,
				"startLine": 14,
				"endLine": 1,
				"startColumn": 0,
				"endColumn": 24,
				"tokenTypeIdx": 10,
				"payload": {
				"model_name": {
					"value": "linear",
					"token_type": "model_name",
					"start_offset": 6,
					"end_offset": 12
				},
				"model_repo_name": {
					"value": "dse.fmi.mcl",
					"token_type": "model_repo_name",
					"start_offset": 13,
					"end_offset": 24
				},
				"model_arch": {
					"value": "linux-amd64",
					"token_type": "model_arch",
					"start_offset": 24,
					"end_offset": 35
				},
				"model_uid": {
					"value": "",
					"token_type": "model_uid",
					"start_offset": null,
					"end_offset": null
				}
				}
			},
			"children": {
				"channels": [
				{
					"type": "Channel",
					"object": {
					"image": "channel physical signal_channel",
					"startOffset": 0,
					"endOffset": 30,
					"startLine": 15,
					"endLine": 1,
					"startColumn": 0,
					"endColumn": 31,
					"tokenTypeIdx": 4,
					"payload": {
						"channel_name": {
						"value": "physical",
						"token_type": "channel_name",
						"start_offset": 8,
						"end_offset": 17
						},
						"channel_alias": {
						"value": "signal_channel",
						"token_type": "channel_alias",
						"start_offset": 18,
						"end_offset": 33
						}
					}
					},
					"children": {
					"networks": [
						
					]
					}
				}
				],
				"files": [
				
				],
				"env_vars": [
				{
					"type": "EnvVar",
					"object": {
					"image": "envar MEASUREMENT_FILE /sim/measurement.mf4",
					"startOffset": 0,
					"endOffset": 42,
					"startLine": 16,
					"endLine": 1,
					"startColumn": 0,
					"endColumn": 43,
					"tokenTypeIdx": 11,
					"payload": {
						"env_var_name": {
						"value": "MEASUREMENT_FILE",
						"token_type": "env_variable_name",
						"start_offset": 6,
						"end_offset": 23
						},
						"env_var_value": {
						"value": "/sim/measurement.mf4",
						"token_type": "env_variable_value",
						"start_offset": 23,
						"end_offset": 44
						}
					}
					}
				}
				],
				"workflow": [
				{
					"type": "Workflow",
					"object": {
					"image": "workflow generate-fmimcl",
					"startOffset": 0,
					"endOffset": 23,
					"startLine": 17,
					"endLine": 1,
					"startColumn": 0,
					"endColumn": 24,
					"tokenTypeIdx": 12,
					"payload": {
						"workflow_name": {
						"value": "generate-fmimcl",
						"token_type": "workflow_name",
						"start_offset": 9,
						"end_offset": 25
						}
					}
					},
					"children": {
					"workflow_vars": [
						{
						"type": "Var",
						"object": {
							"image": "var FMU_DIR uses linear_fmu",
							"startOffset": 0,
							"endOffset": 26,
							"startLine": 18,
							"endLine": 1,
							"startColumn": 0,
							"endColumn": 27,
							"tokenTypeIdx": 9,
							"payload": {
							"var_name": {
								"value": "FMU_DIR",
								"token_type": "variable_name",
								"start_offset": 4,
								"end_offset": 12
							},
							"var_reference_type": {
								"value": "uses",
								"token_type": "variable_reference_type",
								"start_offset": 12,
								"end_offset": 17
							},
							"var_value": {
								"value": "linear_fmu",
								"token_type": "variable_value",
								"start_offset": 17,
								"end_offset": 28
							}
							}
						}
						},
						{
						"type": "Var",
						"object": {
							"image": "var OUT_DIR {{.PATH}}/data",
							"startOffset": 0,
							"endOffset": 25,
							"startLine": 19,
							"endLine": 1,
							"startColumn": 0,
							"endColumn": 26,
							"tokenTypeIdx": 9,
							"payload": {
							"var_name": {
								"value": "OUT_DIR",
								"token_type": "variable_name",
								"start_offset": 4,
								"end_offset": 12
							},
							"var_reference_type": {
								"value": "",
								"token_type": "variable_reference_type",
								"start_offset": null,
								"end_offset": null
							},
							"var_value": {
								"value": "{{.PATH}}/data",
								"token_type": "variable_value",
								"start_offset": 12,
								"end_offset": 27
							}
							}
						}
						},
						{
						"type": "Var",
						"object": {
							"image": "var MCL_PATH {{.PATH}}/lib/libfmimcl.so",
							"startOffset": 0,
							"endOffset": 38,
							"startLine": 20,
							"endLine": 1,
							"startColumn": 0,
							"endColumn": 39,
							"tokenTypeIdx": 9,
							"payload": {
							"var_name": {
								"value": "MCL_PATH",
								"token_type": "variable_name",
								"start_offset": 4,
								"end_offset": 13
							},
							"var_reference_type": {
								"value": "",
								"token_type": "variable_reference_type",
								"start_offset": null,
								"end_offset": null
							},
							"var_value": {
								"value": "{{.PATH}}/lib/libfmimcl.so",
								"token_type": "variable_value",
								"start_offset": 13,
								"end_offset": 40
							}
							}
						}
						}
					]
					}
				}
				]
			}
			}
		]
		}
	}
	]
}
}
```
</details>

###### 2. Transform the parsed JSON into a YAML-based Abstract Syntax Tree (AST)

```bash
$ dse-ast convert -input openloop.ast.json -output openloop.yaml
```

<details>
<summary>openloop.yaml</summary>

```yaml
---
kind: Simulation
metadata:
  labels:
    generator: ast convert
    input_file: /mnt/c/Users/NUZ2KOR/Desktop/dse.sdp/examples/vscode/openloop.ast.json
spec:
  arch: linux-amd64
  channels:
    - name: physical
      networks: []
  endtime: 0.005
  stacks:
    - models:
        - arch: linux-amd64
          channels:
            - alias: signal_channel
              name: physical
          env:
            - name: CSV_FILE
              value: model/input/data/input.csv
          files:
            - name: input.csv
              value: input/openloop.csv
            - name: signalgroup.yaml
              value: input/signalgroup.yaml
          metadata:
            container: {}
            repository: ghcr.io/boschglobal/dse-modelc
            models:
              dse.modelc.csv:
                channels:
                  - alias: signal_channel
                    displayName: dse.modelc.csv
                    name: csv
                    path: examples/csv
                    platforms:
                      - linux-amd64
                      - linux-x86
                      - linux-i386
                      - windows-x64
                      - windows-x86
                    workflows: []
            package:
              download: '{{.REPO}}/releases/download/v{{.TAG}}/ModelC-{{.TAG}}-{{.PLATFORM_ARCH}}.zip'
            tasks: {}
          model: dse.modelc.csv
          name: input
          uses: dse.modelc
          workflows: []
        - arch: linux-amd64
          channels:
            - alias: signal_channel
              name: physical
          env:
            - name: MEASUREMENT_FILE
              value: /sim/measurement.mf4
          metadata:
            container: {}
            repository: ghcr.io/boschglobal/dse-fmi
            models:
              dse.fmi.mcl:
                channels:
                  - alias: signal_channel
                  - alias: network_channel
                displayName: dse.fmi.mcl
                mcl: true
                name: fmimcl
                path: fmimcl
                platforms:
                  - linux-amd64
                  - linux-x86
                  - linux-i386
                  - windows-x64
                  - windows-x86
                workflows:
                  - generate-fmimcl
                  - patch-signalgroup
            package:
              download: '{{.REPO}}/releases/download/v{{.TAG}}/Fmi-{{.TAG}}-{{.PLATFORM_ARCH}}.zip'
            tasks:
              generate-fmimcl:
                generates:
                  - data/model.yaml
                  - data/signalgroup.yaml
          model: dse.fmi.mcl
          name: linear
          uses: dse.fmi
          workflows:
            - name: generate-fmimcl
              vars:
                - name: FMU_DIR
                  reference: uses
                  value: linear_fmu
                - name: OUT_DIR
                  value: '{{.PATH}}/data'
                - name: MCL_PATH
                  value: '{{.PATH}}/lib/libfmimcl.so'
      name: default
  stepsize: 0.0005
  uses:
    - metadata:
        container: {}
        repository: ghcr.io/boschglobal/dse-modelc
      name: dse.modelc
      url: https://github.com/boschglobal/dse.modelc
      version: v2.1.23
    - metadata:
        container: {}
        repository: ghcr.io/boschglobal/dse-fmi
      name: dse.fmi
      url: https://github.com/boschglobal/dse.fmi
      version: v1.1.23
    - metadata: {}
      name: linear_fmu
      path: examples/fmu/linear/fmi2/linear.fmu
      url: https://github.com/boschglobal/dse.fmi/releases/download/v1.1.23/Fmi-1.1.23-linux-amd64.zip
  vars: []

```
</details>

##### Key Characteristics
- Each DSL element (e.g., `model`, `stack`, `var`, `workflow`) is represented as a node in the AST.
- Relationships and nesting are preserved — e.g., models within a stack, or variables inside a workflow.
- Additional metadata (like `uid`, `arch`, `stacked`, `sequential`) is included as attributes in relevant nodes.

##### Use Cases
- Automated code generation (e.g.YAML, configuration files etc...).
- Enabling tooling such as live previews.

##### Output Format
The AST is typically output as a `.yaml` or `.json` file, which can be further processed by downstream tools like `dse-ast generate` and `resolve`.

## Troubleshooting

If you are unable to run the following commands:

- `dse-ast`
- `dse-graph`
- `dse-mdf2csv`

It may be because the directory containing these executables (`$HOME/.local/bin`) is not in your `PATH`.

You can fix this by running the below command or by restarting the terminal:

```bash
source ~/.bashrc 
```

This ensures your shell can find the installed dse-* commands every time you open a terminal.