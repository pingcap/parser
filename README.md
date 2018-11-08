# parser

TiDB SQL Parser

## How to update parser for TiDB

Assuming that you want to file a PR (pull request) to TiDB, and your PR includes a change in the parser, implement the following steps to update the parser in TiDB.

### Step 1, make changes in your parser repository

Fork this repository to your own account and commit the changes to your repository.

* Don't forget to run `make test` before you commit!
* Make sure the `parser.go` is updated.

Suppose the forked repository is `https://github.com/your-repo/parser`.

### Step 2: Make your parser changes take effect in TiDB and run CI

In your TiDB repository, modify the `go.mod` file, remove `github.com/pingcap/parser` from the `require` instruction, add a new line at the end of file like this:

```
replace github.com/pingcap/parser => github.com/your-repo/parser v0.0.0-20181102150703-4acd198f5092
```

This change tells TiDB to use the modified parser from your repository.

You can get correct version information by running this command in your TiDB directory:

```
GO111MODULE=on go get -u github.com/your-repo/parser@master
```

If you get some error output, never mind. You can still edit the `go.mod` file manually.

Pull a request to TiDB.

### Step 3, merge the parser to this repository

Pull a request to this repository. **Link the related PR in TiDB in your PR description or comment**.

The PR will be reviewed, if everything goes well, it will be merged.

### Step 4, update TiDB to use the latest parser

In your TiDB pull request, modify the `go.mod` file manually or use this command:

```
GO111MODULE=on go get -u github.com/pingcap/parser@master
```

Make sure the `replace` instruction is changed back to the `require` instruction and the version is latest.
