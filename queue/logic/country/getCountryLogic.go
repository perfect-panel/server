package countrylogic

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/perfect-panel/server/internal/svc"
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

	return nil
}
