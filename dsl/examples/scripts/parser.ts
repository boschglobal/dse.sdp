// Copyright 2024 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

import {
    parse
} from "../../lib/parser/parsing.js";
import {
    readFileSync,
    writeFileSync
} from 'fs';

interface AST {
    [key: string]: any;
}
const data = fs.readFileSync('../input.fsil', 'utf8');
let astOutput: AST = parse(data);
console.log(astOutput);
const jsonAst: string = JSON.stringify(astOutput, null, 2);
fs.writeFileSync('../../out/AST.json', jsonAst, 'utf8');
console.log('File generated : dsl/out/AST.json');
