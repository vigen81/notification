package types

import (
	"database/sql/driver"
	"fmt"
)

type Address string

func (a Address) Value() (driver.Value, error) {
	return string(a), nil
}

func (a *Address) Scan(src interface{}) error {
	switch v := src.(type) {
	case string:
		*a = Address(v)
		return nil
	case []byte:
		*a = Address(v)
		return nil
	default:
		return fmt.Errorf("cannot scan %T into Address", src)
	}
}
