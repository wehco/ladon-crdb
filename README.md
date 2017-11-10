# ladon-crdb
CockroachDB support for the ladon library

# To Use

IMPORTANT: The migration tool used by Ladon does not currently support
CockroachDB, therefore some special care must be taken by creating the
initial "ladon" database/schema manually.

This will be fixed with the merge of https://github.com/go-gorp/gorp/pull/359

```go
import (
  "fmt"

  "github.com/jmoiron/sqlx"
  manager "github.com/ory/ladon/manager/sql"

  _ "github.com/wehco/ladon-crdb"
)

func main() {
  schema := "ladon"
  migrations_table := "ladon_migrations"

  crdb, _ := sqlx.Open(
    "cockroachdb",
    fmt.Sprintf("postgres://...@.../%s", schema),
  )

  if _, err := crdb.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s;", schema)); err != nil {
    panic(err)
  }

  crdbManager := manager.NewSQLManager(crdb, nil)
  _, err := crdbManager.CreateSchemas("", migrations_table)

  warden := ladon.Ladon{
    Manager: crdbManager,
  }
}
```


