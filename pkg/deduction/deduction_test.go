package deduction

import (
	"math"
	"testing"
	"time"
)

func TestSubscribe_Validate(t *testing.T) {
	tests := []struct {
		name    string
		sub     Subscribe
		wantErr bool
		errType error
	}{
		{
			name: "valid subscription",
			sub: Subscribe{
				StartTime:      time.Now(),
				ExpireTime:     time.Now().Add(24 * time.Hour),
				Traffic:        1000,
				Download:       100,
				Upload:         200,
				UnitTime:       UnitTimeMonth,
				DeductionRatio: 50,
			},
			wantErr: false,
		},
		{
			name: "negative traffic",
			sub: Subscribe{
				StartTime:      time.Now(),
				ExpireTime:     time.Now().Add(24 * time.Hour),
				Traffic:        -1000,
				Download:       100,
				Upload:         200,
				UnitTime:       UnitTimeMonth,
				DeductionRatio: 50,
			},
			wantErr: true,
			errType: ErrInvalidTraffic,
		},
		{
			name: "negative download",
			sub: Subscribe{
				StartTime:      time.Now(),
				ExpireTime:     time.Now().Add(24 * time.Hour),
				Traffic:        1000,
				Download:       -100,
				Upload:         200,
				UnitTime:       UnitTimeMonth,
				DeductionRatio: 50,
			},
			wantErr: true,
			errType: ErrInvalidTraffic,
		},
		{
			name: "download + upload exceeds traffic",
			sub: Subscribe{
				StartTime:      time.Now(),
				ExpireTime:     time.Now().Add(24 * time.Hour),
				Traffic:        1000,
				Download:       600,
				Upload:         500,
				UnitTime:       UnitTimeMonth,
				DeductionRatio: 50,
			},
			wantErr: true,
		},
		{
			name: "expire time before start time",
			sub: Subscribe{
				StartTime:      time.Now(),
				ExpireTime:     time.Now().Add(-24 * time.Hour),
				Traffic:        1000,
				Download:       100,
				Upload:         200,
				UnitTime:       UnitTimeMonth,
				DeductionRatio: 50,
			},
			wantErr: true,
			errType: ErrInvalidTimeRange,
		},
		{
			name: "invalid deduction ratio - negative",
			sub: Subscribe{
				StartTime:      time.Now(),
				ExpireTime:     time.Now().Add(24 * time.Hour),
				Traffic:        1000,
				Download:       100,
				Upload:         200,
				UnitTime:       UnitTimeMonth,
				DeductionRatio: -10,
			},
			wantErr: true,
			errType: ErrInvalidDeductionRatio,
		},
		{
			name: "invalid deduction ratio - over 100",
			sub: Subscribe{
				StartTime:      time.Now(),
				ExpireTime:     time.Now().Add(24 * time.Hour),
				Traffic:        1000,
				Download:       100,
				Upload:         200,
				UnitTime:       UnitTimeMonth,
				DeductionRatio: 150,
			},
			wantErr: true,
			errType: ErrInvalidDeductionRatio,
		},
		{
			name: "invalid unit time",
			sub: Subscribe{
				StartTime:      time.Now(),
				ExpireTime:     time.Now().Add(24 * time.Hour),
				Traffic:        1000,
				Download:       100,
				Upload:         200,
				UnitTime:       "InvalidUnit",
				DeductionRatio: 50,
			},
			wantErr: true,
			errType: ErrInvalidUnitTime,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.sub.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Subscribe.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.errType != nil && err != tt.errType {
				t.Errorf("Subscribe.Validate() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

func TestOrder_Validate(t *testing.T) {
	tests := []struct {
		name    string
		order   Order
		wantErr bool
		errType error
	}{
		{
			name:    "valid order",
			order:   Order{Amount: 1000, Quantity: 2},
			wantErr: false,
		},
		{
			name:    "zero quantity",
			order:   Order{Amount: 1000, Quantity: 0},
			wantErr: true,
			errType: ErrInvalidQuantity,
		},
		{
			name:    "negative quantity",
			order:   Order{Amount: 1000, Quantity: -1},
			wantErr: true,
			errType: ErrInvalidQuantity,
		},
		{
			name:    "negative amount",
			order:   Order{Amount: -1000, Quantity: 2},
			wantErr: true,
			errType: ErrInvalidAmount,
		},
		{
			name:    "zero amount is valid",
			order:   Order{Amount: 0, Quantity: 1},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.order.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Order.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.errType != nil && err != tt.errType {
				t.Errorf("Order.Validate() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

func TestSafeMultiply(t *testing.T) {
	tests := []struct {
		name    string
		a, b    int64
		want    int64
		wantErr bool
	}{
		{
			name:    "normal multiplication",
			a:       10,
			b:       20,
			want:    200,
			wantErr: false,
		},
		{
			name:    "zero multiplication",
			a:       10,
			b:       0,
			want:    0,
			wantErr: false,
		},
		{
			name:    "negative multiplication",
			a:       -10,
			b:       20,
			want:    -200,
			wantErr: false,
		},
		{
			name:    "overflow case",
			a:       math.MaxInt64,
			b:       2,
			want:    0,
			wantErr: true,
		},
		{
			name:    "large numbers no overflow",
			a:       1000000,
			b:       1000000,
			want:    1000000000000,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := safeMultiply(tt.a, tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("safeMultiply() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("safeMultiply() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSafeAdd(t *testing.T) {
	tests := []struct {
		name    string
		a, b    int64
		want    int64
		wantErr bool
	}{
		{
			name:    "normal addition",
			a:       10,
			b:       20,
			want:    30,
			wantErr: false,
		},
		{
			name:    "negative addition",
			a:       -10,
			b:       5,
			want:    -5,
			wantErr: false,
		},
		{
			name:    "overflow case",
			a:       math.MaxInt64,
			b:       1,
			want:    0,
			wantErr: true,
		},
		{
			name:    "underflow case",
			a:       math.MinInt64,
			b:       -1,
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := safeAdd(tt.a, tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("safeAdd() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("safeAdd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSafeDivide(t *testing.T) {
	tests := []struct {
		name    string
		a, b    int64
		want    int64
		wantErr bool
	}{
		{
			name:    "normal division",
			a:       20,
			b:       10,
			want:    2,
			wantErr: false,
		},
		{
			name:    "division by zero",
			a:       20,
			b:       0,
			want:    0,
			wantErr: true,
		},
		{
			name:    "negative division",
			a:       -20,
			b:       10,
			want:    -2,
			wantErr: false,
		},
		{
			name:    "zero dividend",
			a:       0,
			b:       10,
			want:    0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := safeDivide(tt.a, tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("safeDivide() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("safeDivide() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateWeights(t *testing.T) {
	tests := []struct {
		name              string
		deductionRatio    int64
		wantTrafficWeight float64
		wantTimeWeight    float64
	}{
		{
			name:              "zero ratio",
			deductionRatio:    0,
			wantTrafficWeight: 0,
			wantTimeWeight:    0,
		},
		{
			name:              "50% ratio",
			deductionRatio:    50,
			wantTrafficWeight: 0.5,
			wantTimeWeight:    0.5,
		},
		{
			name:              "75% ratio",
			deductionRatio:    75,
			wantTrafficWeight: 0.75,
			wantTimeWeight:    0.25,
		},
		{
			name:              "100% ratio",
			deductionRatio:    100,
			wantTrafficWeight: 1.0,
			wantTimeWeight:    0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTrafficWeight, gotTimeWeight := calculateWeights(tt.deductionRatio)
			if gotTrafficWeight != tt.wantTrafficWeight {
				t.Errorf("calculateWeights() trafficWeight = %v, want %v", gotTrafficWeight, tt.wantTrafficWeight)
			}
			if gotTimeWeight != tt.wantTimeWeight {
				t.Errorf("calculateWeights() timeWeight = %v, want %v", gotTimeWeight, tt.wantTimeWeight)
			}
		})
	}
}

func TestCalculateProportionalAmount(t *testing.T) {
	tests := []struct {
		name      string
		unitPrice int64
		remaining int64
		total     int64
		want      int64
		wantErr   bool
	}{
		{
			name:      "normal calculation",
			unitPrice: 100,
			remaining: 50,
			total:     100,
			want:      50,
			wantErr:   false,
		},
		{
			name:      "zero total",
			unitPrice: 100,
			remaining: 50,
			total:     0,
			want:      0,
			wantErr:   false,
		},
		{
			name:      "zero remaining",
			unitPrice: 100,
			remaining: 0,
			total:     100,
			want:      0,
			wantErr:   false,
		},
		{
			name:      "quarter remaining",
			unitPrice: 200,
			remaining: 25,
			total:     100,
			want:      50,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculateProportionalAmount(tt.unitPrice, tt.remaining, tt.total)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateProportionalAmount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("calculateProportionalAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateNoLimitAmount(t *testing.T) {
	tests := []struct {
		name    string
		sub     Subscribe
		order   Order
		want    int64
		wantErr bool
	}{
		{
			name: "normal no limit calculation",
			sub: Subscribe{
				Traffic:  1000,
				Download: 300,
				Upload:   200,
			},
			order: Order{
				Amount: 1000,
			},
			want:    500, // (1000 - 300 - 200) / 1000 * 1000 = 500
			wantErr: false,
		},
		{
			name: "zero traffic",
			sub: Subscribe{
				Traffic:  0,
				Download: 0,
				Upload:   0,
			},
			order: Order{
				Amount: 1000,
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "overused traffic",
			sub: Subscribe{
				Traffic:  1000,
				Download: 600,
				Upload:   500,
			},
			order: Order{
				Amount: 1000,
			},
			want:    0, // usedTraffic would be negative, clamped to 0
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculateNoLimitAmount(tt.sub, tt.order)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateNoLimitAmount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("calculateNoLimitAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateRemainingAmount(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		sub     Subscribe
		order   Order
		wantErr bool
	}{
		{
			name: "valid no limit subscription",
			sub: Subscribe{
				StartTime:      now.Add(-24 * time.Hour),
				ExpireTime:     now.Add(24 * time.Hour),
				Traffic:        1000,
				Download:       300,
				Upload:         200,
				UnitTime:       UnitTimeNoLimit,
				ResetCycle:     ResetCycleNone,
				DeductionRatio: 0,
			},
			order: Order{
				Amount:   1000,
				Quantity: 1,
			},
			wantErr: false,
		},
		{
			name: "invalid subscription",
			sub: Subscribe{
				StartTime:      now,
				ExpireTime:     now.Add(-24 * time.Hour), // Invalid: expire before start
				Traffic:        1000,
				Download:       300,
				Upload:         200,
				UnitTime:       UnitTimeMonth,
				DeductionRatio: 0,
			},
			order: Order{
				Amount:   1000,
				Quantity: 1,
			},
			wantErr: true,
		},
		{
			name: "invalid order",
			sub: Subscribe{
				StartTime:      now.Add(-24 * time.Hour),
				ExpireTime:     now.Add(24 * time.Hour),
				Traffic:        1000,
				Download:       300,
				Upload:         200,
				UnitTime:       UnitTimeMonth,
				DeductionRatio: 0,
			},
			order: Order{
				Amount:   1000,
				Quantity: 0, // Invalid: zero quantity
			},
			wantErr: true,
		},
		{
			name: "no limit with reset cycle",
			sub: Subscribe{
				StartTime:      now.Add(-24 * time.Hour),
				ExpireTime:     now.Add(24 * time.Hour),
				Traffic:        1000,
				Download:       300,
				Upload:         200,
				UnitTime:       UnitTimeNoLimit,
				ResetCycle:     ResetCycleMonthly, // Should return 0
				DeductionRatio: 0,
			},
			order: Order{
				Amount:   1000,
				Quantity: 1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CalculateRemainingAmount(tt.sub, tt.order)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateRemainingAmount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCalculateRemainingAmount_NoLimitWithResetCycle(t *testing.T) {
	now := time.Now()
	sub := Subscribe{
		StartTime:      now.Add(-24 * time.Hour),
		ExpireTime:     now.Add(24 * time.Hour),
		Traffic:        1000,
		Download:       300,
		Upload:         200,
		UnitTime:       UnitTimeNoLimit,
		ResetCycle:     ResetCycleMonthly,
		DeductionRatio: 0,
	}
	order := Order{
		Amount:   1000,
		Quantity: 1,
	}

	got, err := CalculateRemainingAmount(sub, order)
	if err != nil {
		t.Errorf("CalculateRemainingAmount() error = %v", err)
		return
	}
	if got != 0 {
		t.Errorf("CalculateRemainingAmount() = %v, want 0", got)
	}
}

// Benchmark tests
func BenchmarkCalculateRemainingAmount(b *testing.B) {
	now := time.Now()
	sub := Subscribe{
		StartTime:      now.Add(-24 * time.Hour),
		ExpireTime:     now.Add(24 * time.Hour),
		Traffic:        1000,
		Download:       300,
		Upload:         200,
		UnitTime:       UnitTimeMonth,
		ResetCycle:     ResetCycleNone,
		DeductionRatio: 50,
	}
	order := Order{
		Amount:   1000,
		Quantity: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = CalculateRemainingAmount(sub, order)
	}
}

func BenchmarkSafeMultiply(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = safeMultiply(12345, 67890)
	}
}
