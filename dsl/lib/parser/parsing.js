// Copyright 2024 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

import {
    fsilLexer,
    allTokens,
    Simulation,
    Channel,
    Network,
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
            children.networks = $.SUBRULE($.networks);
            children.uses = $.SUBRULE($.uses);
            children.vars = $.SUBRULE($.vars);
            children.models = $.SUBRULE($.models);
            children.stacked_models = $.SUBRULE($.stacked_models);
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
                    channels.push({
                        type: channel.tokenType.name.replace(' ', ''),
                        object: updateTokenObject(channel),
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

        $.RULE("workflow", () => {
            const workflow = $.CONSUME(Workflow);
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
            return {
                type: workflow.tokenType.name.replace(' ', ''),
                object: updateTokenObject(workflow),
                children: {
                    workflow_vars: workflow_vars,
                }
            };
        });

        $.RULE("models", () => {
            const models = [];
            $.MANY({
                DEF: () => {
                    const model = $.CONSUME(Model);
                    const model_channels = $.SUBRULE($.channels);
                    const env_vars = $.SUBRULE($.envvars);
                    const workflow = $.SUBRULE($.workflow);
                    models.push({
                        type: model.tokenType.name,
                        object: updateTokenObject(model),
                        children: {
                            channels: model_channels,
                            env_vars: env_vars,
                            workflow: workflow,
                        }
                    });
                }
            });
            return models;
        });

        $.RULE("stacked_models", () => {
            let stacks = []
            $.MANY({
                DEF: () => {
                    const stack = $.CONSUME(Stack);
                    const stacked_models = $.SUBRULE($.models);
                    stacks.push({
                        type: stack.tokenType.name,
                        object: updateTokenObject(stack),
                        children: {
                            models: stacked_models,
                        }
                    });
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