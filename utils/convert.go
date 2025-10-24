package utils

type Bytes int64

func (b Bytes) ToKiB() float64 {
	return float64(b / 1024)
}

func (b Bytes) ToMiB() float64 {
	return float64(b / 1024 / 1024)
}

func (b Bytes) ToGiB() float64 {
	return float64(b / 1024 / 1024 / 1024)
}

func (b Bytes) ToTiB() float64 {
	return float64(b / 1024 / 1024 / 1024 / 1024)
}

func (b Bytes) ToPiB() float64 {
	return float64(b / 1024 / 1024 / 1024 / 1024 / 1024)
}
