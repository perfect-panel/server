package user

import (
	"context"
	"time"

	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"

	"github.com/perfect-panel/server/internal/model/traffic"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/constant"
	"github.com/perfect-panel/server/pkg/logger"
	"gorm.io/gorm"
)

type GetUserSubscribeTrafficLogsLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get user subcribe traffic logs
func NewGetUserSubscribeTrafficLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserSubscribeTrafficLogsLogic {
	return &GetUserSubscribeTrafficLogsLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserSubscribeTrafficLogsLogic) GetUserSubscribeTrafficLogs(req *types.GetUserSubscribeTrafficLogsRequest) (resp *types.GetUserSubscribeTrafficLogsResponse, err error) {
	resp = &types.GetUserSubscribeTrafficLogsResponse{}
	u := l.ctx.Value(constant.CtxKeyUser).(*user.User)
	var traffics []traffic.TrafficLog
	err = l.svcCtx.UserModel.Transaction(l.ctx, func(db *gorm.DB) error {
		return db.Model(traffic.TrafficLog{}).Where("user_id = ? and `timestamp` >= ? and `timestamp` < ?", u.Id, time.UnixMilli(req.StartTime), time.UnixMilli(req.EndTime)).Find(&traffics).Error
	})

	if err != nil {
		l.Errorw("get user subscribe traffic logs failed", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), err.Error())
	}

	//合并多条记录为以天为单位
	trafficMap := make(map[string]*traffic.TrafficLog)
	for _, traf := range traffics {
		key := traf.Timestamp.Format(time.DateOnly)
		existTraf := trafficMap[key]
		if existTraf == nil {
			trafficMap[key] = &traf
		} else {
			existTraf.Upload = existTraf.Download + traf.Upload
			existTraf.Download = existTraf.Download + traf.Download
			trafficMap[key] = existTraf
		}
	}

	startTime := time.UnixMilli(req.StartTime)
	EndTime := time.UnixMilli(req.EndTime)
	res := make(map[string]traffic.TrafficLog)

	// 循环遍历每一天
	for current := startTime; !current.After(EndTime); current = current.AddDate(0, 0, 1) {
		dateStr := current.Format(time.DateOnly) // 格式化为日期字符串
		if trafficMap[dateStr] == nil {
			res[dateStr] = traffic.TrafficLog{
				Timestamp: current,
			}
		} else {
			res[dateStr] = *trafficMap[dateStr]
		}
		resp.List = append(resp.List, types.TrafficLog{
			Id:        res[dateStr].Id,
			ServerId:  res[dateStr].ServerId,
			Upload:    res[dateStr].Upload,
			Download:  res[dateStr].Download,
			Timestamp: res[dateStr].Timestamp.UnixMilli(),
		})
	}

	return
}
