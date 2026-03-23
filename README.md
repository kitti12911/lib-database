# lib-database

database library using [pgx/v5](https://github.com/jackc/pgx) and [go-sqlbuilder](https://github.com/huandu/go-sqlbuilder). provides connection pooling, transactions, and query helpers with generics. supports PostgreSQL (pgx), SQL Server ([go-mssqldb](https://github.com/microsoft/go-mssqldb)), and Oracle ([go-ora](https://github.com/sijms/go-ora)).

## install

```bash
go get github.com/kitti12911/lib-database
```

## postgresql

uses [pgx/v5](https://github.com/jackc/pgx) directly (not `database/sql`).

```yaml
database:
  example_db:
    host: "localhost"
    port: "5432"
    user: "postgres"
    password: "secret"
    database: "mydb"
    pool:
      maxConns: 40
      minConns: 20
      maxConnLifeTime: "6h"
      maxConnIdleTime: "1h"
```

in your service config struct:

```go
type Config struct {
    Database map[string]database.Config `mapstructure:"database"`
}
```

```go
import "github.com/kitti12911/lib-database"

db, err := database.New(ctx, cfg.Database["example_db"])
if err != nil {
    log.Fatal(err)
}
defer db.Close()
```

### query helpers

all helpers check context for an active transaction automatically.

```go
// returns (*T, nil) or (nil, nil) when not found
user, err := database.FindOne[User](ctx, db, "SELECT id, name FROM users WHERE id = $1", 42)

// returns []T (empty slice if no rows)
users, err := database.FindAll[User](ctx, db, "SELECT id, name FROM users WHERE status = $1", 1)

// returns rows affected
n, err := database.Exec(ctx, db, "DELETE FROM users WHERE status = $1", 0)
```

### with go-sqlbuilder

use `sqlbuilder.PostgreSQL` flavor for correct `$1, $2, ...` placeholders:

```go
sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
sb.Select("id", "name", "email").From("users")
sb.Where(sb.Equal("status", 1))
sb.OrderByDesc("created_at")
sb.Limit(10)

users, err := database.FindAllB[User](ctx, db, sb)
```

### pagination

pass builder without LIMIT/OFFSET. `FindAndCountB` handles it:

```go
sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
sb.Select("id", "name").From("users")
sb.Where(sb.Equal("status", 1))
sb.OrderByDesc("created_at")

users, total, err := database.FindAndCountB[User](ctx, db, sb, 20, 0) // limit=20, offset=0
```

### returning

use `FindOneB` / `FindAllB` with RETURNING:

```go
ib := sqlbuilder.PostgreSQL.NewInsertBuilder()
ib.InsertInto("users")
ib.Cols("name", "email")
ib.Values("Alice", "alice@example.com")
ib.SQL("RETURNING id, created_at")

result, err := database.FindOneB[CreateResult](ctx, db, ib)
```

### transactions

```go
err := db.Transaction(ctx, func(ctx context.Context) error {
    _, err := database.ExecB(ctx, db, insertBuilder)
    if err != nil {
        return err // rollback
    }

    _, err = database.ExecB(ctx, db, updateBuilder)
    return err // nil = commit
})
```

nested calls reuse the existing transaction:

```go
err := db.Transaction(ctx, func(ctx context.Context) error {
    return db.Transaction(ctx, func(ctx context.Context) error {
        return doWork(ctx)
    })
})
```

### struct scanning

pgx v5 maps columns to struct fields using the `db` tag:

```go
type User struct {
    ID        int64     `db:"id"`
    Name      string    `db:"name"`
    Email     *string   `db:"email"`
    Status    int       `db:"status"`
    CreatedAt time.Time `db:"created_at"`
    UpdatedAt time.Time `db:"updated_at"`
}
```

## sql server

uses `database/sql` with [go-mssqldb](https://github.com/microsoft/go-mssqldb) via the `dbsql` layer.

```yaml
database:
  example_db:
    host: "localhost"
    port: "1433"
    user: "sa"
    password: "secret"
    database: "mydb"
    encrypt: "disable"
    trust_server_certificate: true
    pool:
      max_open_conns: 40
      max_idle_conns: 20
      conn_max_lifetime: "6h"
      conn_max_idle_time: "1h"
```

```go
import "github.com/kitti12911/lib-database/mssql"

db, err := mssql.New(ctx, cfg.Database["example_db"])
if err != nil {
    log.Fatal(err)
}
defer db.Close()
```

## oracle

uses `database/sql` with [go-ora](https://github.com/sijms/go-ora) via the `dbsql` layer.

```yaml
database:
  example_db:
    host: "localhost"
    port: "1521"
    service_name: "ORCLPDB1"
    user: "system"
    password: "secret"
    pool:
      max_open_conns: 40
      max_idle_conns: 20
      conn_max_lifetime: "6h"
      conn_max_idle_time: "1h"
```

connect using SID instead of service name:

```yaml
database:
  example_db:
    host: "localhost"
    port: "1521"
    sid: "ORCL"
    user: "system"
    password: "secret"
```

ssl with oracle wallet:

```yaml
database:
  example_db:
    host: "localhost"
    port: "2484"
    service_name: "ORCLPDB1"
    user: "system"
    password: "secret"
    ssl: true
    ssl_verify: false
    wallet: "/path/to/wallet"
```

```go
import "github.com/kitti12911/lib-database/oracle"

db, err := oracle.New(ctx, cfg.Database["example_db"])
if err != nil {
    log.Fatal(err)
}
defer db.Close()
```

## dbsql shared features

sql server and oracle both use the `dbsql` layer. query helpers, transactions, and struct scanning work the same for both. use `sqlbuilder.SQLServer` or `sqlbuilder.Oracle` flavor for correct bind variable placeholders.

```go
user, err := dbsql.FindOne[User](ctx, db, "SELECT id, name FROM users WHERE id = :1", 42)

users, err := dbsql.FindAll[User](ctx, db, "SELECT id, name FROM users WHERE status = :1", 1)

n, err := dbsql.Exec(ctx, db, "DELETE FROM users WHERE status = :1", 0)
```

### dbsql with go-sqlbuilder

```go
sb := sqlbuilder.Oracle.NewSelectBuilder()
sb.Select("id", "name", "email").From("users")
sb.Where(sb.Equal("status", 1))
sb.OrderByDesc("created_at")

users, err := dbsql.FindAllB[User](ctx, db, sb)
```

### dbsql pagination

```go
sb := sqlbuilder.Oracle.NewSelectBuilder()
sb.Select("id", "name").From("users")
sb.Where(sb.Equal("status", 1))
sb.OrderByDesc("created_at")

users, total, err := dbsql.FindAndCountB[User](ctx, db, sb, 20, 0)
```

### dbsql transactions

```go
err := db.Transaction(ctx, func(ctx context.Context) error {
    _, err := dbsql.ExecB(ctx, db, insertBuilder)
    if err != nil {
        return err // rollback
    }

    _, err = dbsql.ExecB(ctx, db, updateBuilder)
    return err // nil = commit
})
```

### dbsql struct scanning

the `dbsql` layer uses the same `db` tag as pgx for struct scanning:

```go
type User struct {
    ID        int64     `db:"id"`
    Name      string    `db:"name"`
    Email     *string   `db:"email"`
    Status    int       `db:"status"`
    CreatedAt time.Time `db:"created_at"`
    UpdatedAt time.Time `db:"updated_at"`
}
```
