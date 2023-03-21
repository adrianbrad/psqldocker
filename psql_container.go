package psqldocker

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"sync/atomic"
	"time"

	// import for init func.
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

// ensure Container implements the io.Closer interface.
var _ io.Closer = (*Container)(nil)

// Container represents a Docker container
// running a PostgreSQL image.
type Container struct {
	psqlUser         string
	psqlPassword     string
	psqlDBName       string
	psqlInstancePort string
	sqls             []string

	runOptions          *dockertest.RunOptions
	pool                *dockertest.Pool
	poolEndpoint        string
	containerExpiration uint
	dockerContainer     *dockertest.Resource
	hostPort            string
	pingRetryTimeout    time.Duration

	closed atomic.Bool
}

// Port returns the container host port mapped
// to the database running inside it.
func (c *Container) Port() string {
	return c.hostPort
}

// Close removes the Docker container.
func (c *Container) Close() error {
	if c.closed.Swap(true) { // returns true if already closed.
		// make Close() idempotent.
		return nil
	}

	return c.dockerContainer.Close()
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

	container := initContainer(user, password, dbName, options)

	if err := container.start(); err != nil {
		return nil, fmt.Errorf("start container: %w", err)
	}

	return container, nil
}

func initContainer(user,
	password,
	dbName string,
	options options,
) *Container {
	return &Container{
		psqlUser:         user,
		psqlPassword:     password,
		psqlDBName:       dbName,
		psqlInstancePort: options.dbPort,
		sqls:             options.sqls,
		runOptions: &dockertest.RunOptions{
			Name:         options.containerName,
			Cmd:          []string{"-p " + options.dbPort},
			Repository:   "postgres",
			Tag:          options.imageTag,
			ExposedPorts: []string{options.dbPort},
			Env:          envVars(user, password, dbName),
		},
		pool:                options.pool,
		poolEndpoint:        options.poolEndpoint,
		containerExpiration: options.expirationSeconds,
		pingRetryTimeout:    options.pingRetryTimeout,

		dockerContainer: nil,
		hostPort:        "",
	}
}

func (c *Container) start() error {
	pool, err := getPool(c.pool, c.poolEndpoint, c.pingRetryTimeout)
	if err != nil {
		return fmt.Errorf("get pool: %w", err)
	}

	if err := pool.Client.Ping(); err != nil {
		return fmt.Errorf("ping docker server: %w", err)
	}

	c.pool = pool

	res, err := pool.RunWithOptions(
		c.runOptions,
		func(config *docker.HostConfig) {
			config.AutoRemove = true
			config.RestartPolicy = docker.RestartPolicy{Name: "no"}
		},
	)
	if err != nil {
		return fmt.Errorf("start container: %w", err)
	}

	c.dockerContainer = res
	c.hostPort = c.dockerContainer.GetPort(c.psqlInstancePort + "/tcp")

	if err := c.dockerContainer.Expire(c.containerExpiration); err != nil {
		return handleErrWithClose(fmt.Errorf("expire: %w", err), c.dockerContainer)
	}

	if err := pool.Retry(
		func() error {
			return pingDB(
				c.psqlUser,
				c.psqlPassword,
				c.psqlDBName,
				c.hostPort,
			)
		}); err != nil {
		return handleErrWithClose(fmt.Errorf("ping db: %w", err), c.dockerContainer)
	}

	if err := executeSQLs(
		c.psqlUser,
		c.psqlPassword,
		c.psqlDBName,
		c.hostPort,
		c.sqls,
	); err != nil {
		return handleErrWithClose(fmt.Errorf("execute sqls: %w", err), c.dockerContainer)
	}

	return nil
}

func handleErrWithClose(err error, dockerContainer *dockertest.Resource) error {
	if closeErr := dockerContainer.Close(); closeErr != nil {
		return errors.Join(fmt.Errorf("close: %w", closeErr), err)
	}

	return err
}

// ErrWithPoolAndWithPoolEndpoint is returned when both
// WithPool and WithPoolEndpoint options are given to the
// NewContainer constructor.
var ErrWithPoolAndWithPoolEndpoint = errors.New(
	"with pool and with pool endpoint are mutually exclusive",
)

func getPool(
	existingPool *dockertest.Pool,
	poolEndpoint string,
	pingRetryTimeout time.Duration,
) (*dockertest.Pool, error) {
	if existingPool != nil && poolEndpoint != "" {
		return nil, ErrWithPoolAndWithPoolEndpoint
	}

	if existingPool != nil {
		existingPool.MaxWait = pingRetryTimeout

		return existingPool, nil
	}

	newPool, err := dockertest.NewPool(poolEndpoint)
	if err != nil {
		return nil, fmt.Errorf("dockertest new pool%w", err)
	}

	newPool.MaxWait = pingRetryTimeout

	return newPool, nil
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
	db, err := sql.Open("pgx", fmt.Sprintf(
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

	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping: %w", err)
	}

	if err := db.Close(); err != nil {
		return fmt.Errorf("close: %w", err)
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
		"pgx",
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

	defer db.Close()

	for i := range sqls {
		_, err = db.Exec(sqls[i])
		if err != nil {
			return fmt.Errorf("execute sql %d: %w", i, err)
		}
	}

	if err := db.Close(); err != nil {
		return fmt.Errorf("close db: %w", err)
	}

	return nil
}
