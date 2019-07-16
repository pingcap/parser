## compatibility_reporter

compatibility_reporter is a tool for generating and testing the compatibility of sql case between MySQL and TiDB.

## How to use

```
Usage of compatibility_reporter:
  -N uint
        The number of output sql case, set 0 for infinite
  -P int
        Port number to use for MySQL connection (default 3306)
  -R    Generator SQL randomly
  -a    Output all test case, regardless of success or failure
  -b string
        BNF file path
  -h string
        Connect to MySQL host (default "127.0.0.1")
  -n string
        Production name to test
  -o string
        Output path of csv format report (default "./report.csv")
  -p string
        Password to use when connecting to MySQL server
  -u string
        MySQL User for login (default "root")
```

## Example

```
./compatibility_reporter  -p 123456 -b "../compatibility_reporter/mysql80_bnf.txt ../compatibility_reporter/mysql80_lexical.txt ../compatibility_reporter/mysql80_custom.txt" -n alter_table_stmt -R -N 1000
```

## License
Apache 2.0 license. See the [LICENSE](../LICENSE) file for details.