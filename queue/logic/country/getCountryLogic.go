package countrylogic

import (
	"context"
	"encoding/json"

	"github.com/perfect-panel/ppanel-server/pkg/logger"

	"github.com/hibiken/asynq"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/pkg/ip"
	"github.com/perfect-panel/ppanel-server/queue/types"
)

type GetNodeCountryLogic struct {
	svcCtx *svc.ServiceContext
}

func NewGetNodeCountryLogic(svcCtx *svc.ServiceContext) *GetNodeCountryLogic {
	return &GetNodeCountryLogic{
		svcCtx: svcCtx,
	}
}
func (l *GetNodeCountryLogic) ProcessTask(ctx context.Context, task *asynq.Task) error {
	var payload types.GetNodeCountry
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		logger.WithContext(ctx).Error("[GetNodeCountryLogic] Unmarshal payload failed",
			logger.Field("error", err.Error()),
			logger.Field("payload", task.Payload()),
		)
		return nil
	}
	serverAddr := payload.ServerAddr
	resp, err := ip.GetRegionByIp(serverAddr)
	if err != nil {
		logger.WithContext(ctx).Error("[GetNodeCountryLogic] ", logger.Field("error", err.Error()), logger.Field("serverAddr", serverAddr))
		return nil
	}

	servers, err := l.svcCtx.ServerModel.FindNodeByServerAddrAndProtocol(ctx, payload.ServerAddr, payload.Protocol)
	if err != nil {
		logger.WithContext(ctx).Error("[GetNodeCountryLogic] FindNodeByServerAddrAnd", logger.Field("error", err.Error()), logger.Field("serverAddr", serverAddr))
		return err
	}
	if len(servers) == 0 {
		return nil
	}
	for _, ser := range servers {
		ser.Country = resp.Country
		ser.City = resp.City
		ser.Latitude = resp.Latitude
		ser.Longitude = resp.Longitude
		err := l.svcCtx.ServerModel.Update(ctx, ser)
		if err != nil {
			logger.WithContext(ctx).Error("[GetNodeCountryLogic] ", logger.Field("error", err.Error()), logger.Field("id", ser.Id))
		}
	}
	logger.WithContext(ctx).Info("[GetNodeCountryLogic] ", logger.Field("country", resp.Country), logger.Field("city", resp.Country))
	return nil
}
