package test_driver

// Set is for MySQL Set type.
type Set struct {
	Name  string
	Value uint64
}

// String implements fmt.Stringer interface.
func (e Set) String() string {
	return e.Name
}

// ToNumber changes Set to float64 for numeric operation.
func (e Set) ToNumber() float64 {
	return float64(e.Value)
}

var (
	setIndexValue       []uint64
	setIndexInvertValue []uint64
)

func init() {
	setIndexValue = make([]uint64, 64)
	setIndexInvertValue = make([]uint64, 64)

	for i := 0; i < 64; i++ {
		setIndexValue[i] = 1 << uint64(i)
		setIndexInvertValue[i] = ^setIndexValue[i]
	}
}
