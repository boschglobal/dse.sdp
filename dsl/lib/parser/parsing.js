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
} from "../lexer/lexing.js";
import {
    EmbeddedActionsParser
} from "chevrotain";

class FsilParser extends EmbeddedActionsParser {
    constructor() {
        super(allTokens);
        const $ = this;

        function updateTokenObject(tokenObject) {
            let updatedObj = Object.assign({}, tokenObject);
            delete updatedObj.tokenType;
            return updatedObj;
        }

        // The main rule that parses the entire simulation statement.
        $.RULE("simulation", () => {
            const simulation = $.CONSUME(Simulation);
            const children = {};
            children.channels = $.SUBRULE($.channels);
            children.uses = $.SUBRULE($.uses);
            children.vars = $.SUBRULE($.vars);
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
            const channels = [];
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
            const networks = [];
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
            const uses = [];
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
            const vars = [];
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
            const files = [];
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
            const env_vars = [];
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
            const workflow_vars = [];
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
            const workflows = [];
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
            const models = [];
            
            $.MANY(() => {
                const model = $.CONSUME(Model);
                const model_channels = $.SUBRULE($.channels);
        
                let model_files = null;
                let env_vars = null;
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
            let stack = '';
            $.MANY({
                DEF: () => {
                    stack = $.CONSUME(Stack);
                }
            });
            return stack;
        });

        $.RULE("stacked_models", () => {
            let stacks = []
            $.MANY({
                DEF: () => {
                    const stack = $.SUBRULE($.stack);
                    const env_vars = $.SUBRULE($.envvars);
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
const parserInstance = new FsilParser();
export function parse(inputText) {
    const lines = inputText.split(/\r?\n/);
    let lineStartOffset = 0;
    let lexingResult = {
        'tokens': [],
        "groups": {},
        "errors": []
    };
    lines.forEach((line) => {
        if (line !== '') {
            const res = fsilLexer.tokenize(line);
            lexingResult['tokens'].push(...res['tokens']);
            Object.assign(lexingResult['groups'], res['groups']);
            lexingResult['errors'].push(...res['errors']);
            if (lexingResult.errors.length > 0) {
                console.log(lexingResult.errors);
                throw Error("Lexing errors detected");
            }
            lineStartOffset += line.length + 1;
        }
    });
    parserInstance.input = lexingResult.tokens;
    const astOutput = parserInstance.simulation();
    if (parserInstance.errors.length > 0) {
        throw Error("Parsing errors detected!\n" + parserInstance.errors[0].message);
    }
    return astOutput;
}