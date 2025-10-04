package repository

type SortDirection int

const (
	SortAscending  SortDirection = 1
	SortDescending SortDirection = -1
)

type SortOption struct {
	Field     string
	Direction SortDirection
}

func (o *SortOption) apply(defaultField string, defaultDir SortDirection) (string, SortDirection) {
	if o == nil || o.Field == "" {
		return defaultField, defaultDir
	}
	dir := o.Direction
	if dir != SortAscending && dir != SortDescending {
		dir = defaultDir
	}
	return o.Field, dir
}
