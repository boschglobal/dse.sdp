# DSL AST Generation

## Setup
```bash
# Node Setup.
# installs nvm (Node Version Manager)
$ curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.0/install.sh | bash

# download and install Node.js (you may need to restart the terminal)
$ nvm install 22

# verifies the right Node.js version is in the environment
$ node -v

# verifies the right npm version is in the environment
$ npm -v

# For running typeScript files
$ npm install -g tsx
```

## Build
```bash
# Get the repo.
$ git clone https://github.boschdevcloud.com/fsil/dse.sdp.git
$ cd dse.sdp/dsl

# Build.
$ make

# Run tests.
$ make test
```

## Run
```bash
# parse2ast <input_file> <output_file>
$ parse2ast examples/dsl/detailed.dse AST.json

```


## Validate Scripts

The check-only compiler frontend parses DSE Script and constructs its AST. It
does not resolve repositories, download metadata, check task parameters, or run
tasks. Success means the script can be parsed into an AST, not that its external
references or simulation are valid.

```bash
# From the dsl directory, install dependencies and build the executable.
cd dsl
make
make install

# From the repository root, check a script (text output is the default).
dse-lang-compile examples/dsl/detailed.dse

# Return machine-readable diagnostics.
dse-lang-compile --format json examples/dsl/detailed.dse

# Write the AST only when parsing succeeds.
dse-lang-compile --ast AST.json examples/dsl/detailed.dse

# Write the AST to stdout.
dse-lang-compile --ast - examples/dsl/detailed.dse

# Read an agent-generated script from stdin.
cat generated.dse | dse-lang-compile --format json -

# Show command syntax and exit codes.
dse-lang-compile --help
```

## Docker Image

Run the default example. `DOCKER_SCRIPT` is relative to the repository root:

```bash
# Run the image directly. Mount the repository so the script path exists in the container.
docker run --rm --volume "$PWD:/workspace:ro" --workdir /workspace dse-lang-compile:latest --format json examples/openloop/openloop.dse

# JSON diagnostics from stdin.
cat generated.dse | docker run --rm -i dse-lang-compile:latest --format json -
```

`docker-run` mounts the repository read-only by default. Override
`DOCKER_MOUNT_MODE=rw` when using `--ast` to write an output file. The image runs
as a non-root user and contains only Node.js and the bundled compiler.

The container's `compile` command writes the AST to stdout by default. Pass
`--out output.json` to write it to a file instead. `--ast` remains accepted as
an alias.

Text diagnostics show one-based start and exclusive-end positions, followed by
the source line and a caret range. Same-line spans use `file:line:start-end`;
multi-line spans include the end line.
JSON output is a single object with `valid` and `diagnostics`; each diagnostic
contains `file`, `code`, `severity` (`0` means error), `message`, `source`, and
`range`. Parser diagnostics may also include `expected`, an array of alternative
token sequences; each sequence lists token names in order, and `EOF` means no
additional input is expected. Every script diagnostic includes `expectedSyntax`,
one or more concrete DSE statement forms relevant to the error. For example, a
malformed variable statement reports `var <name> [uses|network|var] <value> [mimetype|signal]`.
JSON ranges use zero-based lines and columns with an exclusive end.
JSON diagnostics go to stdout without progress logs. Text errors go to stderr.

Exit codes are `0` for a valid script, `1` for script errors, and `2` for usage,
I/O, or unexpected tool failures. For tool failures JSON has `valid: null`, an
empty `diagnostics` array, and an `error` object with code `DSE0001` and a message.
Script error codes are `DSE1001` (lexing), `DSE1002` (grammar), and `DSE1003`
(workflow scope). Invalid scripts do not create or overwrite the requested AST.
The validator reports multiple independent malformed statements in one run when
it can safely recover, and diagnostics are returned in source order. Text inside
embedded `-- filename.ext --` file-content blocks is not treated as a DSE statement.

An agent can generate a script, run this command, use the diagnostic positions
and messages to revise it, and repeat until exit code `0`. A tool failure should
be handled separately rather than treated as feedback about the script.

## Example

### Input File
```bash
# Input file.
$ ls dse.sdp/dsl/examples/dsl
detailed.dse
```

### Lexer
```bash
$ cd dse.sdp/dsl/examples/scripts

# Test lexer.
$ npx tsx lexer.ts
```

### Parser
```bash
# Test parser.
# Parses and generates an AST object.
$ npx tsx parser.ts
```