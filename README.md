# go-selectivetesting
Perform selective testing on a go project based on the list of input files.

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
        path to output debug information for analyzer
  - `-basepkg=<string>`
        base package path/module name, will be used instead of "&lt;modulepath&gt;/go.mod"
  - `-buildflags=<string,string,...>`
        build flags to use
  - `-cfgpath=<string>`
        config file to use for command configuration
  - `-depth=<int>`
        depth of the test search from input files
  - `-moduledir=<string>`
        path to the directory of the module
  - `-patterns=<string,string,...>`
        patterns to use for package search
  - `-prettyoutput`
        whether to output indented json
  - `-relativepath=<string>`
        relative path from current working directory for input files

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
