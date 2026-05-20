package system

import (
	"context"
	"reflect"

	"github.com/perfect-panel/server/internal/repository"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/tool"
)

type configFieldValue struct {
	key   string
	value string
}

func convertedConfigFields(data any) []configFieldValue {
	return configFields(data, tool.ConvertValueToString)
}

func stringConfigFields(data any) []configFieldValue {
	return configFields(data, func(value reflect.Value) string {
		return value.String()
	})
}

func configFields(data any, valueFn func(reflect.Value) string) []configFieldValue {
	v := reflect.ValueOf(data)
	t := v.Type()
	fields := make([]configFieldValue, 0, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		fields = append(fields, configFieldValue{
			key:   t.Field(i).Name,
			value: valueFn(v.Field(i)),
		})
	}
	return fields
}

func updateConfigFields(ctx context.Context, svcCtx *svc.ServiceContext, category string, fields []configFieldValue, cacheKeys ...string) error {
	return svcCtx.Store.InTx(ctx, func(store repository.Store) error {
		systemStore := store.System()
		for _, field := range fields {
			if err := systemStore.UpdateValueByCategoryKey(ctx, category, field.key, field.value); err != nil {
				return err
			}
		}
		if len(cacheKeys) == 0 {
			return nil
		}
		return svcCtx.Redis.Del(ctx, cacheKeys...).Err()
	})
}
