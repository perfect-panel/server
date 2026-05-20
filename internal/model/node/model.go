package node

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/perfect-panel/server/pkg/orm"
	"github.com/perfect-panel/server/pkg/tool"
	"gorm.io/gorm"
)

type customServerLogicModel interface {
	FilterServerList(ctx context.Context, params *FilterParams) (int64, []*Server, error)
	FilterNodeList(ctx context.Context, params *FilterNodeParams) (int64, []*Node, error)
	QueryNodeSorts(ctx context.Context) ([]SortItem, error)
	QueryServerSorts(ctx context.Context) ([]SortItem, error)
	UpdateNodeSort(ctx context.Context, id int64, sort int64) error
	UpdateServerSort(ctx context.Context, id int64, sort int64) error
	QueryNodeTags(ctx context.Context) ([]string, error)
	CountEnabledNodes(ctx context.Context) (int64, error)
	CountServersByReportStatus(ctx context.Context, cutoff time.Time) (int64, int64, error)
	QueryServerAddresses(ctx context.Context) ([]string, error)
	QueryEnabledNodeProtocols(ctx context.Context) ([]string, error)
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

type SortItem struct {
	Id   int64
	Sort int64
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
		query = query.Scopes(orm.PrefixLike([]string{"name", "address"}, params.Search))
	}
	if len(params.Ids) > 0 {
		query = query.Where("id IN ?", params.Ids)
	}
	err := query.Count(&total).Order("sort ASC").Limit(params.Size).Offset((params.Page - 1) * params.Size).Find(&servers).Error
	return total, servers, err
}

func (m *customServerModel) QueryServerList(ctx context.Context, ids []int64) (servers []*Server, err error) {
	query := m.WithContext(ctx).Model(&Server{})
	err = query.Where("id IN (?)", ids).Find(&servers).Error
	return
}

func (m *customServerModel) QueryServerSorts(ctx context.Context) ([]SortItem, error) {
	var items []SortItem
	err := m.WithContext(ctx).Model(&Server{}).Select("id", "sort").Order("sort ASC").Find(&items).Error
	return items, err
}

func (m *customServerModel) UpdateServerSort(ctx context.Context, id int64, sort int64) error {
	server, err := m.FindOneServer(ctx, id)
	if err != nil {
		return err
	}
	server.Sort = int(sort)
	return m.UpdateServer(ctx, server)
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
		pattern := orm.LikePrefixPattern(params.Search)
		condition := "(name LIKE ? ESCAPE '\\' OR address LIKE ? ESCAPE '\\' OR tags LIKE ? ESCAPE '\\'"
		args := []interface{}{pattern, pattern, pattern}
		if port, err := strconv.ParseUint(params.Search, 10, 16); err == nil {
			condition += " OR port = ?"
			args = append(args, uint16(port))
		}
		condition += ")"
		query = query.Where(condition, args...)
	}
	if len(params.NodeId) > 0 {
		query = query.Where("id IN ?", params.NodeId)
	}
	if len(params.ServerId) > 0 {
		query = query.Where("server_id IN ?", params.ServerId)
	}
	if len(params.Tag) > 0 {
		query = query.Scopes(InSet("tags", params.Tag))
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

func (m *customServerModel) QueryNodeSorts(ctx context.Context) ([]SortItem, error) {
	var items []SortItem
	err := m.WithContext(ctx).Model(&Node{}).Select("id", "sort").Order("sort ASC").Find(&items).Error
	return items, err
}

func (m *customServerModel) UpdateNodeSort(ctx context.Context, id int64, sort int64) error {
	node, err := m.FindOneNode(ctx, id)
	if err != nil {
		return err
	}
	node.Sort = int(sort)
	return m.UpdateNode(ctx, node)
}

func (m *customServerModel) QueryNodeTags(ctx context.Context) ([]string, error) {
	var tags []string
	err := m.WithContext(ctx).Model(&Node{}).Pluck("tags", &tags).Error
	return tags, err
}

func (m *customServerModel) CountEnabledNodes(ctx context.Context) (int64, error) {
	var total int64
	err := m.WithContext(ctx).Model(&Node{}).Where("enabled = ?", true).Count(&total).Error
	return total, err
}

func (m *customServerModel) CountServersByReportStatus(ctx context.Context, cutoff time.Time) (int64, int64, error) {
	var online int64
	if err := m.WithContext(ctx).Model(&Server{}).Where("last_reported_at > ?", cutoff).Count(&online).Error; err != nil {
		return 0, 0, err
	}

	var offline int64
	if err := m.WithContext(ctx).Model(&Server{}).Where("last_reported_at <= ? OR last_reported_at IS NULL", cutoff).Count(&offline).Error; err != nil {
		return 0, 0, err
	}

	return online, offline, nil
}

func (m *customServerModel) QueryServerAddresses(ctx context.Context) ([]string, error) {
	var addresses []string
	err := m.WithContext(ctx).Model(&Server{}).Pluck("address", &addresses).Error
	return addresses, err
}

func (m *customServerModel) QueryEnabledNodeProtocols(ctx context.Context) ([]string, error) {
	var protocols []string
	err := m.WithContext(ctx).Model(&Node{}).Where("enabled = ?", true).Pluck("protocol", &protocols).Error
	return protocols, err
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
					cacheKeys = append(cacheKeys, keys...)
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

// InSet 支持多值 OR 查询
func InSet(field string, values []string) func(db *gorm.DB) *gorm.DB {
	return orm.CommaSeparatedContains(field, values)
}
