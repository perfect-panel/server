package node

import "context"

type customServerLogicModel interface {
	FilterServerList(ctx context.Context, params *FilterParams) (int64, []*Server, error)
	FilterNodeList(ctx context.Context, params *FilterParams) (int64, []*Node, error)
}

// FilterParams Filter Server Params
type FilterParams struct {
	Page   int
	Size   int
	Search string
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
	err := query.Count(&total).Limit(params.Size).Offset((params.Page - 1) * params.Size).Find(&servers).Error
	return total, servers, err
}

// FilterNodeList Filter Node List
func (m *customServerModel) FilterNodeList(ctx context.Context, params *FilterParams) (int64, []*Node, error) {
	var nodes []*Node
	var total int64
	query := m.WithContext(ctx).Model(&Node{})
	if params == nil {
		params = &FilterParams{
			Page: 1,
			Size: 10,
		}
	}
	if params.Search != "" {
		s := "%" + params.Search + "%"
		query = query.Where("`name` LIKE ? OR `address` LIKE ? OR `tags` LIKE ? OR `port` LIKE ? ", s, s, s, s)
	}
	err := query.Count(&total).Limit(params.Size).Offset((params.Page - 1) * params.Size).Find(&nodes).Error
	return total, nodes, err
}
