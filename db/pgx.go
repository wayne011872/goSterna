package db

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wayne011872/goSterna/util"
)

const (
	CtxPgxKey = util.CtxKey("ctxPgxKey")
)

func GetPgxClientByGin(c *gin.Context) PgxPoolClient {
	cltInter ,_:= c.Get(string(CtxPgxKey))
	if dbclt, ok := cltInter.(PgxPoolClient); ok {
		return dbclt
	}
	return nil
}

func GetPgxClientByReq(req *http.Request) PgxPoolClient {
	return GetPgxClientByCtx(req.Context())
}

func GetPgxClientByCtx(ctx context.Context) PgxPoolClient {
	cltInter := ctx.Value(CtxPgxKey)
	if dbclt, ok := cltInter.(PgxPoolClient); ok {
		return dbclt
	}
	return nil
}

type PgxDI interface {
	NewPgxClient(ctx context.Context) (PgxPoolClient, error)
	SetAuth(user, pwd string)
	GetUri() string
}

type PgxConf struct {
	Uri       string `yaml:"uri"`
	User      string `yaml:"user"`
	Pass      string `yaml:"pass"`
	DefaultDB string `yaml:"defaul"`
	SslMode   string `yaml:"sslmode"`

	authUri  string	
}

func (pc *PgxConf) SetAuth(user, pwd string) {
	pc.authUri = strings.Replace(pc.Uri, "{User}", user, 1)
	pc.authUri = strings.Replace(pc.authUri, "{Pwd}", pwd, 1)
}

func (pc *PgxConf) GetUri() string {
	if pc.Uri == "" {
		panic("Postgresql uri not set")
	}
	if pc.DefaultDB == "" {
		panic("Postgresql default db not set")
	}
	return pc.Uri + "/" + pc.DefaultDB + "?sslmode="+pc.SslMode
}
func (pc *PgxConf) GetAuthUri() string {
	if pc.Uri == "" {
		panic("Postgresql uri not set")
	}
	if pc.DefaultDB == "" {
		panic("Postgresql default db not set")
	}
	pc.SetAuth(pc.User, pc.Pass)
	return pc.authUri+ "/" + pc.DefaultDB + "?sslmode="+pc.SslMode
}
func (pc *PgxConf) GetPgxConfig() (*pgxpool.Config) {
	var uri string
	if pc.User == "" || pc.Pass == "" {
		uri = pc.GetUri()
	}else{
		uri = pc.GetAuthUri()
	}
	config, err := pgxpool.ParseConfig(uri)
	if err != nil{
		panic(err)
	}
	config.MaxConns = int32(4)
	config.MinConns = int32(0)
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = time.Minute * 30
	config.HealthCheckPeriod = time.Minute
	config.ConnConfig.ConnectTimeout = time.Second * 5
	return config
}

func (pc *PgxConf) NewPgxClient(ctx context.Context) (PgxPoolClient, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	pgxPool, err := pgxpool.NewWithConfig(ctx, pc.GetPgxConfig())
	if err != nil {
		cancel()
		return nil, err
	}
	err = pgxPool.Ping(ctx)
	if err != nil {
		cancel()
		return nil, err
	}

	return &pgxClientImpl{
		ctx:     ctx,
		cancel:  cancel,
		pgxPool: pgxPool,
	}, nil

}

type pgxClientImpl struct {
	ctx             context.Context
	cancel          context.CancelFunc
	pgxPool         *pgxpool.Pool
}

func (p *pgxClientImpl) GetPgxPool() *pgxpool.Pool {
	return p.pgxPool
}

func (p *pgxClientImpl) AcquireConnection(ctx context.Context) *pgxpool.Conn{
	connection ,err := p.pgxPool.Acquire(ctx)
	if err != nil {
		panic(err)
	}
	return connection
}

func (p *pgxClientImpl) Ping() error {
	return p.pgxPool.Ping(p.ctx)
}

func (p *pgxClientImpl) Close() {
	p.pgxPool.Close()
}

type PgxPoolClient interface {
	Ping() error
	Close()
	GetPgxPool() *pgxpool.Pool
	AcquireConnection(ctx context.Context) *pgxpool.Conn
}