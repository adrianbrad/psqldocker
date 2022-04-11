package psqldocker

import (
	"github.com/ory/dockertest/v3"
)

type options struct {
	containerName,
	dbPort string
	sqls []string
	pool *dockertest.Pool
}

func defaultOptions() options {
	return options{
		containerName: "go-psqldocker",
		dbPort:        "5432",
		sqls:          nil,
		pool:          nil,
	}
}

// Option configures an BTC Node Docker.
type Option interface {
	apply(*options)
}

type containerNameOption string

func (c containerNameOption) apply(opts *options) {
	opts.containerName = string(c)
}

// WithContainerName configures the PSQL Container Name, if
// empty, a random one will be picked.
func WithContainerName(name string) Option {
	return containerNameOption(name)
}

type sqlOption string

func (c sqlOption) apply(opts *options) {
	opts.sqls = append(opts.sqls, string(c))
}

// WithSql specifies a sqls file, to initiate the
// db with.
func WithSql(sql string) Option {
	return sqlOption(sql)
}

type dbPortOption string

func (c dbPortOption) apply(opts *options) {
	opts.dbPort = string(c)
}

// WithDBPort sets database port running in the container, default 5432.
func WithDBPort(name string) Option {
	return dbPortOption(name)
}

type poolOption struct {
	p *dockertest.Pool
}

func (p poolOption) apply(opts *options) {
	opts.pool = p.p
}

// WithPool sets the docker container newPool.
func WithPool(pool *dockertest.Pool) Option {
	return poolOption{pool}
}
