#!/usr/bin/env node

// Copyright 2024 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

import {
    parse
} from "../lib/parser/parsing";
import {
    readFileSync,
    writeFileSync
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
const jsonAst = JSON.stringify(astOutput, null, 2);

// Write the generated AST.
console.log("Writing to file: %s", outputFile);
writeFileSync(outputFile, jsonAst, 'utf8');

// Done.
process.exit(0);
