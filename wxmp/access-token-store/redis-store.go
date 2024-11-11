package accesstokenstore

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/anchel/wechat-official-account-admin/wxmp/common"
	mpoptions "github.com/anchel/wechat-official-account-admin/wxmp/mp-options"
	"github.com/redis/go-redis/v9"
)

type AccessTokenStore interface {
	Store(accessToken string, expire time.Duration) error
	Get() (string, error)
	Refresh(wait bool) (string, error)
}

type RedisStore struct {
	mpoptions *mpoptions.MpOptions
	rdb       *redis.Client
}

func NewRedisStore(mpoptions *mpoptions.MpOptions, rdb *redis.Client) *RedisStore {
	return &RedisStore{
		mpoptions: mpoptions,
		rdb:       rdb,
	}
}

func (r *RedisStore) Store(accessToken string, expire time.Duration) error {
	ret := r.rdb.Set(context.TODO(), r.mpoptions.AppId+"_access_token", accessToken, expire)
	if ret.Err() != nil {
		return ret.Err()
	}
	return nil
}

func (r *RedisStore) Get() (string, error) {
	accessToken := ""
	token, err := r.rdb.Get(context.TODO(), r.mpoptions.AppId+"_access_token").Result()
	if err != nil {
		if err != redis.Nil {
			return "", err
		}
	}
	accessToken = token

	if accessToken == "" { // 不存在
		log.Println("access token not exists")
		accessToken, err = r.Refresh(true)
	} else { // 判断还有多久过期，小于5分钟则刷新
		ex, err := r.rdb.TTL(context.TODO(), r.mpoptions.AppId+"_access_token").Result()
		if err != nil {
			log.Println("get access token ttl fail", err)
			return "", err
		}
		// log.Println("access token ttl:", ex)
		if ex < 300*time.Second {
			r.Refresh(false)
		}
	}

	return accessToken, err
}

func (r *RedisStore) Refresh(wait bool) (string, error) {
	ok, err := r.rdb.SetNX(context.TODO(), r.mpoptions.AppId+"_access_token_lock", "1", 10*time.Second).Result()
	if err != nil {
		return "", err
	}
	if ok { // 获取到锁
		log.Println("get access token lock ok")
		defer r.rdb.Del(context.TODO(), r.mpoptions.AppId+"_access_token_lock")
		newToken, expire, err := common.GetAccessToken(r.mpoptions.AppId, r.mpoptions.AppSecret)
		if err != nil {
			return "", err
		}
		err = r.Store(newToken, time.Duration(expire)*time.Second)
		if err != nil {
			return "", err
		}
		return newToken, nil
	}
	log.Println("get access token lock not ok")
	if wait {
		// 等待其他协程刷新token
		count := 0
		for {
			count++
			if count > 10 {
				return "", errors.New("wait access token timeout")
			}
			time.Sleep(500 * time.Millisecond)

			token, err := r.rdb.Get(context.TODO(), r.mpoptions.AppId+"_access_token").Result()
			if err != nil {
				if err != redis.Nil {
					return "", err
				}
			}

			if err == nil { // 这里通常来说，token不会是空字符串
				return token, nil
			}
		}
	}
	return "", nil
}
