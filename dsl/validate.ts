#!/usr/bin/env node

import { readFileSync, writeFileSync } from "node:fs";
import { parseArgs } from "node:util";
import { parse } from "./lib/parser/parsing";

const usage = `Usage: dse-lang-compile [--format text|json] [--ast output.json|-] <script.dse|->

Check DSE Script syntax and construct its AST without resolving or running tasks.
Use '-' to read from stdin or write the AST to stdout. --ast writes the AST only when validation succeeds.
Exit codes: 0 valid, 1 invalid script, 2 usage or tool failure.
JSON diagnostic ranges use zero-based lines and columns; text ranges are one-based.
Both formats use an exclusive range end.
`;

function main(): number {
  const args = process.argv.slice(2);
  let optionsEnded = false;
  for (let index = 0; index < args.length; index += 1) {
    if (args[index] === "--") {
      optionsEnded = true;
      continue;
    }
    if (!optionsEnded && args[index] === "--ast" &&
      (index + 1 === args.length || args[index + 1] === "--" || args[index + 1].startsWith("--"))) {
      args[index] = "--ast=-";
    }
  }
  let json = args.some((argument, index, argumentsList) =>
    argument === "--format=json" ||
    (argument === "json" && argumentsList[index - 1] === "--format"));
  try {
    const { values, positionals } = parseArgs({
      args,
      options: {
        format: { type: "string", default: "text" },
        ast: { type: "string" },
        help: { type: "boolean", short: "h" },
      },
      allowPositionals: true,
    });
    if (values.help) {
      process.stdout.write(usage);
      return 0;
    }
    json = values.format === "json";
    if (!["text", "json"].includes(values.format!) || positionals.length !== 1) {
      throw new Error(usage.trim());
    }
    const file = positionals[0];
    const input = readFileSync(file === "-" ? 0 : file, "utf8");
    const result = parse(input);
    const diagnostics = Array.isArray(result)
      ? result.map(diagnostic => ({ ...diagnostic, file }))
      : [];
    const valid = diagnostics.length === 0;
    const displayFile = file === "-" ? "stdin" : file;
    if (valid && values.ast) {
      const astText = JSON.stringify(result, null, 2) + "\n";
      if (values.ast === "-") {
        process.stdout.write(astText);
      } else {
        writeFileSync(values.ast, astText, "utf8");
      }
    }
    if (json) {
      if (values.ast !== "-") {
        process.stdout.write(JSON.stringify({ valid, diagnostics }) + "\n");
      }
    } else if (valid) {
      if (!values.ast) {
        process.stdout.write(`Validation passed: ${displayFile}\n`);
      }
    } else {
      const sourceLines = input.split(/\r?\n/);
      for (const diagnostic of diagnostics) {
        const { start, end } = diagnostic.range;
        const sourceLine = sourceLines[start.line] ?? "";
        const startColumn = Math.max(0, Math.min(start.character, sourceLine.length));
        const endColumn = Math.max(startColumn, Math.min(
          end.line === start.line ? end.character : sourceLine.length,
          sourceLine.length,
        ));
        const message = diagnostic.code === "DSE1002" && diagnostic.message.startsWith("Unexpected or malformed statement:")
          ? "Unexpected or malformed statement"
          : diagnostic.message;
        const code = diagnostic.code ? `[${diagnostic.code}]` : "";
        const endLocation = end.line === start.line
          ? `${end.character + 1}`
          : `${end.line + 1}:${end.character + 1}`;

        process.stderr.write(`${displayFile}:${start.line + 1}:${start.character + 1}-${endLocation}: error${code}: ${message}\n`);
        process.stderr.write(`  | ${sourceLine}\n`);
        process.stderr.write(`  | ${sourceLine.slice(0, startColumn).replace(/[^\t]/g, " ")}${"^".repeat(Math.max(1, endColumn - startColumn))}\n`);
        for (const syntax of diagnostic.expectedSyntax) {
          process.stderr.write(`  = expected: ${syntax}\n`);
        }
        process.stderr.write("\n");
      }
      process.stderr.write(`Found ${diagnostics.length} error${diagnostics.length === 1 ? "" : "s"}.\n`);
    }
    return valid ? 0 : 1;
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    if (json) {
      process.stdout.write(JSON.stringify({ valid: null, diagnostics: [], error: { code: "DSE0001", message } }) + "\n");
    } else {
      process.stderr.write(`dse-lang-compile: ${message}\n`);
    }
    return 2;
  }
}

process.exitCode = main();