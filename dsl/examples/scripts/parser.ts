// Copyright 2024 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

import {
    parse
} from "../../lib/parser/parsing";
import {
    readFileSync,
} from 'fs';

interface AST {
    [key: string]: any;
}
const data = readFileSync('../dsl/detailed.dse', 'utf8');
console.log("Parsing ...");
let astOutput: AST = parse(data);
const jsonAst: string = JSON.stringify(astOutput, null, 2);
console.log('Generated AST : \n', jsonAst);
