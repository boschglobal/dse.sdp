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
```

## Run
```bash
# parse2ast <input_file> <output_file>
$ parse2ast examples/input.fsil ./out
```


## Example

### Input File
```bash
# Input file.
$ cd dse.sdp/dsl/examples
$ ls
input.fsil
```

### Lexer
```bash
$ cd lexer

# Test lexer.
$ node lexer.js
$ cd ..
```

### Parser
```bash
$ cd parser

# Test parser.
# Parses and generates an AST object.
$ npx tsx parser.ts

# Output file.
$ cd ..
$ cat AST.json
```