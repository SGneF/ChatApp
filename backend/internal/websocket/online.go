package websocket

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const onlineTTL = 120 * time.Second

// Redis 在线状态
type OnlineService struct {
	rdb *redis.Client
}

func NewOnlineService(rdb *redis.Client) *OnlineService {
	return &OnlineService{
		rdb: rdb,
	}
}

func onlineUserKey(userID uint64) string {
	return fmt.Sprintf("im:online:user:%d", userID)
}

func (s *OnlineService) SetOnline(ctx context.Context, userID uint64) error {
	if s == nil || s.rdb == nil {
		return nil
	}

	return s.rdb.Set(ctx, onlineUserKey(userID), "1", onlineTTL).Err() //为什么要设置 TTL，因为：服务器可能突然崩溃，来不及执行：SetOffline()，如果没有 TTL，用户会永远显示在线。如果没有 TTL，用户会永远显示在线。

}

func (s *OnlineService) RefreshOnline(ctx context.Context, userID uint64) error {
	if s == nil || s.rdb == nil {
		return nil
	}

	return s.rdb.Expire(ctx, onlineUserKey(userID), onlineTTL).Err()
}

// 用户最后一个连接断开时，立即删除在线状态
func (s *OnlineService) SetOffline(ctx context.Context, userID uint64) error {
	if s == nil || s.rdb == nil {
		return nil
	}

	return s.rdb.Del(ctx, onlineUserKey(userID)).Err()
}

func (s *OnlineService) IsOnline(ctx context.Context, userID uint64) (bool, error) {
	if s == nil || s.rdb == nil {
		return false, nil
	}

	n, err := s.rdb.Exists(ctx, onlineUserKey(userID)).Result()
	if err != nil {
		return false, err
	}

	return n > 0, nil
}
