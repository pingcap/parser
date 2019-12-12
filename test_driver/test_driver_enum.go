package test_driver

// Enum is for MySQL enum type.
type Enum struct {
	Name  string
	Value uint64
}

// String implements fmt.Stringer interface.
func (e Enum) String() string {
	return e.Name
}

// ToNumber changes enum index to float64 for numeric operation.
func (e Enum) ToNumber() float64 {
	return float64(e.Value)
}
