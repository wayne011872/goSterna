package db

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wayne011872/goSterna/util"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	CtxGormKey = util.CtxKey("ctxGormKey")
	HeaderGormKey = "raccGormDB"
)

func GetGormDBClientByGin(c *gin.Context) *gorm.DB {
	dbInter, _ := c.Get(string(CtxGormKey))
	if dbclt, ok := dbInter.(*gorm.DB); ok {
		return dbclt
	}
	return nil
}

func GetGormDBClientByReq(req *http.Request) *gorm.DB {
	return GetGormDBClientByCtx(req.Context())
}

func GetGormDBClientByCtx(ctx context.Context) *gorm.DB {
	dbInter := ctx.Value(CtxGormKey)
	if dbclt, ok := dbInter.(*gorm.DB); ok {
		return dbclt
	}
	return nil
}

type GormDI interface {
	NewGormDBClient(ctx context.Context, userDB string) (*gorm.DB, error)
	SetAuth(user, pwd string)
	GetUri() string
}

type GormConf struct {
	Uri       string `yaml:"uri"`
	User      string `yaml:"user"`
	Pass      string `yaml:"pass"`
	DefaultDB string `yaml:"default"`

	authUri  string
	lock     *sync.Mutex
	connPool map[string]*gorm.DB
}

func (gc *GormConf) SetAuth(user, pwd string) {
	gc.authUri = strings.Replace(gc.Uri, "{User}", user, 1)
	gc.authUri = strings.Replace(gc.authUri, "{Pwd}", pwd, 1)
}

func (gc *GormConf) GetUri() string {
	if gc.authUri != "" {
		return gc.authUri
	}
	return gc.Uri
}

func (gc *GormConf) NewGormDBClient(ctx context.Context, userDB string) (*gorm.DB, error) {
	if gc.Uri == "" {
		panic("gorm uri not set")
	}
	if gc.DefaultDB == "" {
		panic("gorm default db not set")
	}

	db, err := gorm.Open(postgres.Open(gc.GetUri()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetConnMaxIdleTime(time.Minute * 10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(10)

	return db, nil
}

func NewGormSession(db *gorm.DB) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true})
}

func GetDBList(db *gorm.DB) ([]string, error) {
	var dbNames []string
	err := db.Raw("SELECT datname FROM pg_database WHERE datistemplate = false;").Scan(&dbNames).Error
	if err != nil {
		return nil, err
	}
	return dbNames, nil
}

func Close(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		fmt.Println("error getting sqlDB: " + err.Error())
		return
	}
	err = sqlDB.Close()
	if err != nil {
		fmt.Println("disconnect error: " + err.Error())
	}
}

func Ping(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

const (
	GormCoreDB = "core"
	GormUserDB = "user"
)
