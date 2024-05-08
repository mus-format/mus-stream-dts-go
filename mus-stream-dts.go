package dts

import (
	com "github.com/mus-format/common-go"
	muss "github.com/mus-format/mus-stream-go"
)

// New creates a new DTS.
func New[T any](dtm com.DTM, m muss.Marshaller[T], u muss.Unmarshaller[T],
	s muss.Sizer[T]) DTS[T] {
	return DTS[T]{dtm, m, u, s}
}

// DTS provides data type metadata (DTM) support for the mus-stream-go
// serializer. It helps to encode DTM + data.
//
// Implements muss.Marshaller, muss.Unmarshaller and muss.Sizer interfaces.
type DTS[T any] struct {
	dtm com.DTM
	m   muss.Marshaller[T]
	u   muss.Unmarshaller[T]
	s   muss.Sizer[T]
}

// DTM returns the value with which DTS was initialized.
func (dts DTS[T]) DTM() com.DTM {
	return dts.dtm
}

// MarshalMUS marshals DTM + data.
func (dts DTS[T]) MarshalMUS(t T, w muss.Writer) (n int, err error) {
	n, err = MarshalDTMUS(dts.dtm, w)
	if err != nil {
		return
	}
	var n1 int
	n1, err = dts.m.MarshalMUS(t, w)
	n += n1
	return
}

// UnmarshalMUS unmarshals DTM + data.
//
// Returns ErrWrongDTM if the unmarshalled DTM differs from the dts.DTM().
func (dts DTS[T]) UnmarshalMUS(r muss.Reader) (t T, n int, err error) {
	dtm, n, err := UnmarshalDTMUS(r)
	if err != nil {
		return
	}
	if dtm != dts.dtm {
		err = ErrWrongDTM
		return
	}
	var n1 int
	t, n1, err = dts.UnmarshalDataMUS(r)
	n += n1
	return
}

// SizeMUS calculates the size of the DTM + data.
func (dts DTS[T]) SizeMUS(t T) (size int) {
	size = SizeDTMUS(dts.dtm)
	return size + dts.s.SizeMUS(t)
}

// UnmarshalData unmarshals only data.
func (dts DTS[T]) UnmarshalDataMUS(r muss.Reader) (t T, n int, err error) {
	return dts.u.UnmarshalMUS(r)
}
