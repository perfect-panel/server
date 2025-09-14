package server

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type customServerLogicModel interface {
	FindServerListByFilter(ctx context.Context, filter *ServerFilter) (total int64, list []*Server, err error)
	ClearCache(ctx context.Context, id int64) error
	QueryServerCountByServerGroups(ctx context.Context, groupIds []int64) (int64, error)
	QueryAllGroup(ctx context.Context) ([]*Group, error)
	BatchDeleteNodeGroup(ctx context.Context, ids []int64) error
	InsertGroup(ctx context.Context, data *Group) error
	FindOneGroup(ctx context.Context, id int64) (*Group, error)
	UpdateGroup(ctx context.Context, data *Group) error
	DeleteGroup(ctx context.Context, id int64) error
	FindServerDetailByGroupIdsAndIds(ctx context.Context, groupId, ids []int64) ([]*Server, error)
	FindServerListByGroupIds(ctx context.Context, groupId []int64) ([]*Server, error)
	FindAllServer(ctx context.Context) ([]*Server, error)
	FindNodeByServerAddrAndProtocol(ctx context.Context, serverAddr string, protocol string) ([]*Server, error)
	FindServerMinSortByIds(ctx context.Context, ids []int64) (int64, error)
	FindServerListByIds(ctx context.Context, ids []int64) ([]*Server, error)
	InsertRuleGroup(ctx context.Context, data *RuleGroup) error
	FindOneRuleGroup(ctx context.Context, id int64) (*RuleGroup, error)
	UpdateRuleGroup(ctx context.Context, data *RuleGroup) error
	DeleteRuleGroup(ctx context.Context, id int64) error
	QueryAllRuleGroup(ctx context.Context) ([]*RuleGroup, error)
	FindServersByTag(ctx context.Context, tag string) ([]*Server, error)
	FindServerTags(ctx context.Context) ([]string, error)

	SetDefaultRuleGroup(ctx context.Context, id int64) error
}

var (
	CacheServerDetailPrefix     = "cache:server:detail:"
	cacheServerGroupAllKeys     = "cache:serverGroup:all"
	cacheServerRuleGroupAllKeys = "cache:serverRuleGroup:all"
)

// ClearCache Clear Cache
func (m *customServerModel) ClearCache(ctx context.Context, id int64) error {
	serverIdKey := fmt.Sprintf("%s%v", cacheServerIdPrefix, id)
	//configKey := fmt.Sprintf("%s%d", config.ServerConfigCacheKey, id)
	//userListKey := fmt.Sprintf("%s%v", config.ServerUserListCacheKey, id)

	return m.DelCacheCtx(ctx, serverIdKey)
}

// QueryServerCountByServerGroups Query  Server Count By Server Groups
func (m *customServerModel) QueryServerCountByServerGroups(ctx context.Context, groupIds []int64) (int64, error) {
	var count int64
	err := m.QueryNoCacheCtx(ctx, &count, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Server{}).Where("group_id IN ?", groupIds).Count(&count).Error
	})
	return count, err
}

// QueryAllGroup returns all groups.
func (m *customServerModel) QueryAllGroup(ctx context.Context) ([]*Group, error) {
	var groups []*Group
	err := m.QueryCtx(ctx, &groups, cacheServerGroupAllKeys, func(conn *gorm.DB, v interface{}) error {
		return conn.Find(&groups).Error
	})
	return groups, err
}

// BatchDeleteNodeGroup deletes multiple groups.
func (m *customServerModel) BatchDeleteNodeGroup(ctx context.Context, ids []int64) error {
	return m.Transaction(ctx, func(tx *gorm.DB) error {
		for _, id := range ids {
			if err := m.Delete(ctx, id); err != nil {
				return err
			}
		}
		return nil
	})
}

// InsertGroup inserts a group.
func (m *customServerModel) InsertGroup(ctx context.Context, data *Group) error {
	return m.ExecCtx(ctx, func(conn *gorm.DB) error {
		return conn.Create(data).Error
	}, cacheServerGroupAllKeys)
}

// FindOneGroup finds a group.
func (m *customServerModel) FindOneGroup(ctx context.Context, id int64) (*Group, error) {
	var group Group
	err := m.QueryCtx(ctx, &group, fmt.Sprintf("cache:serverGroup:%v", id), func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Group{}).Where("id = ?", id).First(&group).Error
	})
	return &group, err
}

// UpdateGroup updates a group.
func (m *customServerModel) UpdateGroup(ctx context.Context, data *Group) error {
	return m.ExecCtx(ctx, func(conn *gorm.DB) error {
		return conn.Model(&Group{}).Where("id = ?", data.Id).Updates(data).Error
	}, cacheServerGroupAllKeys, fmt.Sprintf("cache:serverGroup:%v", data.Id))
}

// DeleteGroup deletes a group.
func (m *customServerModel) DeleteGroup(ctx context.Context, id int64) error {
	return m.ExecCtx(ctx, func(conn *gorm.DB) error {
		return conn.Where("id = ?", id).Delete(&Group{}).Error
	}, cacheServerGroupAllKeys, fmt.Sprintf("cache:serverGroup:%v", id))
}

// FindServerDetailByGroupIdsAndIds finds server details by group IDs and IDs.
func (m *customServerModel) FindServerDetailByGroupIdsAndIds(ctx context.Context, groupId, ids []int64) ([]*Server, error) {
	if len(groupId) == 0 && len(ids) == 0 {
		return []*Server{}, nil
	}
	var list []*Server
	err := m.QueryNoCacheCtx(ctx, &list, func(conn *gorm.DB, v interface{}) error {
		conn = conn.
			Model(&Server{}).
			Where("`enable` = ?", true)
		if len(groupId) > 0 && len(ids) > 0 {
			// OR is used to connect group_id and id conditions
			conn = conn.Where("(`group_id` IN ? OR `id` IN ?)", groupId, ids)
		} else if len(groupId) > 0 {
			conn = conn.Where("`group_id` IN ?", groupId)
		} else if len(ids) > 0 {
			conn = conn.Where("`id` IN ?", ids)
		}

		return conn.Order("sort ASC").Find(v).Error
	})
	return list, err
}

func (m *customServerModel) FindServerListByGroupIds(ctx context.Context, groupId []int64) ([]*Server, error) {
	var data []*Server
	err := m.QueryNoCacheCtx(ctx, &data, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Server{}).Where("group_id IN ?", groupId).Find(v).Error
	})
	return data, err
}

func (m *customServerModel) FindAllServer(ctx context.Context) ([]*Server, error) {
	var data []*Server
	err := m.QueryNoCacheCtx(ctx, &data, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Server{}).Order("sort ASC").Find(v).Error
	})
	return data, err
}

func (m *customServerModel) FindNodeByServerAddrAndProtocol(ctx context.Context, serverAddr string, protocol string) ([]*Server, error) {
	var data []*Server
	err := m.QueryNoCacheCtx(ctx, &data, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Server{}).Where("server_addr = ?  and protocol = ?", serverAddr, protocol).Order("sort ASC").Find(v).Error
	})
	return data, err
}

func (m *customServerModel) FindServerMinSortByIds(ctx context.Context, ids []int64) (int64, error) {
	var minSort int64
	err := m.QueryNoCacheCtx(ctx, &minSort, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Server{}).Where("id IN ?", ids).Select("COALESCE(MIN(sort), 0)").Scan(v).Error
	})
	return minSort, err
}

func (m *customServerModel) FindServerListByIds(ctx context.Context, ids []int64) ([]*Server, error) {
	var list []*Server
	err := m.QueryNoCacheCtx(ctx, &list, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Server{}).Where("id IN ?", ids).Find(v).Error
	})
	return list, err
}

// InsertRuleGroup inserts a group.
func (m *customServerModel) InsertRuleGroup(ctx context.Context, data *RuleGroup) error {
	return m.ExecCtx(ctx, func(conn *gorm.DB) error {
		return conn.Where(&RuleGroup{}).Create(data).Error
	}, cacheServerRuleGroupAllKeys, fmt.Sprintf("cache:serverRuleGroup:%v", data.Id))
}

// FindOneRuleGroup finds a group.
func (m *customServerModel) FindOneRuleGroup(ctx context.Context, id int64) (*RuleGroup, error) {
	var group RuleGroup
	err := m.QueryCtx(ctx, &group, fmt.Sprintf("cache:serverRuleGroup:%v", id), func(conn *gorm.DB, v interface{}) error {
		return conn.Where(&RuleGroup{}).Model(&RuleGroup{}).Where("id = ?", id).First(&group).Error
	})
	return &group, err
}

// UpdateRuleGroup updates a group.
func (m *customServerModel) UpdateRuleGroup(ctx context.Context, data *RuleGroup) error {
	return m.ExecCtx(ctx, func(conn *gorm.DB) error {
		return conn.Where(&RuleGroup{}).Model(&RuleGroup{}).Where("id = ?", data.Id).Save(data).Error
	}, cacheServerRuleGroupAllKeys, fmt.Sprintf("cache:serverRuleGroup:%v", data.Id))
}

// DeleteRuleGroup deletes a group.
func (m *customServerModel) DeleteRuleGroup(ctx context.Context, id int64) error {
	return m.ExecCtx(ctx, func(conn *gorm.DB) error {
		return conn.Where(&RuleGroup{}).Where("id = ?", id).Delete(&RuleGroup{}).Error
	}, cacheServerRuleGroupAllKeys, fmt.Sprintf("cache:serverRuleGroup:%v", id))
}

// QueryAllRuleGroup returns all rule groups.
func (m *customServerModel) QueryAllRuleGroup(ctx context.Context) ([]*RuleGroup, error) {
	var groups []*RuleGroup
	err := m.QueryCtx(ctx, &groups, cacheServerRuleGroupAllKeys, func(conn *gorm.DB, v interface{}) error {
		return conn.Where(&RuleGroup{}).Find(&groups).Error
	})
	return groups, err
}

func (m *customServerModel) FindServerListByFilter(ctx context.Context, filter *ServerFilter) (total int64, list []*Server, err error) {
	var data []*Server
	if filter == nil {
		filter = &ServerFilter{
			Page: 1,
			Size: 10,
		}
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Size <= 0 {
		filter.Size = 10
	}

	err = m.QueryNoCacheCtx(ctx, &data, func(conn *gorm.DB, v interface{}) error {
		query := conn.Model(&Server{}).Order("sort ASC")
		if filter.Group > 0 {
			query = conn.Where("group_id = ?", filter.Group)
		}
		if filter.Search != "" {
			query = query.Where("name LIKE ? OR server_addr LIKE ? OR tags LIKE ?", "%"+filter.Search+"%", "%"+filter.Search+"%", "%"+filter.Search+"%")
		}
		if len(filter.Tags) > 0 {
			for i, tag := range filter.Tags {
				if i == 0 {
					query = query.Where("tags LIKE ?", "%"+tag+"%")
				} else {
					query = query.Or("tags LIKE ?", "%"+tag+"%")
				}
			}
		}
		return query.Count(&total).Limit(filter.Size).Offset((filter.Page - 1) * filter.Size).Find(v).Error
	})
	if err != nil {
		return 0, nil, err
	}
	return total, data, nil
}

func (m *customServerModel) FindServerTags(ctx context.Context) ([]string, error) {
	var data []string
	err := m.QueryNoCacheCtx(ctx, &data, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Server{}).Distinct("tags").Pluck("tags", v).Error
	})
	var tags []string
	for _, tag := range data {
		if strings.Contains(tag, ",") {
			tags = append(tags, strings.Split(tag, ",")...)
		} else {
			tags = append(tags, tag)
		}
	}
	return tags, err
}

func (m *customServerModel) FindServersByTag(ctx context.Context, tag string) ([]*Server, error) {
	var data []*Server
	err := m.QueryNoCacheCtx(ctx, &data, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Server{}).Where("FIND_IN_SET(?, tags)", tag).Order("sort ASC").Find(v).Error
	})
	return data, err
}

// SetDefaultRuleGroup sets the default rule group.

func (m *customServerModel) SetDefaultRuleGroup(ctx context.Context, id int64) error {
	return m.ExecCtx(ctx, func(conn *gorm.DB) error {
		// Reset all groups to not default
		if err := conn.Model(&RuleGroup{}).Where("`id` != ?", id).Update("default", false).Error; err != nil {
			return err
		}
		// Set the specified group as default
		return conn.Model(&RuleGroup{}).Where("`id` = ?", id).Update("default", true).Error
	}, cacheServerRuleGroupAllKeys, fmt.Sprintf("cache:serverRuleGroup:%v", id))
}
