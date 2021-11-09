package app

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"gitee.com/kirile/kapi/app/toml"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"strings"
	"time"
	"xorm.io/core"
	"xorm.io/xorm"
)

var (
	DB          *xorm.Engine
	Redis       *redis.Client
	TablePrefix = "api_base_"

	Config toml.Document //以toml作为配置文件, 放在和可执行文件相同目录下 config.toml
)

const PasswordSalt = "qijin@gyuasda"

func Sha256(string2 string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(string2+PasswordSalt)))
}

// InitDB 以mysql连接串初始化数据库连接
func InitDB(dsn string) (err error) {
	DB, err = xorm.NewEngine("mysql", dsn)
	if err != nil {
		return err
	}
	DB.SetTableMapper(core.NewPrefixMapper(core.SameMapper{}, TablePrefix))
	DB.SetColumnMapper(core.SameMapper{})
	return nil
}

// InitDB2 以配置文件节(Section)来初始化数据库连接 默认为 [db]
func InitDB2(sec ...string) error {
	pathEnable := "db.enable"
	if !Config.Bool(pathEnable) {
		//LOG
		return nil
	}
	path := "db.mysql"
	if len(sec) > 0 {
		path = sec[0] + ".mysql"
	}

	dsn := Config.String(path, "")
	if strings.Trim(dsn, " ") == "" {
		return errors.New("未配置数据库连接")
	}
	return InitDB(dsn)
}

func InitRedis(address string, password string, db ...int) (err error) {
	mydb := 1
	if len(db) > 0 {
		mydb = db[0]
	}
	Redis = redis.NewClient(&redis.Options{
		DB:          mydb,
		Addr:        address,
		Password:    password,
		DialTimeout: time.Second * 2,
	})
	//if err := Redis.Ping().Err(); err != nil {
	//	panic(err)
	//}
	return nil
}
func InitRedis2(sec ...string) error {
	pathEnable := "redis.enable"
	if !Config.Bool(pathEnable) {
		//LOG
		return nil
	}
	path := "redis"
	if len(sec) > 0 {
		path = sec[0]
	}
	redisAddress := Config.String(path+".address", "")
	redisPassword := Config.String(path+".password", "")
	redisDB := Config.Int(path+".db", 1)
	return InitRedis(redisAddress, redisPassword, redisDB)
}

func Lock(appId string) bool {
	if Redis.Exists(appId).Val() == 1 {
		return false
	}
	return Redis.SetNX(appId, 1, time.Second*0).Val()
}

func UnLock(appId string) int64 {
	return Redis.Del(appId).Val()
}

type DoHandler func(s *xorm.Session) error

// Trans 开始一个事务
func Trans(d *xorm.Engine, f DoHandler) error {
	session := d.NewSession()
	err := session.Begin()
	if err != nil {
		return err
	}
	defer session.Close()
	err = f(session)
	if err != nil {
		err1 := session.Rollback()
		if err1 != nil {
			return err1
		}
		return err
	}
	err = session.Commit()
	if err != nil {
		err2 := session.Rollback()
		if err2 != nil {
			return err2
		}
		return err
	}
	return nil
}

func init() {
	Config = toml.ParseFile("config.toml")
}
