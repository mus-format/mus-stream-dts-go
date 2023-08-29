package dts

import "errors"

// ErrWrongDTM happens when DTS tries to unmarshal data from bs with wrong DTM.
var ErrWrongDTM = errors.New("wrong data type metadata")
