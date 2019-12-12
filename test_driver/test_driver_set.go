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
