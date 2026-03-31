package types

import (
	"database/sql/driver"
	"fmt"
)

// Address represents a notification address (email, phone, push token)
type Address string

// Value implements the driver.Valuer interface for database storage
func (a Address) Value() (driver.Value, error) {
	return string(a), nil
}

// Scan implements the sql.Scanner interface for database retrieval
func (a *Address) Scan(src interface{}) error {
	switch v := src.(type) {
	case string:
		*a = Address(v)
		return nil
	case []byte:
		*a = Address(v)
		return nil
	case nil:
		*a = ""
		return nil
	default:
		return fmt.Errorf("cannot scan %T into Address", src)
	}
}

// String returns the string representation of the address
func (a Address) String() string {
	return string(a)
}

// IsEmpty checks if the address is empty
func (a Address) IsEmpty() bool {
	return string(a) == ""
}
