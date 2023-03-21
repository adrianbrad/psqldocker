package psqldocker

import (
	"time"

	"github.com/ory/dockertest/v3"
)

type options struct {
	containerName,
	imageTag,
	poolEndpoint,
	dbPort string
	sqls              []string
	pool              *dockertest.Pool
	expirationSeconds uint
	pingRetryTimeout  time.Duration
}

func defaultOptions() options {
	return options{
		containerName:     "go-psqldocker",
		imageTag:          "alpine",
		poolEndpoint:      "",
		dbPort:            "5432",
		sqls:              nil,
		pool:              nil,
		expirationSeconds: 20,
		pingRetryTimeout:  20 * time.Second,
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

type imageTagOption string

func (t imageTagOption) apply(opts *options) {
	opts.imageTag = string(t)
}

// WithImageTag configures the PSQL Container image tag, default: alpine.
func WithImageTag(tag string) Option {
	return imageTagOption(tag)
}

type sqlOption string

func (c sqlOption) apply(opts *options) {
	opts.sqls = append(opts.sqls, string(c))
}

// WithSQL specifies a sqls file, to initiate the
// db with.
func WithSQL(sql string) Option {
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

// WithPool sets the docker container getPool.
// ! This is mutually exclusive with WithPoolEndpoint, and an error
// will be thrown if both are used.
func WithPool(pool *dockertest.Pool) Option {
	return poolOption{pool}
}

type poolEndpoint struct {
	e string
}

func (p poolEndpoint) apply(opts *options) {
	opts.poolEndpoint = p.e
}

// WithPoolEndpoint sets the docker container pool endpoint.
// ! This is mutually exclusive with WithPool, and an error
// will be thrown if both are used.
func WithPoolEndpoint(endpoint string) Option {
	return poolEndpoint{endpoint}
}

type pingRetryTimeout time.Duration

func (p pingRetryTimeout) apply(opts *options) {
	opts.pingRetryTimeout = time.Duration(p)
}

// WithPingRetryTimeout sets the timeout in seconds
// for the  ping retry function.
func WithPingRetryTimeout(seconds uint) Option {
	return pingRetryTimeout(time.Duration(seconds) * time.Second)
}

type expirationSeconds uint

func (e expirationSeconds) apply(opts *options) {
	opts.expirationSeconds = uint(e)
}

// WithExpiration terminates the container after a period has passed.
func WithExpiration(seconds uint) Option {
	return expirationSeconds(seconds)
}
