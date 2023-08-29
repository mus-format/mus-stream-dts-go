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

// DTS provides data type metadata support for the mus-go serializer.
//
// It implements muss.Marshaller, muss.Unmarshaller and muss.Sizer interfaces.
type DTS[T any] struct {
	dtm com.DTM

	m muss.Marshaller[T]
	u muss.Unmarshaller[T]
	s muss.Sizer[T]
}

// DTM returns a data type metadata.
func (dts DTS[T]) DTM() com.DTM {
	return dts.dtm
}

// MarshalMUS marshals DTM and data to the MUS format.
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

// UnmarshalMUS unmarshals DTM and data from the MUS format.
//
// Returns ErrWrongDTM if DTM from bs is different.
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

// SizeMUS calculates the DTM and data size in the MUS format.
func (dts DTS[T]) SizeMUS(t T) (size int) {
	size = SizeDTMUS(dts.dtm)
	return size + dts.s.SizeMUS(t)
}

// UnmarshalMUS unmarshals data without DTM from the MUS format.
func (dts DTS[T]) UnmarshalDataMUS(r muss.Reader) (t T, n int, err error) {
	return dts.u.UnmarshalMUS(r)
}
