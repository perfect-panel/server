package node

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type (
	customCacheLogicModel interface {
		StatusCache(ctx context.Context, serverId int64) (Status, error)
		UpdateStatusCache(ctx context.Context, serverId int64, status *Status) error
		OnlineUserSubscribe(ctx context.Context, serverId int64, protocol string) (OnlineUserSubscribe, error)
		UpdateOnlineUserSubscribe(ctx context.Context, serverId int64, protocol string, subscribe OnlineUserSubscribe) error
		OnlineUserSubscribeGlobal(ctx context.Context) (int64, error)
		UpdateOnlineUserSubscribeGlobal(ctx context.Context, subscribe OnlineUserSubscribe) error
		AliveListByUID(ctx context.Context) (map[int64]int64, error)
		AliveIPsByUID(ctx context.Context) (map[int64][]string, error)
		ListOnlineIPsByUID(ctx context.Context, uid int64) ([]string, error)
		CleanupOnlineUserUIDIndex(ctx context.Context) error
		RemoveAliveIP(ctx context.Context, uid int64, ip string) error
		IncrRejectCount(ctx context.Context, uid int64, serverId int64, delta int64, reason string) error
		RejectCount24hByUID(ctx context.Context, uid int64) (int64, error)
		RejectCount24hBatch(ctx context.Context, uids []int64) (map[int64]int64, error)
	}

	Status struct {
		Cpu       float64 `json:"cpu"`
		Mem       float64 `json:"mem"`
		Disk      float64 `json:"disk"`
		UpdatedAt int64   `json:"updated_at"`
	}

	OnlineUserSubscribe map[int64][]string
)

// Marshal  to json string
func (s *Status) Marshal() string {
	type Alias Status
	data, _ := json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	})
	return string(data)
}

// Unmarshal from json string
func (s *Status) Unmarshal(data string) error {
	type Alias Status
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	return json.Unmarshal([]byte(data), &aux)
}

const (
	Expiry                                = 300 * time.Second              // Cache expiry time in seconds
	StatusCacheKey                        = "node:status:%d"               // Node status cache key format (Server ID and protocol) Example: node:status:1:shadowsocks
	OnlineUserCacheKeyWithSubscribe       = "node:online:subscribe:%d:%s"  // Online user subscribe cache key format (Server ID and protocol) Example: node:online:subscribe:1:shadowsocks
	OnlineUserSubscribeCacheKeyWithGlobal = "node:online:subscribe:global" // Online user global subscribe cache key
	OnlineUserUIDCacheKey                 = "node:online:uid:%d"           // Per-uid online IP ZSet (member=IP, score=expire ts)
	OnlineUserUIDIndexKey                 = "node:online:uid:index"        // Active uid index Set for AliveList aggregation
	AliveListCacheKey                     = "node:alivelist:cache"         // 2s local cache for aggregated alivelist
	AliveListCacheTTL                     = 2 * time.Second
	// OnlineScoreWindow MUST be > PushInterval (default 60s) or the ZSet drops
	// every IP between two pushes and alivelist returns empty. Real-time eviction
	// is delegated to the node-side LRU; the server window only needs to be long
	// enough to survive one missed push cycle.
	OnlineScoreWindow   = 3 * time.Minute
	OnlineUserUIDKeyTTL = 2 * OnlineScoreWindow // safety TTL (2x score window)
	// RejectCounterKey is a single Hash. Field = "<uid>:<server_id>" to prevent double-counting
	// across multiple node instances on the same host. Whole-key TTL rolls forward on every write.
	RejectCounterKey    = "user:reject:counter"
	RejectCounterKeyTTL = 24 * time.Hour
)

// UpdateStatusCache Update server status to cache
func (m *customServerModel) UpdateStatusCache(ctx context.Context, serverId int64, status *Status) error {
	key := fmt.Sprintf(StatusCacheKey, serverId)
	return m.Cache.Set(ctx, key, status.Marshal(), Expiry).Err()

}

// DeleteStatusCache Delete server status from cache
func (m *customServerModel) DeleteStatusCache(ctx context.Context, serverId int64) error {
	key := fmt.Sprintf(StatusCacheKey, serverId)
	return m.Cache.Del(ctx, key).Err()
}

// StatusCache Get server status from cache
func (m *customServerModel) StatusCache(ctx context.Context, serverId int64) (Status, error) {
	var status Status
	key := fmt.Sprintf(StatusCacheKey, serverId)

	result, err := m.Cache.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return status, nil
		}
		return status, err
	}
	if result == "" {
		return status, nil
	}
	err = status.Unmarshal(result)
	return status, err
}

// OnlineUserSubscribe Get online user subscribe
func (m *customServerModel) OnlineUserSubscribe(ctx context.Context, serverId int64, protocol string) (OnlineUserSubscribe, error) {
	key := fmt.Sprintf(OnlineUserCacheKeyWithSubscribe, serverId, protocol)
	result, err := m.Cache.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return OnlineUserSubscribe{}, nil
		}
		return nil, err
	}
	if result == "" {
		return OnlineUserSubscribe{}, nil
	}
	var subscribe OnlineUserSubscribe
	err = json.Unmarshal([]byte(result), &subscribe)
	return subscribe, err
}

// UpdateOnlineUserSubscribe Update online user subscribe
func (m *customServerModel) UpdateOnlineUserSubscribe(ctx context.Context, serverId int64, protocol string, subscribe OnlineUserSubscribe) error {
	key := fmt.Sprintf(OnlineUserCacheKeyWithSubscribe, serverId, protocol)
	data, err := json.Marshal(subscribe)
	if err != nil {
		return err
	}
	return m.Cache.Set(ctx, key, data, Expiry).Err()
}

// DeleteOnlineUserSubscribe Delete online user subscribe
func (m *customServerModel) DeleteOnlineUserSubscribe(ctx context.Context, serverId int64, protocol string) error {
	key := fmt.Sprintf(OnlineUserCacheKeyWithSubscribe, serverId, protocol)
	return m.Cache.Del(ctx, key).Err()
}

// OnlineUserSubscribeGlobal Get global online user subscribe count
func (m *customServerModel) OnlineUserSubscribeGlobal(ctx context.Context) (int64, error) {
	now := time.Now().Unix()
	// Clear expired data
	if err := m.Cache.ZRemRangeByScore(ctx, OnlineUserSubscribeCacheKeyWithGlobal, "-inf", fmt.Sprintf("%d", now)).Err(); err != nil {
		return 0, err
	}
	return m.Cache.ZCard(ctx, OnlineUserSubscribeCacheKeyWithGlobal).Result()
}

// UpdateOnlineUserSubscribeGlobal Update global online user subscribe count and per-uid IP ZSet.
// Writes three structures atomically via pipeline:
//  1. node:online:subscribe:global  -> ZSet of unique uids (console total count)
//  2. node:online:uid:%d            -> ZSet of IPs per uid (feeds alivelist)
//  3. node:online:uid:index         -> Set of active uids (enables cheap aggregation)
func (m *customServerModel) UpdateOnlineUserSubscribeGlobal(ctx context.Context, subscribe OnlineUserSubscribe) error {
	now := time.Now()
	nowUnix := now.Unix()
	expireTime := now.Add(OnlineScoreWindow).Unix()
	expireStr := fmt.Sprintf("%d", nowUnix)

	pipe := m.Cache.Pipeline()

	// (1) Global uid ZSet - expire old entries, upsert current
	pipe.ZRemRangeByScore(ctx, OnlineUserSubscribeCacheKeyWithGlobal, "-inf", expireStr)

	for uid, ips := range subscribe {
		// (1) Global: count unique uids
		pipe.ZAdd(ctx, OnlineUserSubscribeCacheKeyWithGlobal, redis.Z{
			Score:  float64(expireTime),
			Member: uid,
		})

		// (2) Per-uid ZSet: IPs with expire score
		uidKey := fmt.Sprintf(OnlineUserUIDCacheKey, uid)
		pipe.ZRemRangeByScore(ctx, uidKey, "-inf", expireStr)
		if len(ips) > 0 {
			members := make([]redis.Z, 0, len(ips))
			for _, ip := range ips {
				if ip == "" {
					continue
				}
				members = append(members, redis.Z{
					Score:  float64(expireTime),
					Member: ip,
				})
			}
			if len(members) > 0 {
				pipe.ZAdd(ctx, uidKey, members...)
				pipe.Expire(ctx, uidKey, OnlineUserUIDKeyTTL)
			}
		}

		// (3) Active uid index
		pipe.SAdd(ctx, OnlineUserUIDIndexKey, uid)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// AliveListByUID aggregates distinct online IP count per uid across all nodes.
// Uses a 2s local cache (stored in Redis) to absorb bursty /alivelist calls from many nodes.
// Returns map[uid]ip_count; uids with zero IPs are omitted.
func (m *customServerModel) AliveListByUID(ctx context.Context) (map[int64]int64, error) {
	// Try short cache first
	if cached, err := m.Cache.Get(ctx, AliveListCacheKey).Result(); err == nil && cached != "" {
		var out map[int64]int64
		if err := json.Unmarshal([]byte(cached), &out); err == nil {
			return out, nil
		}
	}

	uidStrs, err := m.Cache.SMembers(ctx, OnlineUserUIDIndexKey).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}
	if len(uidStrs) == 0 {
		return map[int64]int64{}, nil
	}

	nowStr := fmt.Sprintf("%d", time.Now().Unix())

	// Pipeline expire-trim + ZCARD for each uid
	pipe := m.Cache.Pipeline()
	cardCmds := make(map[int64]*redis.IntCmd, len(uidStrs))
	uids := make([]int64, 0, len(uidStrs))
	for _, s := range uidStrs {
		var uid int64
		if _, err := fmt.Sscan(s, &uid); err != nil {
			continue
		}
		uids = append(uids, uid)
		uidKey := fmt.Sprintf(OnlineUserUIDCacheKey, uid)
		pipe.ZRemRangeByScore(ctx, uidKey, "-inf", nowStr)
		cardCmds[uid] = pipe.ZCard(ctx, uidKey)
	}
	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	result := make(map[int64]int64, len(uids))
	for uid, cmd := range cardCmds {
		if cnt, err := cmd.Result(); err == nil && cnt > 0 {
			result[uid] = cnt
		}
	}

	if data, err := json.Marshal(result); err == nil {
		_ = m.Cache.Set(ctx, AliveListCacheKey, data, AliveListCacheTTL).Err()
	}
	return result, nil
}

// AliveIPsByUID returns live IPs per uid, ordered oldest → newest. The node uses
// the ordering for LRU replacement: when a new IP pushes the count above
// device_limit, the oldest IP is the one to drop locally.
//
// Compared to AliveListByUID (which only returns counts), this sends the full
// IP set — payload is O(users × device_limit) which stays small.
func (m *customServerModel) AliveIPsByUID(ctx context.Context) (map[int64][]string, error) {
	uidStrs, err := m.Cache.SMembers(ctx, OnlineUserUIDIndexKey).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}
	if len(uidStrs) == 0 {
		return map[int64][]string{}, nil
	}

	nowStr := fmt.Sprintf("%d", time.Now().Unix())

	pipe := m.Cache.Pipeline()
	rangeCmds := make(map[int64]*redis.StringSliceCmd, len(uidStrs))
	for _, s := range uidStrs {
		var uid int64
		if _, err := fmt.Sscan(s, &uid); err != nil {
			continue
		}
		uidKey := fmt.Sprintf(OnlineUserUIDCacheKey, uid)
		pipe.ZRemRangeByScore(ctx, uidKey, "-inf", nowStr)
		// ZRange 0..-1 returns members by score ascending (oldest expire first).
		rangeCmds[uid] = pipe.ZRange(ctx, uidKey, 0, -1)
	}
	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	result := make(map[int64][]string, len(rangeCmds))
	for uid, cmd := range rangeCmds {
		ips, err := cmd.Result()
		if err != nil || len(ips) == 0 {
			continue
		}
		result[uid] = ips
	}
	return result, nil
}

// ListOnlineIPsByUID returns the distinct currently-online public IPs for a uid
// across all nodes. Expired members are trimmed first. Empty slice is a valid
// answer (user fully offline).
func (m *customServerModel) ListOnlineIPsByUID(ctx context.Context, uid int64) ([]string, error) {
	uidKey := fmt.Sprintf(OnlineUserUIDCacheKey, uid)
	nowStr := fmt.Sprintf("%d", time.Now().Unix())

	pipe := m.Cache.Pipeline()
	pipe.ZRemRangeByScore(ctx, uidKey, "-inf", nowStr)
	rangeCmd := pipe.ZRange(ctx, uidKey, 0, -1)
	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}
	ips, err := rangeCmd.Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}
	if ips == nil {
		return []string{}, nil
	}
	return ips, nil
}

// CleanupOnlineUserUIDIndex removes uids from the index whose per-uid ZSet is empty
// (either drained by ZRemRangeByScore or never repopulated). Intended for a 5-minute cron.
func (m *customServerModel) CleanupOnlineUserUIDIndex(ctx context.Context) error {
	uidStrs, err := m.Cache.SMembers(ctx, OnlineUserUIDIndexKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}
		return err
	}
	nowStr := fmt.Sprintf("%d", time.Now().Unix())

	pipe := m.Cache.Pipeline()
	cardCmds := make(map[string]*redis.IntCmd, len(uidStrs))
	for _, s := range uidStrs {
		uidKey := fmt.Sprintf(OnlineUserUIDCacheKey, s)
		pipe.ZRemRangeByScore(ctx, uidKey, "-inf", nowStr)
		cardCmds[s] = pipe.ZCard(ctx, uidKey)
	}
	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		return err
	}

	stale := make([]interface{}, 0)
	for s, cmd := range cardCmds {
		if cnt, err := cmd.Result(); err == nil && cnt == 0 {
			stale = append(stale, s)
		}
	}
	if len(stale) == 0 {
		return nil
	}
	return m.Cache.SRem(ctx, OnlineUserUIDIndexKey, stale...).Err()
}

// DeleteOnlineUserSubscribeGlobal Delete global online user subscribe count
func (m *customServerModel) DeleteOnlineUserSubscribeGlobal(ctx context.Context) error {
	return m.Cache.Del(ctx, OnlineUserSubscribeCacheKeyWithGlobal).Err()
}

// RemoveAliveIP forcibly drops a single IP from a uid's alive ZSet. Called when
// a node LRU-evicts an IP locally so the server view catches up immediately
// instead of waiting for the IP's natural score expiry. Idempotent.
//
// Note: if another node is still pushing this IP, its next push restores it
// with a fresh score — that is desired (the IP is genuinely active elsewhere).
func (m *customServerModel) RemoveAliveIP(ctx context.Context, uid int64, ip string) error {
	if uid <= 0 || ip == "" {
		return nil
	}
	uidKey := fmt.Sprintf(OnlineUserUIDCacheKey, uid)
	if err := m.Cache.ZRem(ctx, uidKey, ip).Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}
		return err
	}
	// Bust the 2s alivelist cache so subsequent reads see the fresh state.
	_ = m.Cache.Del(ctx, AliveListCacheKey).Err()
	return nil
}

// rejectField returns the Hash field key for a (uid, server_id) pair.
// Storing per-(uid, server_id) prevents duplicate counting when multiple node
// instances on the same host both observe the same user hitting device_limit.
func rejectField(uid, serverId int64) string {
	return fmt.Sprintf("%d:%d", uid, serverId)
}

// IncrRejectCount atomically adds delta to the reject counter for (uid, server_id)
// and refreshes the 24h rolling TTL. reason is accepted for forward compatibility
// but not yet stored (only aggregate count is tracked per V3.1 scope).
func (m *customServerModel) IncrRejectCount(ctx context.Context, uid int64, serverId int64, delta int64, _ string) error {
	if delta <= 0 {
		return nil
	}
	pipe := m.Cache.Pipeline()
	pipe.HIncrBy(ctx, RejectCounterKey, rejectField(uid, serverId), delta)
	pipe.Expire(ctx, RejectCounterKey, RejectCounterKeyTTL)
	_, err := pipe.Exec(ctx)
	return err
}

// RejectCount24hByUID sums counts across all server_id slots for a single uid.
// Uses HSCAN with MATCH to avoid loading the entire Hash.
func (m *customServerModel) RejectCount24hByUID(ctx context.Context, uid int64) (int64, error) {
	pattern := fmt.Sprintf("%d:*", uid)
	var cursor uint64
	var total int64
	for {
		kvs, next, err := m.Cache.HScan(ctx, RejectCounterKey, cursor, pattern, 200).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				return 0, nil
			}
			return 0, err
		}
		// HScan returns interleaved [field, value, field, value, ...]
		for i := 1; i < len(kvs); i += 2 {
			var n int64
			if _, err := fmt.Sscan(kvs[i], &n); err == nil {
				total += n
			}
		}
		if next == 0 {
			return total, nil
		}
		cursor = next
	}
}

// RejectCount24hBatch aggregates counts for many uids in a single HGETALL pass.
// Preferred over N x RejectCount24hByUID when the admin list needs a whole page.
func (m *customServerModel) RejectCount24hBatch(ctx context.Context, uids []int64) (map[int64]int64, error) {
	out := make(map[int64]int64, len(uids))
	if len(uids) == 0 {
		return out, nil
	}
	wanted := make(map[int64]struct{}, len(uids))
	for _, u := range uids {
		wanted[u] = struct{}{}
	}
	all, err := m.Cache.HGetAll(ctx, RejectCounterKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return out, nil
		}
		return nil, err
	}
	for field, val := range all {
		// field format: "<uid>:<server_id>"
		var uid int64
		if _, err := fmt.Sscanf(field, "%d:", &uid); err != nil {
			continue
		}
		if _, ok := wanted[uid]; !ok {
			continue
		}
		var n int64
		if _, err := fmt.Sscan(val, &n); err == nil {
			out[uid] += n
		}
	}
	return out, nil
}
