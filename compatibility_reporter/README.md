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

## How to create bnf file from mysql yacc file

### Create syntax bnf

1. find `sql_yacc.yy` file

    download mysql source code, and `sql_yacc.yy` should be in path `sql/sql_yacc.yy`.

2. get BNF-like rules through `bison`

    run `bison -v sql_yacc.yy`, we can get a file named `sql_yacc.output`. 
    find "Grammar" section in file `sql_yacc.output`.
    now we see many BNF-like rules here, remove line number and save as a text file.
     
    So we get the syntax bnf file like this.

### Create lexical bnf

1. find `lex.h`

    download mysql source code, and `lex.h` should be in path `sql/lex.h`.

2. get BNF-like rules through `lex.h`
    
    find a const named `SYMBOL` in `lex.h`, we get some code like this:
    
    ```C
        {SYM("&&", AND_AND_SYM)},
        {SYM("<", LT)},
        {SYM("<=", LE)},
        {SYM("<>", NE)},
        {SYM("!=", NE)},
        {SYM("=", EQ)},
    ```
    
    there are lexical rules we need, convert their format into:
    
    ```
    AND_AND_SYM: '&&'
    
    LT: '<'
    
    LE: '<='
    
    NE: '<>'
        | '!='
    
    EQ: '='
    ```
    
    we get the lexical bnf file now.

### Edit custom bnf

There are still some lexical rules that isn't included in `lex.h`, 
so we need add them manually.

## License
Apache 2.0 license. See the [LICENSE](../LICENSE) file for details.