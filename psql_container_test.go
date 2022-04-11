package psqldocker_test

import (
	"testing"

	"github.com/adrianbrad/psqlutil/psqldocker"
	"github.com/matryer/is"
	"github.com/ory/dockertest/v3"
)

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
