package order

import (
	"github.com/perfect-panel/ppanel-server/internal/model/coupon"
)

func calculateCoupon(amount int64, couponInfo *coupon.Coupon) int64 {
	if couponInfo.Type == 1 {
		return int64(float64(amount) * (float64(couponInfo.Discount) / float64(100)))
	} else {
		return min(couponInfo.Discount, amount)
	}
}
