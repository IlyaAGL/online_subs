package errormsgs

import "errors"

var NotFound = errors.New("not found :(")

func IsNotFound(err error) bool {
	return errors.Is(err, NotFound)
}