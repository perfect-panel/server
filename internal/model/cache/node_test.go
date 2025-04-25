package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Create a test Redis client
func newTestRedisClient(t *testing.T) *redis.Client {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	require.NoError(t, client.Ping(context.Background()).Err())
	return client
}

// Clean up test data
func cleanupTestData(t *testing.T, client *redis.Client) {
	ctx := context.Background()
	keys := []string{
		UserTodayUploadTrafficCacheKey,
		UserTodayDownloadTrafficCacheKey,
		UserTodayTotalTrafficCacheKey,
		NodeTodayUploadTrafficCacheKey,
		NodeTodayDownloadTrafficCacheKey,
		NodeTodayTotalTrafficCacheKey,
		UserTodayUploadTrafficRankKey,
		UserTodayDownloadTrafficRankKey,
		UserTodayTotalTrafficRankKey,
		NodeTodayUploadTrafficRankKey,
		NodeTodayDownloadTrafficRankKey,
		NodeTodayTotalTrafficRankKey,
	}

	// Clean up all cache keys
	for _, key := range keys {
		require.NoError(t, client.Del(ctx, key).Err())
	}

	// Clean up user online IP cache
	for uid := int64(1); uid <= 3; uid++ {
		require.NoError(t, client.Del(ctx, fmt.Sprintf(UserOnlineIpCacheKey, uid)).Err())
	}

	// Clean up node online user cache
	for nodeId := int64(1); nodeId <= 3; nodeId++ {
		require.NoError(t, client.Del(ctx, fmt.Sprintf(NodeOnlineUserCacheKey, nodeId)).Err())
	}
}

func TestNodeCacheClient_AddUserTodayTraffic(t *testing.T) {
	client := newTestRedisClient(t)
	defer cleanupTestData(t, client)

	cache := NewNodeCacheClient(client)
	ctx := context.Background()

	tests := []struct {
		name     string
		uid      int64
		upload   int64
		download int64
		wantErr  bool
	}{
		{
			name:     "Add traffic normally",
			uid:      1,
			upload:   100,
			download: 200,
			wantErr:  false,
		},
		{
			name:     "Invalid SID",
			uid:      0,
			upload:   100,
			download: 200,
			wantErr:  true,
		},
		{
			name:     "Invalid upload traffic",
			uid:      1,
			upload:   0,
			download: 200,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cache.AddUserTodayTraffic(ctx, tt.uid, tt.upload, tt.download)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Verify data is added correctly
			upload, err := client.HGet(ctx, UserTodayUploadTrafficCacheKey, "1").Int64()
			assert.NoError(t, err)
			assert.Equal(t, tt.upload, upload)

			download, err := client.HGet(ctx, UserTodayDownloadTrafficCacheKey, "1").Int64()
			assert.NoError(t, err)
			assert.Equal(t, tt.download, download)
		})
	}
}

func TestNodeCacheClient_AddNodeTodayTraffic(t *testing.T) {
	client := newTestRedisClient(t)
	defer cleanupTestData(t, client)

	cache := NewNodeCacheClient(client)
	ctx := context.Background()

	tests := []struct {
		name        string
		nodeId      int64
		userTraffic []UserTraffic
		wantErr     bool
	}{
		{
			name:   "Add node traffic normally",
			nodeId: 1,
			userTraffic: []UserTraffic{
				{UID: 1, Upload: 100, Download: 200},
				{UID: 2, Upload: 300, Download: 400},
			},
			wantErr: false,
		},
		{
			name:   "Invalid node ID",
			nodeId: 0,
			userTraffic: []UserTraffic{
				{UID: 1, Upload: 100, Download: 200},
			},
			wantErr: true,
		},
		{
			name:        "Empty user traffic data",
			nodeId:      1,
			userTraffic: []UserTraffic{},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cache.AddNodeTodayTraffic(ctx, tt.nodeId, tt.userTraffic)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Verify data is added correctly
			upload, err := client.HGet(ctx, NodeTodayUploadTrafficCacheKey, "1").Int64()
			assert.NoError(t, err)
			assert.Equal(t, int64(400), upload) // 100 + 300

			download, err := client.HGet(ctx, NodeTodayDownloadTrafficCacheKey, "1").Int64()
			assert.NoError(t, err)
			assert.Equal(t, int64(600), download) // 200 + 400
		})
	}
}

func TestNodeCacheClient_GetUserTodayTrafficRank(t *testing.T) {
	client := newTestRedisClient(t)
	defer cleanupTestData(t, client)

	cache := NewNodeCacheClient(client)
	ctx := context.Background()

	// Prepare test data
	testData := []struct {
		uid      int64
		upload   int64
		download int64
	}{
		{1, 100, 200},
		{2, 300, 400},
		{3, 500, 600},
	}

	for _, data := range testData {
		err := cache.AddUserTodayTraffic(ctx, data.uid, data.upload, data.download)
		require.NoError(t, err)
	}

	tests := []struct {
		name    string
		n       int64
		wantErr bool
	}{
		{
			name:    "Get top 2 ranks",
			n:       2,
			wantErr: false,
		},
		{
			name:    "Get all ranks",
			n:       3,
			wantErr: false,
		},
		{
			name:    "Invalid N value",
			n:       0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ranks, err := cache.GetUserTodayTotalTrafficRank(ctx, tt.n)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, ranks, int(tt.n))

			// Verify sorting is correct
			for i := 1; i < len(ranks); i++ {
				assert.GreaterOrEqual(t, ranks[i-1].Total, ranks[i].Total)
			}
		})
	}
}

func TestNodeCacheClient_ResetTodayTrafficData(t *testing.T) {
	client := newTestRedisClient(t)
	defer cleanupTestData(t, client)

	cache := NewNodeCacheClient(client)
	ctx := context.Background()

	// Prepare test data
	err := cache.AddUserTodayTraffic(ctx, 1, 100, 200)
	require.NoError(t, err)
	err = cache.AddNodeTodayTraffic(ctx, 1, []UserTraffic{{UID: 1, Upload: 100, Download: 200}})
	require.NoError(t, err)

	// Test reset functionality
	err = cache.ResetTodayTrafficData(ctx)
	assert.NoError(t, err)

	// Verify data is cleared
	keys := []string{
		UserTodayUploadTrafficCacheKey,
		UserTodayDownloadTrafficCacheKey,
		UserTodayTotalTrafficCacheKey,
		NodeTodayUploadTrafficCacheKey,
		NodeTodayDownloadTrafficCacheKey,
		NodeTodayTotalTrafficCacheKey,
	}

	for _, key := range keys {
		exists, err := client.Exists(ctx, key).Result()
		assert.NoError(t, err)
		assert.Equal(t, int64(0), exists)
	}
}

func TestNodeCacheClient_GetNodeTodayTrafficRank(t *testing.T) {
	client := newTestRedisClient(t)
	defer cleanupTestData(t, client)

	cache := NewNodeCacheClient(client)
	ctx := context.Background()

	// Prepare test data
	testData := []struct {
		nodeId  int64
		traffic []UserTraffic
	}{
		{1, []UserTraffic{{UID: 1, Upload: 100, Download: 200}}},
		{2, []UserTraffic{{UID: 2, Upload: 300, Download: 400}}},
		{3, []UserTraffic{{UID: 3, Upload: 500, Download: 600}}},
	}

	for _, data := range testData {
		err := cache.AddNodeTodayTraffic(ctx, data.nodeId, data.traffic)
		require.NoError(t, err)
	}

	tests := []struct {
		name    string
		n       int64
		wantErr bool
	}{
		{
			name:    "Get top 2 ranks",
			n:       2,
			wantErr: false,
		},
		{
			name:    "Get all ranks",
			n:       3,
			wantErr: false,
		},
		{
			name:    "Invalid N value",
			n:       0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ranks, err := cache.GetNodeTodayTotalTrafficRank(ctx, tt.n)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, ranks, int(tt.n))

			// Verify sorting is correct
			for i := 1; i < len(ranks); i++ {
				assert.GreaterOrEqual(t, ranks[i-1].Total, ranks[i].Total)
			}
		})
	}
}

func TestNodeCacheClient_AddNodeOnlineUser(t *testing.T) {
	client := newTestRedisClient(t)
	defer cleanupTestData(t, client)

	cache := NewNodeCacheClient(client)
	ctx := context.Background()

	tests := []struct {
		name    string
		nodeId  int64
		users   []NodeOnlineUser
		wantErr bool
	}{
		{
			name:   "Add online users normally",
			nodeId: 1,
			users: []NodeOnlineUser{
				{SID: 1, IP: "192.168.1.1"},
				{SID: 2, IP: "192.168.1.2"},
			},
			wantErr: false,
		},
		{
			name:   "Invalid node ID",
			nodeId: 0,
			users: []NodeOnlineUser{
				{SID: 1, IP: "192.168.1.1"},
			},
			wantErr: false,
		},
		{
			name:    "Empty user list",
			nodeId:  1,
			users:   []NodeOnlineUser{},
			wantErr: false,
		},
		{
			name:   "Add duplicate user IP",
			nodeId: 1,
			users: []NodeOnlineUser{
				{SID: 1, IP: "192.168.1.1"},
				{SID: 1, IP: "192.168.1.1"},
			},
			wantErr: false,
		},
		{
			name:   "Multiple IPs for same user",
			nodeId: 1,
			users: []NodeOnlineUser{
				{SID: 1, IP: "192.168.1.1"},
				{SID: 1, IP: "192.168.1.2"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cache.AddOnlineUserIP(ctx, tt.users)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Verify data is added correctly
			for _, user := range tt.users {
				// Get user online IPs
				ips, err := cache.GetUserOnlineIp(ctx, user.SID)
				assert.NoError(t, err)
				assert.Contains(t, ips, user.IP)

				// Verify score is within valid range (current time to 5 minutes later)
				score, err := client.ZScore(ctx, fmt.Sprintf(UserOnlineIpCacheKey, user.SID), user.IP).Result()
				assert.NoError(t, err)
				now := time.Now().Unix()
				assert.GreaterOrEqual(t, score, float64(now))
				assert.LessOrEqual(t, score, float64(now+300)) // 5 minutes = 300 seconds

				// Verify key exists
				exists, err := client.Exists(ctx, fmt.Sprintf(UserOnlineIpCacheKey, user.SID)).Result()
				assert.NoError(t, err)
				assert.Equal(t, int64(1), exists)
			}
		})
	}
}

func TestNodeCacheClient_GetUserOnlineIp(t *testing.T) {
	client := newTestRedisClient(t)
	defer cleanupTestData(t, client)

	cache := NewNodeCacheClient(client)
	ctx := context.Background()

	// Prepare test data
	testData := []struct {
		nodeId int64
		users  []NodeOnlineUser
	}{
		{
			nodeId: 1,
			users: []NodeOnlineUser{
				{SID: 1, IP: "192.168.1.1"},
				{SID: 1, IP: "192.168.1.2"},
				{SID: 2, IP: "192.168.1.3"},
			},
		},
	}

	// Add test data
	for _, data := range testData {
		err := cache.AddOnlineUserIP(ctx, data.users)
		require.NoError(t, err)
	}

	tests := []struct {
		name    string
		uid     int64
		wantErr bool
		wantIPs []string
	}{
		{
			name:    "Get existing user IPs",
			uid:     1,
			wantErr: false,
			wantIPs: []string{"192.168.1.1", "192.168.1.2"},
		},
		{
			name:    "Get another user's IPs",
			uid:     2,
			wantErr: false,
			wantIPs: []string{"192.168.1.3"},
		},
		{
			name:    "Get non-existent user IPs",
			uid:     3,
			wantErr: false,
			wantIPs: []string{},
		},
		{
			name:    "Invalid user ID",
			uid:     0,
			wantErr: true,
		},
		{
			name:    "Expired IPs should not be returned",
			uid:     1,
			wantErr: false,
			wantIPs: []string{"192.168.1.1", "192.168.1.2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ips, err := cache.GetUserOnlineIp(ctx, tt.uid)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.wantIPs, ips)

			// Verify all returned IPs are valid
			for _, ip := range ips {
				score, err := client.ZScore(ctx, fmt.Sprintf(UserOnlineIpCacheKey, tt.uid), ip).Result()
				assert.NoError(t, err)
				now := time.Now().Unix()
				assert.GreaterOrEqual(t, score, float64(now))
			}
		})
	}
}

func TestNodeCacheClient_UpdateNodeOnlineUser(t *testing.T) {
	client := newTestRedisClient(t)
	defer cleanupTestData(t, client)

	cache := NewNodeCacheClient(client)
	ctx := context.Background()

	tests := []struct {
		name    string
		nodeId  int64
		users   []NodeOnlineUser
		wantErr bool
	}{
		{
			name:   "Update online users normally",
			nodeId: 1,
			users: []NodeOnlineUser{
				{SID: 1, IP: "192.168.1.1"},
				{SID: 2, IP: "192.168.1.2"},
			},
			wantErr: false,
		},
		{
			name:   "Invalid node ID",
			nodeId: 0,
			users: []NodeOnlineUser{
				{SID: 1, IP: "192.168.1.1"},
			},
			wantErr: true,
		},
		{
			name:    "Empty user list",
			nodeId:  1,
			users:   []NodeOnlineUser{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cache.UpdateNodeOnlineUser(ctx, tt.nodeId, tt.users)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Verify data is updated correctly
			data, err := client.Get(ctx, fmt.Sprintf(NodeOnlineUserCacheKey, tt.nodeId)).Result()
			assert.NoError(t, err)

			var result map[int64][]string
			err = json.Unmarshal([]byte(data), &result)
			assert.NoError(t, err)

			// Verify data content
			for _, user := range tt.users {
				ips, exists := result[user.SID]
				assert.True(t, exists)
				assert.Contains(t, ips, user.IP)
			}
		})
	}
}
