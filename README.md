# go-selectivetesting

Perform selective testing on a Go project based on a list of input files and usage information.

## CLI

### Installation

Run the following command:

```
$ go install github.com/ezraisw/go-selectivetesting/cmd/selectivetesting@latest
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
- `-gotestargs=<string>`
  The arguments to pass to the go test command. The arguments will be put at the end of the command.
- `-gotestparallel=int`
  Maximum number of parallel go test processes. If not set, it will run the test in series.
- `-gotestrun`
  Whether to run go test with the result of the output. Will output the testing information instead.
- `-moduledir=<string>`
  Path to the directory of the module.
- `-patterns=<string,string,...>`
  Patterns to use for package search.
- `-prettyoutput`
  Whether to output indented json. Will be ignored if -gotestrun is set.
- `-relativepath=<string>`
  Relative path from current working directory for input files.
- `-testall`
  Override output with list of all packages within its groups.
- `-outputemptygroups`
  Whether to output untested groups as a group with empty arrays. Default group included.

A configuration JSON file can also be passed in instead with `-cfgpath=<string>`.

```json
{
  "relativePath": "../",
  "prettyOutput": true,
  "patterns": ["./..."],
  "moduleDir": ".",
  "depth": 10,
  "buildFlags": ["mycustombuildflag"],
  "testAll": false,
  "analyzerOutPath": "analyzer.json",
  "goTest": {
    "run": true,
    "args": "-race -count 2",
    "parallel": 10
  },
  "groups": [
    {
      "name": "entity-model",
      "patterns": ["github.com/ezraisw/examplerepo/pkg/entity"]
    },
    {
      "name": "httphandler",
      "patterns": ["github.com/ezraisw/examplerepo/pkg/http/handler/..."]
    },
    {
      "name": "repo-external",
      "patterns": [
        "github.com/ezraisw/examplerepo/pkg/grpc",
        "github.com/ezraisw/examplerepo/pkg/repository"
      ]
    },
    {
      "name": "usecase",
      "patterns": ["github.com/ezraisw/examplerepo/pkg/usecase/..."]
    }
  ],
  "outputEmptyGroups": true,
  "miscUsages": [
    {
      "regexp": "^<<basepath>>/migration/.+\\.sql$",
      "usedBy": [
        {
          "pkgPath": "github.com/ezraisw/go-selectivetesting/example1/...",
          "all": true
        },
        {
          "pkgPath": "github.com/ezraisw/go-selectivetesting/example2/sub",
          "objNames": ["FuncUsingNonGoFiles1", "FuncUsingNonGoFiles2"]
        },
        {
          "pkgPath": "github.com/ezraisw/go-selectivetesting/example2/sub",
          "fileNames": ["foo.go", "bar.go"]
        }
      ]
    }
  ]
}
```

### JSON Output

If you choose not to use `-gotestrun`, the application will output a JSON containing all the testing groups.

```json
{
  "uniqueTestCount": 17,
  "groups": [
    {
      "name": "default",
      "testedPkgs": []
    },
    {
      "name": "entity-model",
      "testedPkgs": [
        {
          "pkgPath": "github.com/ezraisw/examplerepo/pkg/entity",
          "relativePkgPath": "./pkg/entity",
          "hasNotable": true,
          "testNames": ["TestNewWishlist", "TestWishlist_Model"],
          "runRegex": "^(TestNewWishlist|TestWishlist_Model)$"
        }
      ]
    },
    {
      "name": "httphandler",
      "testedPkgs": [
        {
          "pkgPath": "github.com/ezraisw/examplerepo/pkg/http/handler/api/v1/wishlist",
          "relativePkgPath": "./pkg/http/handler/api/v1/wishlist",
          "hasNotable": false,
          "testNames": [
            "TestHandler_CreateWishlist",
            "TestHandler_DeleteWishlist"
          ],
          "runRegex": "^(TestHandler_CreateWishlist|TestHandler_DeleteWishlist)$"
        }
      ]
    },
    {
      "name": "repo-external",
      "testedPkgs": [
        {
          "pkgPath": "github.com/ezraisw/examplerepo/pkg/repository",
          "relativePkgPath": "./pkg/repository",
          "hasNotable": false,
          "testNames": [
            "TestWishlist_Create",
            "TestWishlist_Delete",
            "TestWishlist_GetAllByProductSizeSummaryID",
            "TestWishlist_GetByID",
            "TestWishlist_GetByProductSizeSummaryIDAndUserID",
            "TestWishlist_GetByProductTypeAndProductCode",
            "TestWishlist_ListUserWishlists"
          ],
          "runRegex": "^(TestWishlist_Create|TestWishlist_Delete|TestWishlist_GetAllByProductSizeSummaryID|TestWishlist_GetByID|TestWishlist_GetByProductSizeSummaryIDAndUserID|TestWishlist_GetByProductTypeAndProductCode|TestWishlist_ListUserWishlists)$"
        }
      ]
    },
    {
      "name": "usecase",
      "testedPkgs": [
        {
          "pkgPath": "github.com/ezraisw/examplerepo/pkg/usecase/wishlist",
          "relativePkgPath": "./pkg/usecase/wishlist",
          "hasNotable": false,
          "testNames": [
            "TestWishlistUsecase_CreateWishlist",
            "TestWishlistUsecase_DeleteWishlist",
            "TestWishlistUsecase_GetUserWishlistDetail",
            "TestWishlistUsecase_GetWishlist",
            "TestWishlistUsecase_ListUserWishlists",
            "TestWishlistUsecase_ProcessGeneralPriceUpdate"
          ],
          "runRegex": "^(TestWishlistUsecase_CreateWishlist|TestWishlistUsecase_DeleteWishlist|TestWishlistUsecase_GetUserWishlistDetail|TestWishlistUsecase_GetWishlist|TestWishlistUsecase_ListUserWishlists|TestWishlistUsecase_ProcessGeneralPriceUpdate)$"
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
$ go get github.com/ezraisw/go-selectivetesting
```
