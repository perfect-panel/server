package console

import (
	"context"
	"time"

	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/internal/types"
	"github.com/perfect-panel/ppanel-server/pkg/logger"
	"github.com/perfect-panel/ppanel-server/pkg/xerr"
	"github.com/pkg/errors"
)

type QueryRevenueStatisticsLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Query revenue statistics
func NewQueryRevenueStatisticsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryRevenueStatisticsLogic {
	return &QueryRevenueStatisticsLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryRevenueStatisticsLogic) QueryRevenueStatistics() (resp *types.RevenueStatisticsResponse, err error) {

	var today, monthly, all types.OrdersStatistics
	now := time.Now()
	// Get today's revenue statistics
	todayData, err := l.svcCtx.OrderModel.QueryDateOrders(l.ctx, now)
	if err != nil {
		l.Errorw("[QueryRevenueStatisticsLogic] QueryDateOrders error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "QueryDateOrders error: %v", err)
	} else {
		today = types.OrdersStatistics{
			AmountTotal:        todayData.AmountTotal,
			NewOrderAmount:     todayData.NewOrderAmount,
			RenewalOrderAmount: todayData.RenewalOrderAmount,
		}
	}
	// Get monthly's revenue statistics
	monthlyData, err := l.svcCtx.OrderModel.QueryMonthlyOrders(l.ctx, now)
	if err != nil {
		l.Errorw("[QueryRevenueStatisticsLogic] QueryDateOrders error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "QueryDateOrders error: %v", err)
	} else {
		monthly = types.OrdersStatistics{
			AmountTotal:        monthlyData.AmountTotal,
			NewOrderAmount:     monthlyData.NewOrderAmount,
			RenewalOrderAmount: monthlyData.RenewalOrderAmount,
			List:               make([]types.OrdersStatistics, 0),
		}
	}

	// Get all revenue statistics
	allData, err := l.svcCtx.OrderModel.QueryTotalOrders(l.ctx)
	if err != nil {
		l.Errorw("[QueryRevenueStatisticsLogic] QueryTotalOrders error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "QueryTotalOrders error: %v", err)
	} else {
		all = types.OrdersStatistics{
			AmountTotal:        allData.AmountTotal,
			NewOrderAmount:     allData.NewOrderAmount,
			RenewalOrderAmount: allData.RenewalOrderAmount,
			List:               make([]types.OrdersStatistics, 0),
		}
	}
	return &types.RevenueStatisticsResponse{
		Today:   today,
		Monthly: monthly,
		All:     all,
	}, nil
}
