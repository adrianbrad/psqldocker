<img align="right" width="300" src="https://github.com/adrianbrad/psqldocker/blob/image-data/psql_docker.png?raw=true" alt="adrianbrad psqldocker">

# 🚢 psqldocker

powered by [`ory/dockertest`](https://github.com/ory/dockertest).

[![made-with-Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](https://go.dev/)
[![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/adrianbrad/psqldocker)](https://github.com/adrianbrad/psqldocker)
[![GoDoc reference example](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/adrianbrad/psqldocker)

[![CodeFactor](https://www.codefactor.io/repository/github/adrianbrad/psqldocker/badge)](https://www.codefactor.io/repository/github/adrianbrad/psqldocker)
[![Go Report Card](https://goreportcard.com/badge/github.com/adrianbrad/psqldocker)](https://goreportcard.com/report/github.com/adrianbrad/psqldocker)
[![lint-test](https://github.com/adrianbrad/psqldocker/workflows/lint-test/badge.svg)](https://github.com/adrianbrad/psqldocker/actions?query=workflow%3Alint-test)
[![codecov](https://codecov.io/gh/adrianbrad/psqldocker/branch/main/graph/badge.svg)](https://codecov.io/gh/adrianbrad/psqldocker)

---

Go package providing lifecycle management for PostgreSQL Docker instances.

Leverage Docker to run unit and integration tests against a real PostgreSQL database.

### Usage
The following code shows how to start and stop a PostgreSQL container in a 
`TestMain` function.
```go
func TestMain(m *testing.M) {
    c, err := psqldocker.NewContainer(
        "user",
        "password",
        "dbName",
        psqldocker.WithContainerName("test"),
        psqldocker.WithSql( //initialize with a schema
        "CREATE TABLE users(user_id UUID PRIMARY KEY);",
        ),
        ...
    )
    if err != nil {
        fmt.Printf("new container: %s", err)
        return
    }
	
    var ret int
	
    defer func() {
        _ = c.Close()
		
        os.Exit(ret)
    }   
	
    ret = m.Run()
}
```