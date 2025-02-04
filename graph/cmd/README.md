## CLI design notes

Does FSIL Tool have one CLI and many subcommands?



### Graph CLI

This CLI is for working directly with the graph. Input/output are via
files. The _intention_ is that the internal API will be available for
other workflows.

$ fsil graph --db=local --import FILES
$ fsil graph --db=local --dropall --import FILES
$ fsil graph --db=local --export=TO_FILE
$ fsil graph --db=local --



### Files CLI

This CLI will detect file types, and perhaps also export files: either from
the graph state or the internal state. The _intention_ is that the internal
API will be available for other workflows.

$ fsil file --detect FILES
