# go-selectivetesting
Perform selective testing on a Go project based on a list of input files and usage information.

## CLI

### Installation
Run the following command:
```
$ go install github.com/pwnedgod/go-selectivetesting/cmd/selectivetesting@latest
```

### Configuration
The following flag arguments can be passed in to the CLI command.

```
$ selectivetesting -relativepath="../" -prettyoutput -patterns="./..." -depth=10 pkg/package1/changedfile1.go pkg/package1/changedfile2.go pkg/package2/changedfile3.go
```

  - `-analyzeroutpath=<string>`
        Path to output debug information for analyzer.
  - `-basepkg=<string>`
        Base package path/module name, will be used instead of &lt;modulepath&gt;/go.mod.
  - `-buildflags=<string,string,...>`
        Build flags to use.
  - `-cfgpath=<string>`
        Config file to use for command configuration.
  - `-depth=<int>`
        Depth of the test search from input files.
  - `gotestargs=<string>`
        The arguments to pass to the go test command. The arguments will be put at the end of the command.
  - `gotestparallel=int`
        Maximum number of parallel go test processes. If not set, it will run the test in series.
  - `gotestrun`
        Whether to run go test with the result of the output. Will output the testing information instead.
  - `-moduledir=<string>`
        Path to the directory of the module.
  - `-patterns=<string,string,...>`
        Patterns to use for package search.
  - `-prettyoutput`
        Whether to output indented json. Will be ignored if -gotestrun is set.
  - `-relativepath=<string>`
        Relative path from current working directory for input files.

A configuration JSON file can also be passed in instead with `-cfgpath=<string>`.

```json
{
    "relativePath": "../",
    "prettyOutput": true,
    "patterns": ["./..."],
    "moduleDir": ".",
    "depth": 10,
    "buildFlags": ["mycustombuildflag"],
    "analyzerOutPath": "analyzer.json",
    "goTest": {
        "run": true,
        "args": "-race -count 2",
        "parallel": 10
    },
    "miscUsages": [
        {
            "regexp": "^<<basepath>>/migration/.+\\.sql$",
            "usedBy": [
                {
                    "pkgPath": "github.com/pwnedgod/go-selectivetesting/example1/...",
                    "testNames": ["*"]
                },
                {
                    "pkgPath": "github.com/pwnedgod/go-selectivetesting/example2/sub",
                    "testNames": ["TestFunc1", "TestFunc2"]
                }
            ]
        }
    ]
}
```

## Manual Usage
If you instead want to do your own procedures, you can follow these instructions.

### Dependency Installation
```
$ go get github.com/pwnedgod/go-selectivetesting
```
