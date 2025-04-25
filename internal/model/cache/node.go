package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/perfect-panel/ppanel-server/pkg/logger"
	"github.com/redis/go-redis/v9"
)

type NodeCacheClient struct {
	*redis.Client
	resetMutex sync.Mutex
}

func NewNodeCacheClient(rds *redis.Client) *NodeCacheClient {
	return &NodeCacheClient{
		Client: rds,
	}
}

// AddOnlineUserIP  adds user's online IP
func (c *NodeCacheClient) AddOnlineUserIP(ctx context.Context, users []NodeOnlineUser) error {
	if len(users) == 0 {
		// No users to add
		return nil
	}

	// Use Pipeline to optimize Redis operations
	pipe := c.Pipeline()

	// Add user online IPs and clean up expired IPs for each user
	for _, user := range users {
		if user.SID <= 0 || user.IP == "" {
			logger.Errorf("invalid user data: uid=%d, ip=%s", user.SID, user.IP)
			continue
		}

		key := fmt.Sprintf(UserOnlineIpCacheKey, user.SID)
		now := time.Now()
		expireTime := now.Add(5 * time.Minute)

		// Clean up expired user online IPs
		pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", now.Unix()))
		pipe.ZRemRangeByScore(ctx, AllNodeOnlineUserCacheKey, "0", fmt.Sprintf("%d", now.Unix()))

		// Add or update user online IP
		// XX: Only update elements that already exist
		// NX: Only add new elements
		_ = pipe.ZAdd(ctx, key, redis.Z{
			Score:  float64(expireTime.Unix()),
			Member: user.IP,
		}).Err()
		_ = pipe.ZAdd(ctx, AllNodeOnlineUserCacheKey, redis.Z{
			Score:  float64(expireTime.Unix()),
			Member: user.IP,
		}).Err()

		// Set key expiration to 5 minutes (slightly longer than IP expiration)
		pipe.Expire(ctx, key, 5*time.Minute)
		pipe.Expire(ctx, AllNodeOnlineUserCacheKey, 5*time.Minute)
	}

	// Execute all commands
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to add node user online ip: %w", err)
	}
	return nil
}

// GetUserOnlineIp gets user's online IPs
func (c *NodeCacheClient) GetUserOnlineIp(ctx context.Context, uid int64) ([]string, error) {
	if uid <= 0 {
		return nil, fmt.Errorf("invalid parameters: uid=%d", uid)
	}

	// Get user's online IPs
	ips, err := c.ZRevRangeByScore(ctx, fmt.Sprintf(UserOnlineIpCacheKey, uid), &redis.ZRangeBy{
		Min:    "0",
		Max:    fmt.Sprintf("%d", time.Now().Add(5*time.Minute).Unix()),
		Offset: 0,
		Count:  100,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user online ip: %w", err)
	}
	return ips, nil
}

// UpdateNodeOnlineUser updates node's online users and IPs
func (c *NodeCacheClient) UpdateNodeOnlineUser(ctx context.Context, nodeId int64, users []NodeOnlineUser) error {
	if nodeId <= 0 || len(users) == 0 {
		return fmt.Errorf("invalid parameters: nodeId=%d, users=%v", nodeId, users)
	}
	// Organize data
	data := make(map[int64][]string)
	for _, user := range users {
		data[user.SID] = append(data[user.SID], user.IP)
	}

	value, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	c.Set(ctx, fmt.Sprintf(NodeOnlineUserCacheKey, nodeId), value, time.Minute*5)
	return nil
}

// GetNodeOnlineUser gets node's online users and IPs
func (c *NodeCacheClient) GetNodeOnlineUser(ctx context.Context, nodeId int64) (map[int64][]string, error) {
	if nodeId <= 0 {
		return nil, fmt.Errorf("invalid parameters: nodeId=%d", nodeId)
	}
	value, err := c.Get(ctx, fmt.Sprintf(NodeOnlineUserCacheKey, nodeId)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get node online user: %w", err)
	}
	var data map[int64][]string
	if err := json.Unmarshal([]byte(value), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}
	return data, nil
}

// AddUserTodayTraffic Add user's today traffic
func (c *NodeCacheClient) AddUserTodayTraffic(ctx context.Context, uid int64, upload, download int64) error {
	if uid <= 0 || upload <= 0 {
		return fmt.Errorf("invalid parameters: uid=%d, upload=%d", uid, upload)
	}
	pipe := c.Pipeline()
	// User's today upload traffic
	pipe.HIncrBy(ctx, UserTodayUploadTrafficCacheKey, fmt.Sprintf("%d", uid), upload)
	// User's today download traffic
	pipe.HIncrBy(ctx, UserTodayDownloadTrafficCacheKey, fmt.Sprintf("%d", uid), download)
	// User's today total traffic
	pipe.HIncrBy(ctx, UserTodayTotalTrafficCacheKey, fmt.Sprintf("%d", uid), upload+download)
	// User's today traffic ranking
	pipe.ZIncrBy(ctx, UserTodayUploadTrafficRankKey, float64(upload), fmt.Sprintf("%d", uid))
	pipe.ZIncrBy(ctx, UserTodayDownloadTrafficRankKey, float64(download), fmt.Sprintf("%d", uid))
	pipe.ZIncrBy(ctx, UserTodayTotalTrafficRankKey, float64(upload+download), fmt.Sprintf("%d", uid))

	// All node upload traffic
	pipe.IncrBy(ctx, AllNodeUploadTrafficCacheKey, upload)
	// All node download traffic
	pipe.IncrBy(ctx, AllNodeDownloadTrafficCacheKey, download)
	// Execute commands
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to add user today upload traffic: %w", err)
	}
	return nil
}

// AddNodeTodayTraffic Add node's today traffic
func (c *NodeCacheClient) AddNodeTodayTraffic(ctx context.Context, nodeId int64, userTraffic []UserTraffic) error {
	if nodeId <= 0 || len(userTraffic) == 0 {
		return fmt.Errorf("invalid parameters: nodeId=%d, userTraffic=%v", nodeId, userTraffic)
	}
	pipe := c.Pipeline()
	upload, download, total := c.calculateTraffic(userTraffic)
	pipe.HIncrBy(ctx, NodeTodayUploadTrafficCacheKey, fmt.Sprintf("%d", nodeId), upload)
	pipe.HIncrBy(ctx, NodeTodayDownloadTrafficCacheKey, fmt.Sprintf("%d", nodeId), download)
	pipe.HIncrBy(ctx, NodeTodayTotalTrafficCacheKey, fmt.Sprintf("%d", nodeId), total)
	pipe.ZIncrBy(ctx, NodeTodayUploadTrafficRankKey, float64(upload), fmt.Sprintf("%d", nodeId))
	pipe.ZIncrBy(ctx, NodeTodayDownloadTrafficRankKey, float64(download), fmt.Sprintf("%d", nodeId))
	pipe.ZIncrBy(ctx, NodeTodayTotalTrafficRankKey, float64(total), fmt.Sprintf("%d", nodeId))
	// Execute commands
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to add node today upload traffic: %w", err)
	}
	return nil
}

// Get user's traffic data
func (c *NodeCacheClient) getUserTrafficData(ctx context.Context, uid int64) (upload, download int64, err error) {
	upload, err = c.HGet(ctx, UserTodayUploadTrafficCacheKey, fmt.Sprintf("%d", uid)).Int64()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get user today upload traffic: %w", err)
	}
	download, err = c.HGet(ctx, UserTodayDownloadTrafficCacheKey, fmt.Sprintf("%d", uid)).Int64()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get user today download traffic: %w", err)
	}
	return upload, download, nil
}

// Get node's traffic data
func (c *NodeCacheClient) getNodeTrafficData(ctx context.Context, nodeId int64) (upload, download int64, err error) {
	upload, err = c.HGet(ctx, NodeTodayUploadTrafficCacheKey, fmt.Sprintf("%d", nodeId)).Int64()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get node today upload traffic: %w", err)
	}
	download, err = c.HGet(ctx, NodeTodayDownloadTrafficCacheKey, fmt.Sprintf("%d", nodeId)).Int64()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get node today download traffic: %w", err)
	}
	return upload, download, nil
}

// Parse ID
func (c *NodeCacheClient) parseID(member interface{}, idType string) (int64, error) {
	id, err := strconv.ParseInt(member.(string), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s id %v: %w", idType, member, err)
	}
	return id, nil
}

// GetUserTodayTotalTrafficRank Get user's today total traffic ranking top N
func (c *NodeCacheClient) GetUserTodayTotalTrafficRank(ctx context.Context, n int64) ([]UserTodayTrafficRank, error) {
	if n <= 0 {
		return nil, fmt.Errorf("invalid parameters: n=%d", n)
	}
	data, err := c.ZRevRangeWithScores(ctx, UserTodayTotalTrafficRankKey, 0, n-1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user today total traffic rank: %w", err)
	}
	users := make([]UserTodayTrafficRank, 0, len(data))
	for _, user := range data {
		uid, err := c.parseID(user.Member, "user")
		if err != nil {
			logger.Errorf("%v", err)
			continue
		}
		upload, download, err := c.getUserTrafficData(ctx, uid)
		if err != nil {
			logger.Errorf("%v", err)
			continue
		}
		users = append(users, UserTodayTrafficRank{
			SID:      uid,
			Upload:   upload,
			Download: download,
			Total:    int64(user.Score),
		})
	}
	return users, nil
}

// GetNodeTodayTotalTrafficRank Get node's today total traffic ranking top N
func (c *NodeCacheClient) GetNodeTodayTotalTrafficRank(ctx context.Context, n int64) ([]NodeTodayTrafficRank, error) {
	if n <= 0 {
		return nil, fmt.Errorf("invalid parameters: n=%d", n)
	}
	data, err := c.ZRevRangeWithScores(ctx, NodeTodayTotalTrafficRankKey, 0, n-1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get node today total traffic rank: %w", err)
	}
	nodes := make([]NodeTodayTrafficRank, 0, len(data))
	for _, node := range data {
		nodeId, err := c.parseID(node.Member, "node")
		if err != nil {
			logger.Errorf("%v", err)
			continue
		}
		upload, download, err := c.getNodeTrafficData(ctx, nodeId)
		if err != nil {
			logger.Errorf("%v", err)
			continue
		}
		nodes = append(nodes, NodeTodayTrafficRank{
			ID:       nodeId,
			Upload:   upload,
			Download: download,
			Total:    int64(node.Score),
		})
	}
	return nodes, nil
}

// GetUserTodayUploadTrafficRank Get user's today upload traffic ranking top N
func (c *NodeCacheClient) GetUserTodayUploadTrafficRank(ctx context.Context, n int64) ([]UserTodayTrafficRank, error) {
	if n <= 0 {
		return nil, fmt.Errorf("invalid parameters: n=%d", n)
	}
	data, err := c.ZRevRangeWithScores(ctx, UserTodayUploadTrafficRankKey, 0, n-1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user today upload traffic rank: %w", err)
	}
	users := make([]UserTodayTrafficRank, 0, len(data))
	for _, user := range data {
		uid, err := c.parseID(user.Member, "user")
		if err != nil {
			logger.Errorf("%v", err)
			continue
		}
		upload, download, err := c.getUserTrafficData(ctx, uid)
		if err != nil {
			logger.Errorf("%v", err)
			continue
		}
		users = append(users, UserTodayTrafficRank{
			SID:      uid,
			Upload:   upload,
			Download: download,
			Total:    int64(user.Score),
		})
	}
	return users, nil
}

// GetUserTodayDownloadTrafficRank Get user's today download traffic ranking top N
func (c *NodeCacheClient) GetUserTodayDownloadTrafficRank(ctx context.Context, n int64) ([]UserTodayTrafficRank, error) {
	if n <= 0 {
		return nil, fmt.Errorf("invalid parameters: n=%d", n)
	}
	data, err := c.ZRevRangeWithScores(ctx, UserTodayDownloadTrafficRankKey, 0, n-1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user today download traffic rank: %w", err)
	}
	users := make([]UserTodayTrafficRank, 0, len(data))
	for _, user := range data {
		uid, err := c.parseID(user.Member, "user")
		if err != nil {
			logger.Errorf("%v", err)
			continue
		}
		upload, download, err := c.getUserTrafficData(ctx, uid)
		if err != nil {
			logger.Errorf("%v", err)
			continue
		}
		users = append(users, UserTodayTrafficRank{
			SID:      uid,
			Upload:   upload,
			Download: download,
			Total:    int64(user.Score),
		})
	}
	return users, nil
}

// GetNodeTodayUploadTrafficRank Get node's today upload traffic ranking top N
func (c *NodeCacheClient) GetNodeTodayUploadTrafficRank(ctx context.Context, n int64) ([]NodeTodayTrafficRank, error) {
	if n <= 0 {
		return nil, fmt.Errorf("invalid parameters: n=%d", n)
	}
	data, err := c.ZRevRangeWithScores(ctx, NodeTodayUploadTrafficRankKey, 0, n-1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get node today upload traffic rank: %w", err)
	}
	nodes := make([]NodeTodayTrafficRank, 0, len(data))
	for _, node := range data {
		nodeId, err := c.parseID(node.Member, "node")
		if err != nil {
			logger.Errorf("%v", err)
			continue
		}
		upload, download, err := c.getNodeTrafficData(ctx, nodeId)
		if err != nil {
			logger.Errorf("%v", err)
			continue
		}
		nodes = append(nodes, NodeTodayTrafficRank{
			ID:       nodeId,
			Upload:   upload,
			Download: download,
			Total:    int64(node.Score),
		})
	}
	return nodes, nil
}

// GetNodeTodayDownloadTrafficRank Get node's today download traffic ranking top N
func (c *NodeCacheClient) GetNodeTodayDownloadTrafficRank(ctx context.Context, n int64) ([]NodeTodayTrafficRank, error) {
	if n <= 0 {
		return nil, fmt.Errorf("invalid parameters: n=%d", n)
	}
	data, err := c.ZRevRangeWithScores(ctx, NodeTodayDownloadTrafficRankKey, 0, n-1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get node today download traffic rank: %w", err)
	}
	nodes := make([]NodeTodayTrafficRank, 0, len(data))
	for _, node := range data {
		nodeId, err := c.parseID(node.Member, "node")
		if err != nil {
			logger.Errorf("%v", err)
			continue
		}
		upload, download, err := c.getNodeTrafficData(ctx, nodeId)
		if err != nil {
			logger.Errorf("%v", err)
			continue
		}
		nodes = append(nodes, NodeTodayTrafficRank{
			ID:       nodeId,
			Upload:   upload,
			Download: download,
			Total:    int64(node.Score),
		})
	}
	return nodes, nil
}

// ResetTodayTrafficData Reset today's traffic data
func (c *NodeCacheClient) ResetTodayTrafficData(ctx context.Context) error {
	c.resetMutex.Lock()
	defer c.resetMutex.Unlock()
	pipe := c.Pipeline()
	pipe.Del(ctx, UserTodayUploadTrafficCacheKey)
	pipe.Del(ctx, UserTodayDownloadTrafficCacheKey)
	pipe.Del(ctx, UserTodayTotalTrafficCacheKey)
	pipe.Del(ctx, NodeTodayUploadTrafficCacheKey)
	pipe.Del(ctx, NodeTodayDownloadTrafficCacheKey)
	pipe.Del(ctx, NodeTodayTotalTrafficCacheKey)
	pipe.Del(ctx, UserTodayUploadTrafficRankKey)
	pipe.Del(ctx, UserTodayDownloadTrafficRankKey)
	pipe.Del(ctx, UserTodayTotalTrafficRankKey)
	pipe.Del(ctx, NodeTodayUploadTrafficRankKey)
	pipe.Del(ctx, NodeTodayDownloadTrafficRankKey)
	pipe.Del(ctx, NodeTodayTotalTrafficRankKey)
	pipe.Del(ctx, AllNodeDownloadTrafficCacheKey)
	pipe.Del(ctx, AllNodeUploadTrafficCacheKey)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to reset today traffic data: %w", err)
	}
	return nil
}

// Calculate traffic
func (c *NodeCacheClient) calculateTraffic(data []UserTraffic) (upload, download, total int64) {
	for _, userTraffic := range data {
		upload += userTraffic.Upload
		download += userTraffic.Download
		total += userTraffic.Upload + userTraffic.Download
	}
	return upload, download, total
}

// GetAllNodeOnlineUser Get all node online user
func (c *NodeCacheClient) GetAllNodeOnlineUser(ctx context.Context) ([]string, error) {
	users, err := c.ZRevRange(ctx, AllNodeOnlineUserCacheKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get all node online user: %w", err)
	}
	return users, nil
}

// UpdateNodeStatus Update node status
func (c *NodeCacheClient) UpdateNodeStatus(ctx context.Context, nodeId int64, status NodeStatus) error {
	// 参数验证
	if nodeId <= 0 {
		return fmt.Errorf("invalid node id: %d", nodeId)
	}

	// 验证状态数据
	if status.UpdatedAt <= 0 {
		return fmt.Errorf("invalid status data: updated_at=%d", status.UpdatedAt)
	}

	// 序列化状态数据
	value, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to marshal node status: %w", err)
	}

	// 使用 Pipeline 优化性能
	pipe := c.Pipeline()

	// 设置状态数据
	pipe.Set(ctx, fmt.Sprintf(NodeStatusCacheKey, nodeId), value, time.Minute*5)

	// 执行命令
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update node status: %w", err)
	}

	return nil
}

// GetNodeStatus Get node status
func (c *NodeCacheClient) GetNodeStatus(ctx context.Context, nodeId int64) (NodeStatus, error) {
	status, err := c.Get(ctx, fmt.Sprintf(NodeStatusCacheKey, nodeId)).Result()
	if err != nil {
		return NodeStatus{}, fmt.Errorf("failed to get node status: %w", err)
	}
	var nodeStatus NodeStatus
	if err := json.Unmarshal([]byte(status), &nodeStatus); err != nil {
		return NodeStatus{}, fmt.Errorf("failed to unmarshal node status: %w", err)
	}
	return nodeStatus, nil
}

// GetOnlineNodeStatusCount Get Online Node Status Count
func (c *NodeCacheClient) GetOnlineNodeStatusCount(ctx context.Context) (int64, error) {
	// 获取所有节点Key
	keys, err := c.Keys(ctx, "node:status:*").Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get all node status keys: %w", err)
	}
	var count int64
	for _, key := range keys {
		status, err := c.Get(ctx, key).Result()
		if err != nil {
			logger.Errorf("failed to get node status: %v", err.Error())
			continue
		}
		if status != "" {
			count++
		}
	}
	return count, nil
}

// GetAllNodeUploadTraffic Get all node upload traffic
func (c *NodeCacheClient) GetAllNodeUploadTraffic(ctx context.Context) (int64, error) {
	upload, err := c.Get(ctx, AllNodeUploadTrafficCacheKey).Int64()
	if err != nil {
		return 0, fmt.Errorf("failed to get all node upload traffic: %w", err)
	}
	return upload, nil
}

// GetAllNodeDownloadTraffic Get all node download traffic
func (c *NodeCacheClient) GetAllNodeDownloadTraffic(ctx context.Context) (int64, error) {
	download, err := c.Get(ctx, AllNodeDownloadTrafficCacheKey).Int64()
	if err != nil {
		return 0, fmt.Errorf("failed to get all node download traffic: %w", err)
	}
	return download, nil
}

// UpdateYesterdayNodeTotalTrafficRank Update yesterday node total traffic rank
func (c *NodeCacheClient) UpdateYesterdayNodeTotalTrafficRank(ctx context.Context, nodes []NodeTodayTrafficRank) error {
	expireAt := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local).Add(time.Hour * 24)
	t := time.Until(expireAt)
	pipe := c.Pipeline()
	value, _ := json.Marshal(nodes)
	pipe.Set(ctx, YesterdayNodeTotalTrafficRank, value, t)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update yesterday node total traffic rank: %w", err)
	}
	return nil
}

// UpdateYesterdayUserTotalTrafficRank Update yesterday user total traffic rank
func (c *NodeCacheClient) UpdateYesterdayUserTotalTrafficRank(ctx context.Context, users []UserTodayTrafficRank) error {
	expireAt := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local).Add(time.Hour * 24)
	t := time.Until(expireAt)
	pipe := c.Pipeline()
	value, _ := json.Marshal(users)
	pipe.Set(ctx, YesterdayUserTotalTrafficRank, value, t)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update yesterday user total traffic rank: %w", err)
	}
	return nil
}

// GetYesterdayNodeTotalTrafficRank Get yesterday node total traffic rank
func (c *NodeCacheClient) GetYesterdayNodeTotalTrafficRank(ctx context.Context) ([]NodeTodayTrafficRank, error) {
	value, err := c.Get(ctx, YesterdayNodeTotalTrafficRank).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get yesterday node total traffic rank: %w", err)
	}
	var nodes []NodeTodayTrafficRank
	if err := json.Unmarshal([]byte(value), &nodes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yesterday node total traffic rank: %w", err)
	}
	return nodes, nil
}

// GetYesterdayUserTotalTrafficRank Get yesterday user total traffic rank
func (c *NodeCacheClient) GetYesterdayUserTotalTrafficRank(ctx context.Context) ([]UserTodayTrafficRank, error) {
	value, err := c.Get(ctx, YesterdayUserTotalTrafficRank).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get yesterday user total traffic rank: %w", err)
	}
	var users []UserTodayTrafficRank
	if err := json.Unmarshal([]byte(value), &users); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yesterday user total traffic rank: %w", err)
	}
	return users, nil
}
