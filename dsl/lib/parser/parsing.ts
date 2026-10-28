// Copyright 2024 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

import {
  fsilLexer,
  allTokens,
  Simulation,
  Channel,
  Network,
  File,
  Uses,
  UseItem,
  Model,
  Var,
  EnvVar,
  Annotation,
  Workflow,
  Stack,
  FileDelimiter,
  UnknownStatementKeyword,
  FileContent,
} from "../lexer/lexing";
import { EmbeddedActionsParser, IToken, ILexingError } from "chevrotain";
class ScriptSyntaxError extends Error {
  constructor(message: string, public token: IToken) {
    super(message);
  }
}

class FsilParser extends EmbeddedActionsParser {
  public simulation!: () => any;
  public channels!: () => any;
  public uses!: () => any;
  public vars!: () => any;
  public envvars!: () => any;
  public annotations!: () => any;
  public stacked_models!: () => any;
  public networks!: () => any;
  public workflow_vars!: () => any;
  public workflow!: () => any;
  public globalworkflow!: () => any;
  public files!: () => any;
  public stack!: () => any;
  public models!: () => any;
  public dynamic_files!: () => any;
  public filecontent!: () => any;
  constructor() {
    super(allTokens);
    const $ = this;

    function updateTokenObject(tokenObject: IToken): Omit<IToken, "tokenType"> {
      const { tokenType, ...rest } = tokenObject;
      if (rest.payload?.source_ranges) {
        const payload = { ...rest.payload };
        delete payload.source_ranges;
        rest.payload = payload;
      }
      return rest;
    }

    function mergeByName(global: any[], inner: any[], getName: (item: any) => string): any[] {
      const mergedMap = new Map<string, any>();
      // Add all global vars/envars
      for (const item of global) {
        const name = getName(item);
        mergedMap.set(name, item);
      }
      // Add all inner vars/envars - overrides added global item if same var/envar name
      for (const item of inner) {
        const name = getName(item);
        mergedMap.set(name, item);
      }
      return Array.from(mergedMap.values());
    }

    interface Network {
      signal: string;
      mimetype: string;
    }

    let global_vars: string[] = [];
    let global_env_vars: string[] = [];
    let stack_env_vars: string[] = [];
    let stack_annotations: string[] = [];
    let model_vars: string[] = [];
    let default_stack_workflows: string[] = [];
    let simulation_arch = "";

    // The main rule that parses the entire simulation statement.
    $.RULE("simulation", () => {
      const simulation = $.CONSUME(Simulation);
      $.ACTION(() => {
        simulation_arch = simulation.payload.simulation_arch.value;
      });
      const children: Record<string, any> = {};
      children.channels = $.SUBRULE($.channels);

      let uses = [];
      while ($.LA(1).tokenType === Uses) {
        uses = $.SUBRULE($.uses);
      }
      children.uses = uses;

      while ($.LA(1).tokenType === Var || $.LA(1).tokenType === EnvVar || $.LA(1).tokenType === Workflow) {
        if ($.LA(1).tokenType === Var) {
          global_vars = $.SUBRULE($.vars);
        } else if ($.LA(1).tokenType === EnvVar) {
          global_env_vars = $.SUBRULE($.envvars);
        } else if ($.LA(1).tokenType === Workflow) {
          default_stack_workflows = $.SUBRULE($.globalworkflow);
        }
      }
      children.vars = global_vars;
      children.env_vars = global_env_vars;

      children.stacks = $.SUBRULE($.stacked_models);

      const generated_files = $.SUBRULE($.dynamic_files);
      children.generated_files = generated_files

      return {
        type: simulation.tokenType.name,
        simulation: simulation.image.replace("\n", ""),
        object: updateTokenObject(simulation),
        children: children,
      };
    });

    $.RULE("dynamic_files", () => {
      const dynamic_files: any[] = [];
      let fileContent : any[] = [];
      $.MANY({
        DEF: () => {
          const fileDelimiter = $.CONSUME(FileDelimiter);
          while ($.LA(1).tokenType === FileContent) {
              fileContent = $.SUBRULE($.filecontent);
          }
          $.ACTION(() => {
            fileDelimiter.payload ??= {};
            fileDelimiter.payload.file_content = {
              value: fileContent.join("\n"),
              token_type: "file_content",
            };

            dynamic_files.push({
              type: "File",
              object: updateTokenObject(fileDelimiter),
            });
          });
        },
      });
      return dynamic_files;
    });

    $.RULE("filecontent", () => {
      let fileContent : any[] = [];
       $.MANY({
        DEF: () => {
          const lineObj = $.CONSUME(FileContent);
          fileContent.push(lineObj.image)
        },
      });
      return fileContent
    });

    // Rule for parsing the channels.
    $.RULE("channels", () => {
      const channels: any[] = [];
      $.MANY({
        DEF: () => {
          const channel = $.CONSUME(Channel);
          const networks = $.SUBRULE($.networks);
          channels.push({
            type: channel.tokenType.name.replace(" ", ""),
            object: updateTokenObject(channel),
            children: {
              networks: networks,
            },
          });
        },
      });
      return channels;
    });

    // Rule for parsing the networks.
    $.RULE("networks", () => {
      const networks: any[] = [];
      $.MANY({
        DEF: () => {
          const network = $.CONSUME(Network);
          networks.push({
            type: network.tokenType.name.replace(" ", ""),
            object: updateTokenObject(network),
          });
        },
      });
      return networks;
    });

    // Rule for parsing the uses.
    $.RULE("uses", () => {
      const uses: any[] = [];
      const useBlock = $.CONSUME(Uses);
      $.MANY({
        DEF: () => {
          const use_item = $.CONSUME(UseItem);
          uses.push({
            type: useBlock.tokenType.name,
            object: updateTokenObject(use_item),
          });
        },
      });
      return uses;
    });

    $.RULE("vars", () => {
      const vars: any[] = [];
      $.MANY({
        DEF: () => {
          const variable = $.CONSUME(Var);
          vars.push({
            type: variable.tokenType.name.replace(" ", ""),
            object: updateTokenObject(variable),
          });
        },
      });
      return vars;
    });

    $.RULE("files", () => {
      const files: any[] = [];
      $.MANY({
        DEF: () => {
          const file = $.CONSUME(File);
          files.push({
            type: file.tokenType.name.replace(" ", ""),
            object: updateTokenObject(file),
          });
        },
      });
      return files;
    });

    $.RULE("envvars", () => {
      const env_vars: any[] = [];
      $.MANY({
        DEF: () => {
          const variable = $.CONSUME(EnvVar);
          env_vars.push({
            type: variable.tokenType.name.replace(" ", ""),
            object: updateTokenObject(variable),
          });
        },
      });
      return env_vars;
    });

    $.RULE("annotations", () => {
      const annotations: any[] = [];
      $.MANY({
        DEF: () => {
          const variable = $.CONSUME(Annotation);
          annotations.push({
            type: variable.tokenType.name.replace(" ", ""),
            object: updateTokenObject(variable),
          });
        },
      });
      return annotations;
    });

    $.RULE("workflow_vars", () => {
      const workflow_vars: any[] = [];
      $.MANY({
        DEF: () => {
          const variable = $.CONSUME(Var);
          workflow_vars.push({
            type: variable.tokenType.name.replace(" ", ""),
            object: updateTokenObject(variable),
          });
        },
      });
      return workflow_vars;
    });

    $.RULE("globalworkflow", () => {
      const global_workflows: any[] = [];
      $.MANY({
        DEF: () => {
          const global_workflow = $.CONSUME(Workflow);
          if (this.RECORDING_PHASE !== true) {
            const payload = global_workflow.payload;
            if (!payload.workflow_reference_type.value || payload.workflow_reference_type.value !== "uses") {
              throw new ScriptSyntaxError("Stack workflows must have 'uses'.", global_workflow);
            }
          }

          let workflow_vars = $.SUBRULE($.workflow_vars);
          global_workflows.push({
            type: global_workflow.tokenType.name.replace(" ", ""),
            object: updateTokenObject(global_workflow),
            children: {
              workflow_vars: workflow_vars,
            },
          });
        },
      });
      return global_workflows;
    });

    $.RULE("workflow", () => {
      const workflows: any[] = [];
      $.MANY({
        DEF: () => {
          const workflow = $.CONSUME(Workflow);
          let workflow_vars = $.SUBRULE($.workflow_vars);

          workflow_vars = Array.isArray(workflow_vars) ? workflow_vars : [workflow_vars];
          if (model_vars.length > 0) {
            let merged_vars = mergeByName(model_vars, workflow_vars, env => env.object.payload.var_name.value); // overrides workflow level vars if same var name in model_vars
            workflow_vars = merged_vars;
          }

          workflows.push({
            type: workflow.tokenType.name.replace(" ", ""),
            object: updateTokenObject(workflow),
            children: {
              workflow_vars: workflow_vars,
            },
          });
        },
      });
      return workflows;
    });

    $.RULE("models", () => {
      const models: any[] = [];

      $.MANY(() => {
        let model_files: string[] = [];
        let model_env_vars: string[] = [];
        model_vars = [];
        let model_annotations: string[] = [];

        const model = $.CONSUME(Model);
        const model_channels = $.SUBRULE($.channels);

        // Read model_files and env_vars in any order
        while ($.LA(1).tokenType === File || $.LA(1).tokenType === EnvVar || $.LA(1).tokenType === Var || $.LA(1).tokenType === Annotation) {
          if ($.LA(1).tokenType === File) {
            model_files.push(...$.SUBRULE($.files));
          } else if ($.LA(1).tokenType === EnvVar) {
            model_env_vars.push(...$.SUBRULE($.envvars));
          } else if ($.LA(1).tokenType === Var) {
            model_vars.push(...$.SUBRULE($.vars));
          } else if ($.LA(1).tokenType === Annotation) {
            model_annotations.push(...$.SUBRULE($.annotations));
          }
        }

        model_env_vars = Array.isArray(model_env_vars) ? model_env_vars : [model_env_vars];

        model_vars = Array.isArray(model_vars) ? model_vars : [model_vars];
        if (global_vars.length > 0) {
          let merged_vars = mergeByName(global_vars, model_vars, env => env.object.payload.var_name.value); // overrides global level vars if same envar name in model_vars
          model_vars = merged_vars;
        }

        const workflow = $.SUBRULE($.workflow);

        models.push({
          type: model.tokenType.name,
          object: updateTokenObject(model),
          children: {
            channels: model_channels,
            files: model_files,
            env_vars: model_env_vars,
            vars: model_vars,
            annotations: model_annotations,
            workflow: workflow,
          },
        });
      });

      return models;
    });

    $.RULE("stack", () => {
      let stack: any = "";
      $.MANY({
        DEF: () => {
          stack = $.CONSUME(Stack);
        },
      });
      stack_env_vars = [];
      stack_annotations = [];
      return stack;
    });

    $.RULE("stacked_models", () => {
      let stacks: any[] = [];
      $.MANY({
        DEF: () => {
          let stack: any = "";
          stack_env_vars = [];
          stack_annotations = [];
          if ($.LA(1).tokenType === Stack) {
            stack = $.CONSUME(Stack);
          }
          let stack_workflows: string[] = [];
          // Read env_vars and annotations in any order
          while ($.LA(1).tokenType === EnvVar || $.LA(1).tokenType === Annotation || $.LA(1).tokenType === Workflow) {
            if ($.LA(1).tokenType === EnvVar) {
              stack_env_vars.push(...$.SUBRULE($.envvars));
            } else if ($.LA(1).tokenType === Annotation) {
              stack_annotations.push(...$.SUBRULE($.annotations));
            } else if ($.LA(1).tokenType === Workflow) {
              stack_workflows.push(...$.SUBRULE($.globalworkflow));
            }
          }

          stack_env_vars = Array.isArray(stack_env_vars) ? stack_env_vars : [stack_env_vars];
          if (global_env_vars.length > 0) {
            const merged_envars = mergeByName(global_env_vars, stack_env_vars, env => env.object.payload.env_var_name.value);
            stack_env_vars = merged_envars;
          }

          stack_annotations = Array.isArray(stack_annotations) ? stack_annotations : [stack_annotations];
          stack_workflows = Array.isArray(stack_workflows) ? stack_workflows : [stack_workflows];

          let name = "";
          if (stack && "tokenType" in stack) {
            name = stack.payload.stack_name.value;
          } else {
            name = "default";
            stack = {
              tokenType: Stack,
              payload: {
                stack_name: { value: name, token_type: "stack_name" },
                stacked: { value: "", token_type: "stacked" },
                sequential: { value: "", token_type: "sequential" },
                stack_arch: {
                  value: simulation_arch,
                  token_type: "stack_arch",
                },
              },
            };
            stack_workflows = default_stack_workflows;
          }

          const stacked_models = $.SUBRULE($.models);
          if (stacked_models.length !== 0 && !('description' in stacked_models)) {

            let externalModels: any[] = [];
            let models: any[] = [];
            // external models must be pushed to a seperate stack named "external"
            stacked_models.forEach((model: any) => {
              if (model.object.payload.external.value === "true" && model.object.payload.model_repo_name.value === "") {
                externalModels.push(model);
              } else {
                models.push(model)
              }
            });

            stacks.push({
              type: "Stack",
              name: name,
              object: updateTokenObject(stack),
              env_vars: stack_env_vars,
              annotations: stack_annotations,
              workflows: stack_workflows,
              children: {
                models: models,
              },
            });

            if (externalModels.length !== 0) {
              const externalStack = {
                ...stack,
                payload: {
                  ...stack.payload,
                  stack_name: { value: "external", token_type: "stack_name" },
                },
              };
              stacks.push({
                type: "Stack",
                name: "external",
                object: updateTokenObject(externalStack),
                //env_vars: env_vars,
                children: {
                  models: externalModels,
                },
              });
            }
          }
        },
      });
      stacks = stacks
        .filter(stack => {
          if (stack.name === "external") return true;
          if (stack.name === "default") {
            return Array.isArray(stack.children?.models) && stack.children.models.length > 0;
          }
          return !(Array.isArray(stack.children?.models) && stack.children.models.length === 0);
        }) // keep 'external'; remove 'default' only if other stacks (not 'external') have models
        .sort((a, b) =>
          a.name === "external" ? 1 :
          b.name === "external" ? -1 : 0
        ); // move 'external' stack to the end (for simbus model to appear correct stack)
      return stacks;
    });

    // Perform self-analysis of the grammar to optimize the parser.
    this.performSelfAnalysis();
  }
}

// Initialize the parser instance.
export const parserInstance = new FsilParser();
export type DiagnosticLike = {
  severity: number; // you can define 0 = Error, 1 = Warning, etc.
  code?: string;
  message: string;
  expected?: string[][];
  expectedSyntax: string[];
  range: {
    start: { line: number; character: number };
    end: { line: number; character: number };
  };
  source: string;
};

function expectedSyntax(error: { name?: string; message: string }): string[][] | undefined {
  if (error.name === "NotAllInputParsedException" || error.message.startsWith("Redundant input, expecting EOF")) {
    return [["EOF"]];
  }

  const mismatch = error.message.match(/token of type --> (.+?) <--/);
  if (mismatch) {
    return [[mismatch[1]]];
  }

  const alternatives = [...error.message.matchAll(/\[([^\]]*)\]/g)]
    .map((match) => match[1].split(/,\s*/).filter(Boolean));
  return alternatives.length > 0 ? alternatives : undefined;
}

const syntaxByKeyword: Record<string, string[]> = {
  simulation: ["simulation [arch=<arch>] [stepsize=<number>] [endtime=<number>]"],
  channel: ["channel <name> [alias]"],
  network: ["network <name> '<mime-type>'"],
  uses: ["uses", "<name> <repository-or-url> [v<version>] [path=<path>] [user=<user>] [token=<token>]"],
  var: ["var <name> [uses|network|var] <value> [mimetype|signal]"],
  model: ["model <name> <repository> [arch=<arch>] [uid=<number>] [external=<bool>]"],
  envar: ["envar <name> <value>"],
  file: ["file <name> [uses] <value> [path=<path>]"],
  annotation: ["annotation <name> <value>"],
  workflow: ["workflow <name> [uses <workflow-reference>]"],
  stack: ["stack <name> [stacked=true|false] [sequential=true|false] [arch=<arch>]"],
};

const generalSyntax = [
  "simulation [arch=<arch>] [stepsize=<number>] [endtime=<number>]",
  "channel <name> [alias]",
  "uses, followed by <name> <repository-or-url> [v<version>] [path=<path>]",
  "var <name> [uses|network|var] <value> [mimetype|signal]",
  "model <name> <repository> [arch=<arch>] [uid=<number>] [external=<bool>]",
  "envar <name> <value>",
  "file <name> [uses] <value> [path=<path>]",
  "annotation <name> <value>",
  "workflow <name> [uses <workflow-reference>]",
  "stack <name> [stacked=true|false] [sequential=true|false] [arch=<arch>]",
];

function editDistance(left: string, right: string): number {
  const previous = Array.from({ length: right.length + 1 }, (_, index) => index);
  for (let leftIndex = 1; leftIndex <= left.length; leftIndex += 1) {
    const current = [leftIndex];
    for (let rightIndex = 1; rightIndex <= right.length; rightIndex += 1) {
      current[rightIndex] = Math.min(
        current[rightIndex - 1] + 1,
        previous[rightIndex] + 1,
        previous[rightIndex - 1] + (left[leftIndex - 1] === right[rightIndex - 1] ? 0 : 1),
      );
    }
    previous.splice(0, previous.length, ...current);
  }
  return previous[right.length];
}

function likelyKeywordTypo(keyword: string): string | undefined {
  const candidates = Object.keys(syntaxByKeyword)
    .map((candidate) => ({ candidate, distance: editDistance(keyword, candidate) }))
    .sort((left, right) => left.distance - right.distance);
  const maximumDistance = Math.min(2, Math.max(1, Math.floor(keyword.length / 4)));
  if (candidates[0].distance > maximumDistance || candidates[0].distance === candidates[1]?.distance) {
    return undefined;
  }
  return candidates[0].candidate;
}

function expectedStatementSyntax(text: string, code?: string): string[] {
  if (code === "DSE1003") {
    return ["workflow <name> uses <workflow-reference>"];
  }
  const keyword = text.trimStart().match(/^([a-z]+)/i)?.[1].toLowerCase();
  if (!keyword) {
    return generalSyntax;
  }
  return syntaxByKeyword[keyword] || syntaxByKeyword[likelyKeywordTypo(keyword) || ""] || generalSyntax;
}

export function parse(inputText: string): DiagnosticLike[] | any {
  const lines = inputText.split(/\r?\n/).map(line => line.replace(/\s*#.*$/, ""));
  const diagnostics: DiagnosticLike[] = [];
  let lexingResult = {
    tokens: [] as IToken[],
    groups: {} as Record<string, IToken[]>,
    errors: [] as ILexingError[],
  };
  let insideFileContent = false;
  let insideUsesSection = false;
  let lineIdx = 0;
  lines.forEach((line) => {
    if (line.trim() !== "") {
      const res = fsilLexer.tokenize(line);
      lineIdx += 1;
      res.tokens.forEach((token) => {
        token.startLine = lineIdx - 1;
        token.startColumn = token.startOffset ?? 0;
      });
      lexingResult.errors.push(...res.errors);
      const isFileDelimiter = res.tokens.some(token => token.tokenType === FileDelimiter);
      const isFileContent = res.tokens.length === 1 && res.tokens[0].tokenType === FileContent;
      const isUnknownStatementKeyword = res.tokens.length === 1 && res.tokens[0].tokenType === UnknownStatementKeyword;
      const isWorkflow = res.tokens.length === 1 && res.tokens[0].tokenType === Workflow;
      const hasLexingErrors = res.errors.length > 0;

      for (const error of res.errors) {
        diagnostics.push({
          severity: 0,
          code: "DSE1001",
          message: `Unexpected text ${JSON.stringify(line.slice(error.offset, error.offset + error.length))}. Check the statement syntax.`,
          expectedSyntax: expectedStatementSyntax(line),
          range: {
            start: { line: lineIdx - 1, character: error.offset },
            end: { line: lineIdx - 1, character: error.offset + error.length },
          },
          source: "parser",
        });
      }

      if (isFileDelimiter) {
        insideFileContent = true;
      } else if (insideFileContent && !isFileContent && !isUnknownStatementKeyword) {
        insideFileContent = false;
      }

      if (hasLexingErrors) {
        return;
      }
      if (isUnknownStatementKeyword) {
        const token = res.tokens[0];
        if (insideFileContent || insideUsesSection) {
          lexingResult.tokens.push({
            ...token,
            tokenType: FileContent,
            tokenTypeIdx: FileContent.tokenTypeIdx,
            payload: {
              file_content: {
                value: token.image.trim(),
                token_type: "file_content",
              },
            },
          } as IToken);
        } else {
          const keywordRange = token.payload.source_range;
          const line = token.startLine ?? lineIdx - 1;
          const baseColumn = token.startColumn ?? 0;
          diagnostics.push({
            severity: 0,
            code: "DSE1002",
            message: `Unexpected or malformed statement: ${token.image.trim()}`,
            expectedSyntax: expectedStatementSyntax(token.image),
            range: {
              start: { line, character: baseColumn + keywordRange.start },
              end: { line, character: baseColumn + keywordRange.end },
            },
            source: "parser",
          });
        }
        return;
      }

      const isUsesHeader = res.tokens.some(token => token.tokenType === Uses);
      const isUsesEntry = res.tokens.some(token => token.tokenType === UseItem);
      if (isUsesHeader) {
        insideUsesSection = true;
      } else if (insideUsesSection && !isUsesEntry && !isFileContent) {
        insideUsesSection = false;
      }

      if (isWorkflow && res.tokens[0].payload.syntax_error && !insideFileContent) {
        const token = res.tokens[0];
        const syntaxError = token.payload.syntax_error;
        const line = token.startLine ?? lineIdx - 1;
        const baseColumn = token.startColumn ?? 0;
        diagnostics.push({
          severity: 0,
          code: "DSE1002",
          message: `Unexpected or malformed statement: ${token.image.trim()}`,
          expectedSyntax: expectedStatementSyntax(token.image),
          range: {
            start: { line, character: baseColumn + syntaxError.start },
            end: { line, character: baseColumn + syntaxError.end },
          },
          source: "parser",
        });
        return;
      }

      if (isFileContent && !insideFileContent) {
        const token = res.tokens[0];
        const startLine = token.startLine ?? lineIdx - 1;
        const startColumn = token.startColumn ?? 0;
        const endColumn = token.endColumn ?? startColumn + token.image.length;
        diagnostics.push({
          severity: 0,
          code: "DSE1002",
          message: `Unexpected or malformed statement: ${token.image.trim()}`,
          expectedSyntax: expectedStatementSyntax(token.image),
          range: {
            start: { line: startLine, character: startColumn },
            end: { line: lineIdx - 1, character: endColumn },
          },
          source: "parser",
        });
        return;
      }
      lexingResult.tokens.push(...res.tokens);
    } else {
      lineIdx += 1;
    }
  });

  if (lexingResult.errors.length > 0) {
    return diagnostics;
  }

  parserInstance.input = lexingResult.tokens;
  let astOutput: any;
  try {
    astOutput = parserInstance.simulation();
  } catch (error) {
    if (!(error instanceof ScriptSyntaxError)) {
      throw error;
    }
    const line = error.token.startLine ?? 0;
    const character = error.token.startColumn ?? 0;
    diagnostics.push({
      severity: 0,
      code: "DSE1003",
      message: error.message,
      expectedSyntax: expectedStatementSyntax(error.token.image, "DSE1003"),
      range: {
        start: { line, character },
        end: { line, character: character + error.token.image.length },
      },
      source: "parser",
    });
  }
  if (parserInstance.errors.length > 0) {
    parserInstance.errors.forEach((err) => {
      const token = err.token;
      const hasEarlierSimulationTypo = diagnostics.some(diagnostic =>
        diagnostic.code === "DSE1002" &&
        diagnostic.expectedSyntax.some(syntax => syntax.startsWith("simulation ")) &&
        diagnostic.range.start.line < (token.startLine ?? lines.length)
      );
      if (hasEarlierSimulationTypo && err.message.includes("Simulation")) {
        return;
      }
      if (diagnostics.some(diagnostic => diagnostic.code === "DSE1002") && err.message.startsWith("Redundant input, expecting EOF")) {
        return;
      }
      const line = Number.isFinite(token.startLine) && token.startLine! >= 0
        ? token.startLine! : lines.length - 1;
      const character = Number.isFinite(token.startColumn) && token.startColumn! >= 0
        ? token.startColumn! : lines[line].length;

      diagnostics.push({
        severity: 0,
        code: "DSE1002",
        message: err.message.startsWith("Redundant input, expecting EOF")
          ? `Unexpected or malformed statement: ${token.image.trim()}`
          : err.message,
        expected: expectedSyntax(err),
        expectedSyntax: expectedStatementSyntax(token.image),
        range: {
          start: { line, character },
          end: { line, character: Math.min(lines[line].length, character + token.image.length) },
        },
        source: "parser",
      });
    });
  }

  diagnostics.sort((left, right) =>
    left.range.start.line - right.range.start.line ||
    left.range.start.character - right.range.start.character
  );

  if (diagnostics.length > 0) {
    return diagnostics;
  }
  return astOutput;
}
