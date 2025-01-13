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
$ node lexer.js
```

### Parser
```bash
# Test parser.
# Parses and generates an AST object.
$ npx tsx parser.ts
```