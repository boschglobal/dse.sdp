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
    Workflow,
    Stack,
} from "../lexer/lexing";
import {
    EmbeddedActionsParser,
    IToken,
    ILexingError
} from "chevrotain";
class FsilParser extends EmbeddedActionsParser {
    public simulation!: () => any;
    public channels!: () => any;
    public uses!: () => any;
    public vars!: () => any;
    public envvars!: () => any;
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

        let global_env_vars: string[] = [];

        // The main rule that parses the entire simulation statement.
        $.RULE("simulation", () => {
            const simulation = $.CONSUME(Simulation);
            const children: Record<string, any> = {};
            children.channels = $.SUBRULE($.channels);
            children.uses = $.SUBRULE($.uses);

            let vars = [];
            while ($.LA(1).tokenType === Var || $.LA(1).tokenType === EnvVar) {
                if ($.LA(1).tokenType === Var) {
                    vars = $.SUBRULE($.vars);
                } else if ($.LA(1).tokenType === EnvVar) {
                    global_env_vars = $.SUBRULE($.envvars);
                }
            }
            children.vars = vars;

            children.stacks = $.SUBRULE($.stacked_models);
            return {
                type: simulation.tokenType.name,
                simulation: simulation.image.replace('\n', ''),
                object: updateTokenObject(simulation),
                children: children
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
                        type: channel.tokenType.name.replace(' ', ''),
                        object: updateTokenObject(channel),
                        children: {
                            networks: networks
                        }
                    });
                }
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
                        type: network.tokenType.name.replace(' ', ''),
                        object: updateTokenObject(network),
                    })
                }
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
                        object: updateTokenObject(use_item)
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
                        type: variable.tokenType.name.replace(' ', ''),
                        object: updateTokenObject(variable)
                    })
                }
            });
            return vars;
        });

        $.RULE("files", () => {
            const files: any[] = [];
            $.MANY({
                DEF: () => {
                    const file = $.CONSUME(File);
                    files.push({
                        type: file.tokenType.name.replace(' ', ''),
                        object: updateTokenObject(file)
                    })
                }
            });
            return files;
        });

        $.RULE("envvars", () => {
            const env_vars: any[] = [];
            $.MANY({
                DEF: () => {
                    const variable = $.CONSUME(EnvVar);
                    env_vars.push({
                        type: variable.tokenType.name.replace(' ', ''),
                        object: updateTokenObject(variable)
                    })
                }
            });
            return env_vars;
        });

        $.RULE("workflow_vars", () => {
            const workflow_vars: any[] = [];
            $.MANY({
                DEF: () => {
                    const variable = $.CONSUME(Var);
                    workflow_vars.push({
                        type: variable.tokenType.name.replace(' ', ''),
                        object: updateTokenObject(variable)
                    })
                }
            });
            return workflow_vars;
        });

        $.RULE("workflow", () => {
            const workflows: any[] = [];
            $.MANY({
                DEF: () => {
                    const workflow = $.CONSUME(Workflow);
                    const workflow_vars = $.SUBRULE($.workflow_vars);
                    workflows.push({
                        type: workflow.tokenType.name.replace(' ', ''),
                        object: updateTokenObject(workflow),
                        children: {
                            workflow_vars: workflow_vars,
                        }
                    })
                }
            });
            return workflows;
        });

        $.RULE("models", () => {
            const models: any[] = [];

            $.MANY(() => {
                const model = $.CONSUME(Model);
                const model_channels = $.SUBRULE($.channels);

                let model_files = [];
                let env_vars = [];
                // Read model_files and env_vars in any order
                while ($.LA(1).tokenType === File || $.LA(1).tokenType === EnvVar) {
                    if ($.LA(1).tokenType === File) {
                        model_files = $.SUBRULE($.files);
                    } else if ($.LA(1).tokenType === EnvVar) {
                        env_vars = $.SUBRULE($.envvars);
                    }
                }

                const workflow = $.SUBRULE($.workflow);

                models.push({
                    type: model.tokenType.name,
                    object: updateTokenObject(model),
                    children: {
                        channels: model_channels,
                        files: model_files,
                        env_vars: env_vars,
                        workflow: workflow,
                    }
                });
            });

            return models;
        });



        $.RULE("stack", () => {
            let stack: any = '';
            $.MANY({
                DEF: () => {
                    stack = $.CONSUME(Stack);
                }
            });
            return stack;
        });

        $.RULE("stacked_models", () => {
            let stacks: any[] = []
            $.MANY({
                DEF: () => {
                    const stack = $.SUBRULE($.stack);

                    let env_vars = $.SUBRULE($.envvars);
                    env_vars = Array.isArray(env_vars) ? env_vars : [env_vars]
                    if (global_env_vars.length > 0) {
                        env_vars.push(...global_env_vars);
                    }

                    let name = 'default'
                    if (stack && 'tokenType' in stack) {
                        name = stack.payload.stack_name.value;
                    }

                    const stacked_models = $.SUBRULE($.models);
                    if (stacked_models.length !== 0) {
                        stacks.push({
                            type: 'Stack',
                            name: name,
                            object: updateTokenObject(stack),
                            env_vars: env_vars,
                            children: {
                                models: stacked_models,
                            }
                        });
                    }
                }
            });
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
    const lines = inputText.split(/\r?\n/);
    const diagnostics: DiagnosticLike[] = [];
    let lexingResult = {
        tokens: [] as IToken[],
        groups: {} as Record<string, IToken[]>,
        errors: [] as ILexingError[]
    };
    let lineIdx = 0;
    lines.forEach((line) => {
        if (line.trim() !== '') {
            const res = fsilLexer.tokenize(line);
            lineIdx += 1;
            res.tokens.forEach(token => {
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
                        end: { line: lineIdx - 1, character: lines[lineIdx - 1].length }
                    },
                    source: "parser"
                });
            }
        } else {
            lineIdx += 1;
        }
    });

    parserInstance.input = lexingResult.tokens;
    const astOutput = parserInstance.simulation();
    if (parserInstance.errors.length > 0) {
        console.log("\nParsing errors detected!\n")
        parserInstance.errors.forEach((err, idx) => {
            const token = err.token;
            const line = token?.startLine ?? 0;
            const col = err.message?.split(":").pop()?.length ?? 0;

            diagnostics.push({
                severity: 0,
                message: "Syntax error: Token declaration out of scope.",
                range: {
                    start: { line: line, character: 0 },
                    end: { line: line, character: col + 1 }
                },
                source: "parser"
            });
        })
    }

    if (diagnostics.length > 0) {
        return diagnostics;
    }
    return astOutput;
}