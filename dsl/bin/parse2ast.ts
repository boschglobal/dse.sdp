#!/usr/bin/env node

// Copyright 2024 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

import {
    parse
} from "../lib/parser/parsing";
import {
    readFileSync,
    createWriteStream
} from 'fs';

const cliName = "parse2ast"
const version = "<devel>"

function printUsage() {
    console.log(cliName);
    console.log("-".repeat(cliName.length));
    console.log("usage: %s <input_file> <output_file>", cliName);
}

// Process CLI options.
if (process.argv.length != 4) {
    printUsage();
    process.exit(1);
}
// console.time("Total execution time");
const inputFile = process.argv[2];
const outputFile = process.argv[3];
console.log(`\
${cliName}
${"-".repeat(cliName.length)}
Version: ${version}
Parameters:
  input_file = ${inputFile}
  output_file = ${outputFile}`);

// Parse and generate AST.
console.log("Read from file: %s", inputFile);
const data = readFileSync(inputFile, 'utf8');
console.log("Parsing ...");
let astOutput = parse(data);
// console.log(JSON.stringify(astOutput, null, 2));

if (Array.isArray(astOutput)) { // Error response
    for (const item of astOutput) {
        const line = item?.range?.start?.line;
        const message = item?.message;

        if (typeof line === "number" && typeof message === "string") {
            console.log(`Error found on line ${line + 1}: ${message}`);
        } else {
            console.warn("Invalid diagnostic item:", item);
        }
    }
    process.exit(1);
} else if (astOutput !== null && typeof astOutput === 'object') { // valid ast object
    const stacks = astOutput.children?.stacks ?? [];
    for (const stack of stacks) {
        const stackName = stack.name ?? "";
        console.log(`stack: ${stackName}`);

        const models = stack.children?.models ?? [];
        for (const model of models) {
            const modelName = model.object.payload?.model_name?.value ?? "";
            const repoName = model.object.payload?.model_repo_name?.value ?? "";
            console.log(`model: ${modelName} (${repoName})`);
        }
    }
    console.log("Writing to file: %s", outputFile);
    const out = createWriteStream(outputFile, { encoding: 'utf8' });
    out.on('error', (err) => {
        console.error("Error writing file:", err);
        process.exit(1);
    });
    out.write(JSON.stringify(astOutput));
    out.end();
    out.on('finish', () => {
        // console.timeEnd("Total execution time");
        process.exit(0);
    });
}

