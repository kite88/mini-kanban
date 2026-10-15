package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	rdb     *redis.Client
	enabled bool
)

// Init 连接 Redis，失败时降级为“无缓存”模式，保证服务仍可运行
func Init(addr, password string, db int) {
	rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		enabled = false
		return
	}
	enabled = true
}

func Enabled() bool { return enabled && rdb != nil }

func Client() *redis.Client { return rdb }

func Get(key string) ([]byte, bool) {
	if !Enabled() {
		return nil, false
	}
	b, err := rdb.Get(context.Background(), key).Bytes()
	if err != nil {
		return nil, false
	}
	return b, true
}

func Set(key string, val []byte, ttl time.Duration) {
	if !Enabled() {
		return
	}
	rdb.Set(context.Background(), key, val, ttl)
}

func Del(keys ...string) {
	if !Enabled() || len(keys) == 0 {
		return
	}
	rdb.Del(context.Background(), keys...)
}

// DeletePrefix 按前缀批量清理（写操作后失效缓存）
func DeletePrefix(prefix string) {
	if !Enabled() {
		return
	}
	ctx := context.Background()
	iter := rdb.Scan(ctx, 0, prefix+"*", 200).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
		if len(keys) >= 500 {
			rdb.Del(ctx, keys...)
			keys = keys[:0]
		}
	}
	if len(keys) > 0 {
		rdb.Del(ctx, keys...)
	}
}

// IsNotFound 用于区分“key 不存在”与其他错误
func IsNotFound(err error) bool { return errors.Is(err, redis.Nil) }
