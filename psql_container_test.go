package psqldocker_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/adrianbrad/psqldocker"
	"github.com/matryer/is"
	"github.com/ory/dockertest/v3"
)

func TestMain(m *testing.M) {
	c, err := psqldocker.NewContainer(
		"a",
		"a",
		"a",
		psqldocker.WithContainerName("test"),
		psqldocker.WithSql(
			"CREATE TABLE users(user_id UUID PRIMARY KEY);",
		),
	)
	if err != nil {
		fmt.Printf("new container error: %s\n", err)
		return
	}

	var ret int

	defer func() {
		_ = c.Close()

		os.Exit(ret)
	}()

	m.Run()
}

func TestNewContainer(t *testing.T) {
	t.Parallel()

	const (
		user     = "user"
		password = "pass"
		dbName   = "test"
	)

	t.Run("AllOptions", func(t *testing.T) {
		t.Parallel()

		i := is.New(t)

		p, err := dockertest.NewPool("")
		i.NoErr(err)

		c, err := psqldocker.NewContainer(
			user,
			password,
			dbName,
			psqldocker.WithContainerName("test"),
			psqldocker.WithDBPort("5432"),
			psqldocker.WithPool(p),
			psqldocker.WithSql(
				"CREATE TABLE users(user_id UUID PRIMARY KEY);",
			),
		)
		i.NoErr(err)

		err = c.Close()
		i.NoErr(err)
	})

	t.Run("NoOptions", func(t *testing.T) {
		t.Parallel()

		i := is.New(t)

		c, err := psqldocker.NewContainer(
			user,
			password,
			dbName,
		)
		i.NoErr(err)

		err = c.Close()
		i.NoErr(err)
	})
}
