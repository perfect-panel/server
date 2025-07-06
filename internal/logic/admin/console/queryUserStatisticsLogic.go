package console

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
)

type QueryUserStatisticsLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Query user statistics
func NewQueryUserStatisticsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryUserStatisticsLogic {
	return &QueryUserStatisticsLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryUserStatisticsLogic) QueryUserStatistics() (resp *types.UserStatisticsResponse, err error) {
	if strings.ToLower(os.Getenv("PPANEL_MODE")) == "demo" {
		return l.mockRevenueStatistics(), nil
	}
	resp = &types.UserStatisticsResponse{}
	now := time.Now()
	// query today user register count
	todayUserResisterCount, err := l.svcCtx.UserModel.QueryResisterUserTotalByDate(l.ctx, now)
	if err != nil {
		l.Errorw("[QueryUserStatisticsLogic] QueryResisterUserTotalByDate error", logger.Field("error", err.Error()))
	} else {
		resp.Today.Register = todayUserResisterCount
	}
	// query today user purchase count
	newToday, renewalToday, err := l.svcCtx.OrderModel.QueryDateUserCounts(l.ctx, now)
	if err != nil {
		l.Errorw("[QueryUserStatisticsLogic] QueryDateUserCounts error", logger.Field("error", err.Error()))
	} else {
		resp.Today.NewOrderUsers = newToday
		resp.Today.RenewalOrderUsers = renewalToday
	}
	// query month user register count
	monthUserResisterCount, err := l.svcCtx.UserModel.QueryResisterUserTotalByMonthly(l.ctx, now)
	if err != nil {
		l.Errorw("[QueryUserStatisticsLogic] QueryResisterUserTotalByMonthly error", logger.Field("error", err.Error()))
	} else {
		resp.Monthly.Register = monthUserResisterCount
	}
	// query month user purchase count
	newMonth, renewalMonth, err := l.svcCtx.OrderModel.QueryMonthlyUserCounts(l.ctx, now)
	if err != nil {
		l.Errorw("[QueryUserStatisticsLogic] QueryMonthlyUserCounts error", logger.Field("error", err.Error()))
	} else {
		resp.Monthly.NewOrderUsers = newMonth
		resp.Monthly.RenewalOrderUsers = renewalMonth
		// TODO: Check the purchase status in the past seven days
		resp.Monthly.List = make([]types.UserStatistics, 0)
	}

	// query all user count
	allUserCount, err := l.svcCtx.UserModel.QueryResisterUserTotal(l.ctx)
	if err != nil {
		l.Errorw("[QueryUserStatisticsLogic] QueryResisterUserTotal error", logger.Field("error", err.Error()))
	} else {
		resp.All.Register = allUserCount
	}

	// query all user order counts
	allNewOrderUsers, allRenewalOrderUsers, err := l.svcCtx.OrderModel.QueryTotalUserCounts(l.ctx)
	if err != nil {
		l.Errorw("[QueryUserStatisticsLogic] QueryTotalUserCounts error", logger.Field("error", err.Error()))
	} else {
		resp.All.NewOrderUsers = allNewOrderUsers
		resp.All.RenewalOrderUsers = allRenewalOrderUsers
	}
	return
}

func (l *QueryUserStatisticsLogic) mockRevenueStatistics() *types.UserStatisticsResponse {
	now := time.Now()

	// Generate daily user statistics for the past 7 days (oldest first)
	monthlyList := make([]types.UserStatistics, 7)
	for i := 0; i < 7; i++ {
		dayDate := now.AddDate(0, 0, -(6 - i))
		baseRegister := int64(18 + ((6 - i) * 3) + ((6-i)%3)*8)
		monthlyList[i] = types.UserStatistics{
			Date:              dayDate.Format("2006-01-02"),
			Register:          baseRegister,
			NewOrderUsers:     int64(float64(baseRegister) * 0.65),
			RenewalOrderUsers: int64(float64(baseRegister) * 0.35),
		}
	}

	return &types.UserStatisticsResponse{
		Today: types.UserStatistics{
			Register:          28,
			NewOrderUsers:     18,
			RenewalOrderUsers: 10,
		},
		Monthly: types.UserStatistics{
			Register:          888,
			NewOrderUsers:     588,
			RenewalOrderUsers: 300,
			List:              monthlyList,
		},
		All: types.UserStatistics{
			Register:          18888,
			NewOrderUsers:     0, // This field is not used in All statistics
			RenewalOrderUsers: 0, // This field is not used in All statistics
		},
	}
}
