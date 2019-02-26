package redis

import (
	"github.com/garyburd/redigo/redis"
	"tollsys/tollmon/g"
	"time"
	"os"
	"fmt"
)

var (
	pool *redis.Pool
)
//InitRedis 初始化redis模块
// 通过g.config中redis配置创建redis连接池
func InitRedis() {
	g.LogInfo("Init Redis:", g.Config().Redis.Host)
	pool = &redis.Pool{}
	pool.Dial = func() (redis.Conn, error) {
		c, err := redis.Dial(g.Config().Redis.ConnectType, g.Config().Redis.Host)
		if err != nil {
			panic(err.Error())
		}
		return c, nil
	}
	pool.IdleTimeout = time.Duration(5 * time.Second)
	pool.MaxActive = 0
	pool.MaxConnLifetime = 0
	pool.MaxIdle = g.Config().Redis.MaxPoolSize
	pool.TestOnBorrow = func(c redis.Conn, t time.Time) error {
		if time.Since(t) < time.Minute {
			return nil
		}
		_, err := c.Do("ping")
		return err
	}

	_, err := pool.Get().Do("ping")
	if err != nil {
		fmt.Println(err.Error())
		g.LogDebug(err)
		os.Exit(1)
	}
}

func Set(key string, value interface{}) {
	_, err := pool.Get().Do("SET", key, value)
	if err != nil {
		g.LogError("redis Set err:", err.Error())
		return
	}
}
func Get(key string) []byte {
	s, err := redis.Bytes(pool.Get().Do("GET", key))
	if err != nil {
		g.LogError("redis Get Key", key, " err:", err.Error())
		panic(err.Error())
	}
	return s
}
func Del(key string) {
	_, err := pool.Get().Do("DEL", key)
	if err != nil {
		g.LogError("redis Get Key", key, " err:", err.Error())
		panic(err.Error())
	}
}
func IsExist(key string) bool {
	b, err := redis.Bool(pool.Get().Do("EXISTS", key))
	if err != nil {
		g.LogError("redis check exists err:", err.Error())
		panic(err.Error())
	}
	return b
}
//HSetNx Redis hash操作 --> Create
//要求字段key：string Hash表键值；field：string Hash字段，value：[]byte hash中field字段的值
//HSet Only Not Exists 仅当本次记录不存在是才会写入
func HSetNX(key string, field string, value []byte) {
	n, err := pool.Get().Do("HSETNX", key, field, value)
	if err != nil {
		g.LogError("redis Set Json Key", field, " err:", err.Error())
		panic(err.Error())
	}
	if n == int64(1) {
		g.LogInfo("Create to redis ok", field)
	}
	if n == int64(0) {
		g.LogError("Create to redis fail", field)
	}
}
//HSet Redis hash操作 --> Update
//要求字段key：string Hash表键值；field：string Hash字段，value：[]byte hash中field字段的值
func HSet(key string, field string, value []byte) {
	n, err := pool.Get().Do("HSET", key, field, value)
	if err != nil {
		g.LogError("redis Set Json Key", field, " err:", err.Error())
		panic(err.Error())
	}
	if n.(int64) == 0 {
		g.LogInfo("update to redis ok - ", field)
	}
}
//HGet Redis hash操作 --> Select
//要求字段key：string Hash表键值；field：string Hash字段
//返回[]byte hash中field字段的值
func HGet(key string, field string) []byte {
	b, err := redis.Bytes(pool.Get().Do("HGET", key, field))
	if err != nil {
		g.LogError("redis Hash Get err:", err.Error())
		panic(err.Error())
	}
	return b
}
//HDel Redis hash操作 --> Delete
//要求字段key：string Hash表键值；field：string Hash字段
//删除Hash中field字段
func HDel(key string, field string) {
	_, err := pool.Get().Do("HDEL", key, field)
	if err != nil {
		g.LogError("redis Hash DEL err:", err.Error())
		panic(err.Error())
	}
	g.LogInfo("redis HDel key:", key, " field:", field, " delete ok")
}
//HExists Redis Hash操作 --> isExists
//要求字段key：string Hash表键值；field：string Hash字段
//返回bool 该字段是否存在
func HExists(key string, field string) bool {
	n, err := pool.Get().Do("HEXISTS", key, field)
	if err != nil {
		g.LogError("redis Hash Check Exists err:", err.Error())
		panic(err.Error())
	}
	if n.(int64) == 1 {
		return true
	}
	return false
}
//HGetALL Redis Hash操作 --> select all
//要求字段key：string Hash表键值
//返回Hash中所有值
func HGetALL(key string)[]interface{}{
	b,err := redis.Values(pool.Get().Do("HGETALL",key))
	if err != nil{
		g.LogError("redis Hash GetALL err:", err.Error())
		panic(err.Error())
	}
	return b
}
