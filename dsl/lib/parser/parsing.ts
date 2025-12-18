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
} from "../lexer/lexing";
import { EmbeddedActionsParser, IToken, ILexingError } from "chevrotain";
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
  public files!: () => any;
  public stack!: () => any;
  public models!: () => any;
  constructor() {
    super(allTokens);
    const $ = this;

    function updateTokenObject(tokenObject: IToken): Omit<IToken, "tokenType"> {
      const { tokenType, ...rest } = tokenObject;
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
    let channelCollection: Record<string, Network[]>

    // The main rule that parses the entire simulation statement.
    $.RULE("simulation", () => {
      const simulation = $.CONSUME(Simulation);
      const children: Record<string, any> = {};
      children.channels = $.SUBRULE($.channels);

      let uses = [];
      while ($.LA(1).tokenType === Uses) {
        uses = $.SUBRULE($.uses);
      }
      children.uses = uses;

      while ($.LA(1).tokenType === Var || $.LA(1).tokenType === EnvVar) {
        if ($.LA(1).tokenType === Var) {
          global_vars = $.SUBRULE($.vars);
        } else if ($.LA(1).tokenType === EnvVar) {
          global_env_vars = $.SUBRULE($.envvars);
        }
      }
      children.vars = global_vars;
      children.env_vars = global_env_vars;

      children.stacks = $.SUBRULE($.stacked_models);
      return {
        type: simulation.tokenType.name,
        simulation: simulation.image.replace("\n", ""),
        object: updateTokenObject(simulation),
        children: children,
      };
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
        if (global_env_vars.length > 0) {
          let merged_envars = mergeByName(global_env_vars, stack_env_vars, env => env.object.payload.env_var_name.value); // overrides global level envars if same envar name in stack_env_vars
          model_env_vars = mergeByName(merged_envars, model_env_vars, env => env.object.payload.env_var_name.value); // overrides with modellevel envars if same envar name in merged_envars
        }

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
          const stack = $.SUBRULE($.stack);

          // Read env_vars and annotations in any order
          while ($.LA(1).tokenType === EnvVar || $.LA(1).tokenType === Annotation) {
            if ($.LA(1).tokenType === EnvVar) {
              stack_env_vars.push(...$.SUBRULE($.envvars));
            } else if ($.LA(1).tokenType === Annotation) {
              stack_annotations.push(...$.SUBRULE($.annotations));
            }
          }

          stack_env_vars = Array.isArray(stack_env_vars) ? stack_env_vars : [stack_env_vars];
          if (global_env_vars.length > 0) {
            const merged_envars = mergeByName(global_env_vars, stack_env_vars, env => env.object.payload.env_var_name.value);
            stack_env_vars = merged_envars;
          }

          stack_annotations = Array.isArray(stack_annotations) ? stack_annotations : [stack_annotations];

          let name = "default";
          if (stack && "tokenType" in stack) {
            name = stack.payload.stack_name.value;
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
              children: {
                models: models,
              },
            });

            if (externalModels.length !== 0) {
              stacks.push({
                type: "Stack",
                name: "external",
                object: updateTokenObject(stack),
                //env_vars: env_vars,
                children: {
                  models: externalModels,
                },
              });
            }
          }
        },
      });
      const hasOtherModels = stacks.some(
        s => s.name !== "default" && s.name !== "external" && Array.isArray(s.children?.models) && s.children.models.length > 0
      );
      stacks = stacks
        .filter(stack => {
          if (stack.name === "external") return true;
          if (stack.name === "default") return !hasOtherModels;
          return !(Array.isArray(stack.children?.models) && stack.children.models.length === 0);
        }) // keep 'external'; remove 'default' only if other stacks (not 'external') have models
        .sort((a, b) => (a.name === "external" ? 1 : b.name === "external" ? -1 : 0)); // move 'external' stack to the end (for simbus model to appear correct stack)
      return stacks;
    });

    // Perform self-analysis of the grammar to optimize the parser.
    this.performSelfAnalysis();
  }
}

// Initialize the parser instance.
export const parserInstance = new FsilParser();
type DiagnosticLike = {
  severity: number; // you can define 0 = Error, 1 = Warning, etc.
  message: string;
  range: {
    start: { line: number; character: number };
    end: { line: number; character: number };
  };
  source: string;
};
export function parse(inputText: string): DiagnosticLike[] | any {
  const lines = inputText.split(/\r?\n/).map(line => line.replace(/\s*#.*$/, "").trim());
  const diagnostics: DiagnosticLike[] = [];
  let lexingResult = {
    tokens: [] as IToken[],
    groups: {} as Record<string, IToken[]>,
    errors: [] as ILexingError[],
  };
  let lineIdx = 0;
  lines.forEach((line) => {
    if (line.trim() !== "") {
      const res = fsilLexer.tokenize(line);
      lineIdx += 1;
      res.tokens.forEach((token) => {
        token.startLine = lineIdx - 1;
        token.startColumn = token.startOffset ?? 0;
      });
      lexingResult.tokens.push(...res.tokens);
      lexingResult.errors.push(...res.errors);
      if (res.errors.length > 0) {
        diagnostics.push({
          severity: 0,
          message: "Syntax error: Unexpected token or malformed statement.",
          range: {
            start: { line: lineIdx - 1, character: 0 },
            end: { line: lineIdx - 1, character: lines[lineIdx - 1].length },
          },
          source: "parser",
        });
      }
    } else {
      lineIdx += 1;
    }
  });

  parserInstance.input = lexingResult.tokens;
  const astOutput = parserInstance.simulation();
  if (parserInstance.errors.length > 0) {
    console.log("\nParsing errors detected!\n");
    parserInstance.errors.forEach((err, idx) => {
      const token = err.token;
      const line = token?.startLine ?? 0;
      const col = err.message?.split(":").pop()?.length ?? 0;

      diagnostics.push({
        severity: 0,
        message: "Syntax error: Token declaration out of scope.",
        range: {
          start: { line: line, character: 0 },
          end: { line: line, character: col + 1 },
        },
        source: "parser",
      });
    });
  }

  if (diagnostics.length > 0) {
    return diagnostics;
  }
  return astOutput;
}
