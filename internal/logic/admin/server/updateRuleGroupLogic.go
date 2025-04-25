package server

import (
	"context"
	"strings"

	"github.com/perfect-panel/server/pkg/tool"

	"github.com/perfect-panel/server/internal/model/server"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
)

type UpdateRuleGroupLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUpdateRuleGroupLogic Update rule group
func NewUpdateRuleGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateRuleGroupLogic {
	return &UpdateRuleGroupLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateRuleGroupLogic) UpdateRuleGroup(req *types.UpdateRuleGroupRequest) error {
	rs, err := parseAndValidateRules(req.Rules, req.Name)
	if err != nil {
		return err
	}
	err = l.svcCtx.ServerModel.UpdateRuleGroup(l.ctx, &server.RuleGroup{
		Id:     req.Id,
		Icon:   req.Icon,
		Name:   req.Name,
		Tags:   tool.StringSliceToString(req.Tags),
		Rules:  strings.Join(rs, "\n"),
		Enable: req.Enable,
	})
	if err != nil {
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseUpdateError), err.Error())
	}
	return nil
}
