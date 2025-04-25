package traffic

import "fmt"

const (
	bitsInTiB = 1024.0 * 1024.0 * 1024.0 * 1024.0
	bitsInTb  = 1000.0 * 1000.0 * 1000.0 * 1000.0
	bitsInGiB = 1024.0 * 1024.0 * 1024.0
	bitsInGb  = 1000.0 * 1000.0 * 1000.0
	bitsInMiB = 1024.0 * 1024.0
	bitsInMb  = 1000.0 * 1000.0

	Mb  = "Mb"
	MiB = "MiB"
	Gb  = "Gb"
	GiB = "GiB"
	Tb  = "Tb"
	TiB = "TiB"
)

// Convert converts the given traffic in bits to the specified unit ("MiB" or "GiB")
func Convert(bits int64, unit string) float64 {
	switch unit {
	case Mb:
		return float64(bits) / bitsInMb
	case MiB:
		return float64(bits) / bitsInMiB
	case GiB:
		return float64(bits) / bitsInGiB
	case Gb:
		return float64(bits) / bitsInGb
	case TiB:
		return float64(bits) / bitsInTiB
	case Tb:
		return float64(bits) / bitsInTb
	default:
		return 0 // Return 0 for unsupported units
	}
}

// AutoConvert converts the given traffic in bits to the largest possible unit and returns a formatted string
func AutoConvert(bits int64, useBinary bool) string {
	if useBinary {
		switch {
		case float64(bits) >= bitsInTiB:
			return fmt.Sprintf("%.2f TiB", float64(bits)/bitsInTiB)
		case float64(bits) >= bitsInGiB:
			return fmt.Sprintf("%.2f GiB", float64(bits)/bitsInGiB)
		case float64(bits) >= bitsInMiB:
			return fmt.Sprintf("%.2f MiB", float64(bits)/bitsInMiB)
		default:
			return fmt.Sprintf("%d bits", bits)
		}
	} else {
		switch {
		case float64(bits) >= bitsInTb:
			return fmt.Sprintf("%.2f Tb", float64(bits)/bitsInTb)
		case float64(bits) >= bitsInGb:
			return fmt.Sprintf("%.2f Gb", float64(bits)/bitsInGb)
		case float64(bits) >= bitsInMb:
			return fmt.Sprintf("%.2f Mb", float64(bits)/bitsInMb)
		default:
			return fmt.Sprintf("%d bits", bits)
		}
	}
}
