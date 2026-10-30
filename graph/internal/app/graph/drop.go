package graph

import (
	"errors"
	"fmt"
	"strings"

	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
)

type GraphDropCommand struct {
	base
	optAll bool
}

var dropTargets = map[string]string{"ast": "Ast", "sim": "Sim", "--all": "all"}

func NewGraphDropCommand(name string) *GraphDropCommand {
	c := &GraphDropCommand{}
	c.init(name)
	c.FlagSet().BoolVar(&c.optAll, "all", false, "drop nodes based on label from the database (Usage: drop ast, drop sim, drop --all)")
	return c
}

func (c *GraphDropCommand) Run() error {
	c.setupLogging()
	option, err := c.dropOption()
	if err != nil {
		return err
	}

	ctx, _, closeFn, err := c.connect()
	if err != nil {
		return err
	}
	defer closeFn()

	fmt.Printf("Graph query: MATCH DETACH DELETE %s nodes\n", dropTargets[option])
	return graph.Drop(ctx, option)
}

func (c *GraphDropCommand) dropOption() (string, error) {
	if c.optAll {
		return "--all", nil
	}
	if args := c.FlagSet().Args(); len(args) > 0 {
		// ast/sim are case-insensitive.
		if option := strings.ToLower(args[0]); option == "ast" || option == "sim" {
			return option, nil
		}
	}
	return "", errors.New("invalid usage, use 'graph drop ast', 'graph drop sim', or 'graph drop --all'")
}
