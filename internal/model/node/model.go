package node

import (
	"context"
	"fmt"

	"github.com/perfect-panel/server/pkg/tool"
)

type customServerLogicModel interface {
	FilterServerList(ctx context.Context, params *FilterParams) (int64, []*Server, error)
	FilterNodeList(ctx context.Context, params *FilterNodeParams) (int64, []*Node, error)
	ClearNodeCache(ctx context.Context, params *FilterNodeParams) error
}

const (
	// ServerUserListCacheKey Server User List Cache Key
	ServerUserListCacheKey = "server:user:"

	// ServerConfigCacheKey Server Config Cache Key
	ServerConfigCacheKey = "server:config:"
)

// FilterParams Filter Server Params
type FilterParams struct {
	Page   int
	Size   int
	Ids    []int64 // Server IDs
	Search string
}

type FilterNodeParams struct {
	Page     int      // Page Number
	Size     int      // Page Size
	NodeId   []int64  // Node IDs
	ServerId []int64  // Server IDs
	Tag      []string // Tags
	Search   string   // Search Address or Name
	Protocol string   // Protocol
	Preload  bool     // Preload Server
	Enabled  *bool    // Enabled
}

// FilterServerList Filter Server List
func (m *customServerModel) FilterServerList(ctx context.Context, params *FilterParams) (int64, []*Server, error) {
	var servers []*Server
	var total int64
	query := m.WithContext(ctx).Model(&Server{})
	if params == nil {
		params = &FilterParams{
			Page: 1,
			Size: 10,
		}
	}
	if params.Search != "" {
		s := "%" + params.Search + "%"
		query = query.Where("`name` LIKE ? OR `address` LIKE ?", s, s)
	}
	if len(params.Ids) > 0 {
		query = query.Where("id IN ?", params.Ids)
	}
	err := query.Count(&total).Order("sort ASC").Limit(params.Size).Offset((params.Page - 1) * params.Size).Find(&servers).Error
	return total, servers, err
}

// FilterNodeList Filter Node List
func (m *customServerModel) FilterNodeList(ctx context.Context, params *FilterNodeParams) (int64, []*Node, error) {
	var nodes []*Node
	var total int64
	query := m.WithContext(ctx).Model(&Node{})
	if params == nil {
		params = &FilterNodeParams{
			Page: 1,
			Size: 10,
		}
	}
	if params.Search != "" {
		s := "%" + params.Search + "%"
		query = query.Where("`name` LIKE ? OR `address` LIKE ? OR `tags` LIKE ? OR `port` LIKE ? ", s, s, s, s)
	}
	if len(params.NodeId) > 0 {
		query = query.Where("id IN ?", params.NodeId)
	}
	if len(params.ServerId) > 0 {
		query = query.Where("server_id IN ?", params.ServerId)
	}
	if len(params.Tag) > 0 {
		for _, tag := range params.Tag {
			query = query.Or("FIND_IN_SET(?,tags)", tag)
		}
	}
	if params.Protocol != "" {
		query = query.Where("protocol = ?", params.Protocol)
	}

	if params.Enabled != nil {
		query = query.Where("enabled = ?", *params.Enabled)
	}

	if params.Preload {
		query = query.Preload("Server")
	}

	err := query.Count(&total).Order("sort ASC").Limit(params.Size).Offset((params.Page - 1) * params.Size).Find(&nodes).Error
	return total, nodes, err
}

// ClearNodeCache Clear Node Cache
func (m *customServerModel) ClearNodeCache(ctx context.Context, params *FilterNodeParams) error {
	_, nodes, err := m.FilterNodeList(ctx, params)
	if err != nil {
		return err
	}
	var cacheKeys []string
	for _, node := range nodes {
		cacheKeys = append(cacheKeys, fmt.Sprintf("%s%d", ServerUserListCacheKey, node.ServerId))
		if node.Protocol != "" {
			var cursor uint64
			for {
				keys, newCursor, err := m.Cache.Scan(ctx, cursor, fmt.Sprintf("%s%d*", ServerConfigCacheKey, node.ServerId), 100).Result()
				if err != nil {
					return err
				}
				if len(keys) > 0 {
					cacheKeys = append(keys, keys...)
				}
				cursor = newCursor
				if cursor == 0 {
					break
				}
			}
		}
	}

	if len(cacheKeys) > 0 {
		cacheKeys = tool.RemoveDuplicateElements(cacheKeys...)
		return m.Cache.Del(ctx, cacheKeys...).Err()
	}
	return nil
}

// ClearServerCache Clear Server Cache
func (m *customServerModel) ClearServerCache(ctx context.Context, serverId int64) error {
	var cacheKeys []string
	cacheKeys = append(cacheKeys, fmt.Sprintf("%s%d", ServerUserListCacheKey, serverId))
	var cursor uint64
	for {
		keys, newCursor, err := m.Cache.Scan(ctx, 0, fmt.Sprintf("%s%d*", ServerConfigCacheKey, serverId), 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			cacheKeys = append(cacheKeys, keys...)
		}
		cursor = newCursor
		if cursor == 0 {
			break
		}
	}

	if len(cacheKeys) > 0 {
		cacheKeys = tool.RemoveDuplicateElements(cacheKeys...)
		return m.Cache.Del(ctx, cacheKeys...).Err()
	}
	return nil
}
