// pach_postgres_check.go

package main

import (
    "context"
    "fmt"
    "os"
    "flag"
    "time"

    "github.com/jackc/pgx/v5"
)

func main() {

    connectionString := flag.String("connection_string", "", "postgres database connection string")
    timeout := flag.Float64("timeout", 10, "connection attempt timeout in seconds")
    flag.Parse()

    if connectionString == nil {
        fmt.Fprintf(os.Stderr, "error: database connection string is required with -connection_string")
        os.Exit(1)
    }

    ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeout * float64(time.Second)))
    defer cancel()

    conn, err := pgx.Connect(ctx, *connectionString)
    if err != nil {
        fmt.Fprintf(os.Stderr, "error: database connection failed with connection string \"%s\": %v\n", *connectionString, err)
        os.Exit(2)
    }
    conn.Close(ctx)

}
