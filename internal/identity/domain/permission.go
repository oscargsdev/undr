package domain

import "slices"

type Roles []string

func (r Roles) Include(code string) bool {
	return slices.Contains(r, code)
}
