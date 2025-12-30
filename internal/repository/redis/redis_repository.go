package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/game-apps/pkg/cache"
)

// Repository Redis 数据访问层
type Repository struct {
	cache *cache.Client
}

// NewRepository 创建 Redis 仓库
func NewRepository(cache *cache.Client) *Repository {
	return &Repository{cache: cache}
}

// SessionRepository 会话存储
type SessionRepository struct {
	*Repository
}

// NewSessionRepository 创建会话仓库
func NewSessionRepository(repo *Repository) *SessionRepository {
	return &SessionRepository{Repository: repo}
}

// SetSession 设置会话
func (r *SessionRepository) SetSession(ctx context.Context, userID uint, data map[string]interface{}, expiration time.Duration) error {
	key := fmt.Sprintf("session:%d", userID)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return r.cache.Set(ctx, key, jsonData, expiration)
}

// GetSession 获取会话
func (r *SessionRepository) GetSession(ctx context.Context, userID uint) (map[string]interface{}, error) {
	key := fmt.Sprintf("session:%d", userID)
	data, err := r.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil, err
	}
	return result, nil
}

// DeleteSession 删除会话
func (r *SessionRepository) DeleteSession(ctx context.Context, userID uint) error {
	key := fmt.Sprintf("session:%d", userID)
	return r.cache.Del(ctx, key)
}

// RoomRepository 房间缓存
type RoomRepository struct {
	*Repository
}

// NewRoomRepository 创建房间仓库
func NewRoomRepository(repo *Repository) *RoomRepository {
	return &RoomRepository{Repository: repo}
}

// SetRoomState 设置房间状态
func (r *RoomRepository) SetRoomState(ctx context.Context, roomID uint, data map[string]interface{}, expiration time.Duration) error {
	key := fmt.Sprintf("room:%d", roomID)
	return r.cache.HSet(ctx, key, data)
}

// GetRoomState 获取房间状态
func (r *RoomRepository) GetRoomState(ctx context.Context, roomID uint) (map[string]string, error) {
	key := fmt.Sprintf("room:%d", roomID)
	return r.cache.HGetAll(ctx, key)
}

// AddRoomPlayer 添加房间玩家
func (r *RoomRepository) AddRoomPlayer(ctx context.Context, roomID uint, userID uint) error {
	key := fmt.Sprintf("room:players:%d", roomID)
	return r.cache.SAdd(ctx, key, userID)
}

// RemoveRoomPlayer 移除房间玩家
func (r *RoomRepository) RemoveRoomPlayer(ctx context.Context, roomID uint, userID uint) error {
	key := fmt.Sprintf("room:players:%d", roomID)
	return r.cache.SRem(ctx, key, userID)
}

// GetRoomPlayers 获取房间玩家列表
func (r *RoomRepository) GetRoomPlayers(ctx context.Context, roomID uint) ([]string, error) {
	key := fmt.Sprintf("room:players:%d", roomID)
	return r.cache.SMembers(ctx, key)
}

// IsRoomPlayer 检查是否是房间玩家
func (r *RoomRepository) IsRoomPlayer(ctx context.Context, roomID uint, userID uint) (bool, error) {
	key := fmt.Sprintf("room:players:%d", roomID)
	return r.cache.SIsMember(ctx, key, fmt.Sprintf("%d", userID))
}

// DeleteRoom 删除房间缓存
func (r *RoomRepository) DeleteRoom(ctx context.Context, roomID uint) error {
	roomKey := fmt.Sprintf("room:%d", roomID)
	playersKey := fmt.Sprintf("room:players:%d", roomID)
	return r.cache.Del(ctx, roomKey, playersKey)
}

// Client 获取 Redis 客户端
func (r *RoomRepository) Client() *cache.Client {
	return r.cache
}

// OnlineUserRepository 在线用户管理
type OnlineUserRepository struct {
	*Repository
}

// NewOnlineUserRepository 创建在线用户仓库
func NewOnlineUserRepository(repo *Repository) *OnlineUserRepository {
	return &OnlineUserRepository{Repository: repo}
}

// AddOnlineUser 添加在线用户
func (r *OnlineUserRepository) AddOnlineUser(ctx context.Context, userID uint) error {
	return r.cache.SAdd(ctx, "user:online", userID)
}

// RemoveOnlineUser 移除在线用户
func (r *OnlineUserRepository) RemoveOnlineUser(ctx context.Context, userID uint) error {
	return r.cache.SRem(ctx, "user:online", userID)
}

// IsOnline 检查用户是否在线
func (r *OnlineUserRepository) IsOnline(ctx context.Context, userID uint) (bool, error) {
	return r.cache.SIsMember(ctx, "user:online", userID)
}

// GetOnlineUsers 获取所有在线用户
func (r *OnlineUserRepository) GetOnlineUsers(ctx context.Context) ([]string, error) {
	return r.cache.SMembers(ctx, "user:online")
}

// LockRepository 分布式锁
type LockRepository struct {
	*Repository
}

// NewLockRepository 创建锁仓库
func NewLockRepository(repo *Repository) *LockRepository {
	return &LockRepository{Repository: repo}
}

// AcquireLock 获取锁
func (r *LockRepository) AcquireLock(ctx context.Context, resource string, expiration time.Duration) (bool, error) {
	key := fmt.Sprintf("lock:%s", resource)
	return r.cache.SetNX(ctx, key, "1", expiration)
}

// ReleaseLock 释放锁
func (r *LockRepository) ReleaseLock(ctx context.Context, resource string) error {
	key := fmt.Sprintf("lock:%s", resource)
	return r.cache.Del(ctx, key)
}

