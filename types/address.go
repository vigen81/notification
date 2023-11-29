package types

import (
	`database/sql/driver`
	`gitlab.com/healthcare-integration/golang/notification-service/service/crypt`
)

type Address string

func NewAddress(s string) Address {
	return Address(s)
}
func (a Address) String() string {
	return string(a)
}

func (a Address) Value() (driver.Value, error) {
	return crypt.Encode(string(a))
}

func (a *Address) Scan(src interface{}) error {
	decode, err := crypt.Decode(string(src.([]uint8)))
	if nil != err {
		return err
	}
	*a = Address(decode)
	return nil
}
