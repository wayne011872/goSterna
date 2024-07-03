package db

import(
	"github.com/wayne011872/goSterna/util"
	"context"
	"strings"
	"net/http"
	"github.com/gin-gonic/gin"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq" // Import the PostgreSQL driver
)

const (
	CtxSqlKey = util.CtxKey("ctxSqlKey")
)

func GetSqlDBClientByGin(c *gin.Context) SqlClient {
	cltInter ,_:= c.Get(string(CtxSqlKey))
	if dbclt, ok := cltInter.(SqlClient); ok {
		return dbclt
	}
	return nil
}

func GetSqlDBClientByReq(req *http.Request) SqlClient {
	return GetSqlDBClientByCtx(req.Context())
}

func GetSqlDBClientByCtx(ctx context.Context) SqlClient {
	cltInter := ctx.Value(CtxSqlKey)
	if dbclt, ok := cltInter.(SqlClient); ok {
		return dbclt
	}
	return nil
}

type SqlDI interface {
	NewSqlDB(ctx context.Context) (SqlClient, error)
	SetAuth(user, pwd string)
	GetUri() string
}

type SqlConf struct {
	Uri       string `yaml:"uri"`
	User      string `yaml:"user"`
	Pass      string `yaml:"pass"`
	DefaultDB string `yaml:"defaul"`
	SslMode   string `yaml:"sslmode"`

	authUri  string	
}

func (sc *SqlConf) SetAuth(user, pwd string) {
	sc.authUri = strings.Replace(sc.Uri, "{User}", user, 1)
	sc.authUri = strings.Replace(sc.authUri, "{Pwd}", pwd, 1)
}

func (sc *SqlConf) GetUri() string {
	if sc.authUri != "" {
		return sc.authUri+ "/" + sc.DefaultDB + "?sslmode="+sc.SslMode
	}
	return sc.Uri + "/" + sc.DefaultDB + "?sslmode="+sc.SslMode
}

func (sc *SqlConf) NewSqlDB(ctx context.Context) (SqlClient, error) {
	if sc.Uri == "" {
		panic("sql uri not set")
	}
	if sc.DefaultDB == "" {
		panic("sql default db not set")
	}
	sc.SetAuth(sc.User, sc.Pass)
	fmt.Println(sc.GetUri())
	db, err := sql.Open("postgres", sc.GetUri())
	if err != nil {
		defer db.Close()
		return nil, err
	}

	return &sqlClientImpl{
		db: db,
	}, nil

}

type sqlClientImpl struct {
	db       *sql.DB
}

func (s *sqlClientImpl) Ping() error {
	return s.db.Ping()
}

func (s *sqlClientImpl) Close() {
	err:=s.db.Close()
	if err != nil {
		fmt.Println("disconnect error: " + err.Error())
	}
}

func(s *sqlClientImpl) GetDB() *sql.DB {
	return s.db
}


type SqlClient interface {
	Ping() error
	Close()
	GetDB() *sql.DB
}