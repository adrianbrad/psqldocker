package psqldocker

import (
	"database/sql"
	"errors"
	"fmt"
	"io"

	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

// ensure Container implements the io.Closer interface.
var _ io.Closer = (*Container)(nil)

// Container represents a Docker container
// running a PostgreSQL image.
type Container struct {
	res  *dockertest.Resource
	port string
}

// Port returns the container host port mapped
// to the database running inside it.
func (c Container) Port() string {
	return c.port
}

// Close removes the Docker container.
func (c Container) Close() error {
	return c.res.Close()
}

// NewContainer starts a new psql database in a docker container.
func NewContainer(
	user,
	password,
	dbName string,
	opts ...Option,
) (*Container, error) {
	options := defaultOptions()

	for i := range opts {
		opts[i].apply(&options)
	}

	pool, err := newPool(options)
	if err != nil {
		return nil, fmt.Errorf("new pool: %w", err)
	}

	// create run options
	dockerRunOptions := &dockertest.RunOptions{
		Name:         options.containerName,
		Cmd:          []string{"-p " + options.dbPort},
		Repository:   "postgres",
		Tag:          options.imageTag,
		ExposedPorts: []string{options.dbPort},
		Env:          envVars(user, password, dbName),
	}

	res, err := startContainer(
		pool,
		dockerRunOptions,
	)
	if err != nil {
		return nil, fmt.Errorf("start container: %w", err)
	}

	// set expiration
	_ = res.Expire(options.expirationSeconds)

	hostPort := res.GetPort(options.dbPort + "/tcp")

	err = pool.Retry(
		func() error {
			return pingDB(
				user,
				password,
				dbName,
				hostPort,
			)
		})
	if err != nil {
		_ = res.Close()

		return nil, fmt.Errorf("ping node: %w", err)
	}

	err = executeSQLs(
		user,
		password,
		dbName,
		hostPort,
		options.sqls,
	)
	if err != nil {
		_ = res.Close()

		return nil, fmt.Errorf("execute sqls: %w", err)
	}

	return &Container{
		res:  res,
		port: hostPort,
	}, nil
}

func startContainer(
	pool *dockertest.Pool,
	runOptions *dockertest.RunOptions,
) (*dockertest.Resource, error) {
	res, err := pool.RunWithOptions(
		runOptions,
		func(config *docker.HostConfig) {
			config.AutoRemove = true
			config.RestartPolicy = docker.RestartPolicy{
				Name: "no",
			}
		},
	)
	if err != nil {
		return nil, fmt.Errorf("docker run%w", err)
	}

	return res, nil
}

// ErrWithPoolAndWithPoolEndpoint is returned when both
// WithPool and WithPoolEndpoint options are given to the
// NewContainer constructor.
var ErrWithPoolAndWithPoolEndpoint = errors.New(
	"with pool and with pool endpoint are mutually exclusive",
)

func newPool(opts options) (*dockertest.Pool, error) {
	if opts.pool != nil && opts.poolEndpoint != "" {
		return nil, ErrWithPoolAndWithPoolEndpoint
	}

	if opts.pool != nil {
		opts.pool.MaxWait = opts.pingRetryTimeout

		return opts.pool, nil
	}

	pool, err := dockertest.NewPool(opts.poolEndpoint)
	if err != nil {
		return nil, fmt.Errorf("dockertest new pool%w", err)
	}

	pool.MaxWait = opts.pingRetryTimeout

	return pool, nil
}

func envVars(
	user,
	password,
	dbName string,
) []string {
	return []string{
		fmt.Sprintf("POSTGRES_PASSWORD=%s", password),
		fmt.Sprintf("POSTGRES_USER=%s", user),
		fmt.Sprintf("POSTGRES_DB=%s", dbName),
	}
}

func pingDB(
	user,
	password,
	dbName,
	port string,
) error {
	db, err := sql.Open("postgres", fmt.Sprintf(
		"user=%s "+
			"password=%s "+
			"dbname=%s "+
			"host=localhost "+
			"port=%s "+
			"sslmode=disable",
		user,
		password,
		dbName,
		port))
	if err != nil {
		return fmt.Errorf("sql open: %w", err)
	}

	defer func() {
		_ = db.Close()
	}()

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("ping: %w", err)
	}

	return nil
}

func executeSQLs(
	user,
	password,
	dbName,
	hostPort string,
	sqls []string,
) error {
	db, err := sql.Open(
		"postgres",
		fmt.Sprintf(
			"user=%s "+
				"password=%s "+
				"dbname=%s "+
				"host=localhost "+
				"port=%s "+
				"sslmode=disable",
			user,
			password,
			dbName,
			hostPort),
	)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}

	defer func() {
		_ = db.Close()
	}()

	for i := range sqls {
		_, err = db.Exec(sqls[i])
		if err != nil {
			return fmt.Errorf("execute sql %d: %w", i, err)
		}
	}

	return nil
}
