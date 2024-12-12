// Copyright 2024 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

import {
    lex
} from "../../lib/lexer/lexing.js";
import {
    readFileSync
} from 'fs';

const data = readFileSync('../dsl/detailed.dse', 'utf8');
const lexingResult = lex(data);
console.log(JSON.stringify(lexingResult, null, "\t"));
