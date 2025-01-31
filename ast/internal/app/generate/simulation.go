// Copyright 2025 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

package generate

import (
	"flag"
	"fmt"
	//	"github.com/boschglobal/dse.schemas/code/go/dse/ast"
)

func (c *GenerateCommand) GenerateSimulation() error {

	fmt.Fprintf(flag.CommandLine.Output(), "Writing simulation: %s\n", c.outputPath)

	return nil
}
