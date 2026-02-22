# lib-database

database library using [pgx/v5](https://github.com/jackc/pgx) and [go-sqlbuilder](https://github.com/huandu/go-sqlbuilder). provides connection pooling, transactions, and query helpers with generics.

## install

```bash
go get github.com/kitti12911/lib-database
```

## config

```yaml
database:
  modsdo:
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

## connect

```go
db, err := database.New(ctx, cfg.Database["modsdo"])
if err != nil {
    log.Fatal(err)
}
defer db.Close()
```

## query helpers

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

## transactions

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

## struct scanning

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

