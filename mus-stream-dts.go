package dts

import (
	com "github.com/mus-format/common-go"
	muss "github.com/mus-format/mus-stream-go"
)

// New creates a new DTS.
func New[T any](dtm com.DTM, m muss.Marshaller[T], u muss.Unmarshaller[T],
	s muss.Sizer[T],
	sk muss.Skipper,
) DTS[T] {
	return DTS[T]{dtm, m, u, s, sk}
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
	sk  muss.Skipper
}

// DTM returns the value with which DTS was initialized.
func (dts DTS[T]) DTM() com.DTM {
	return dts.dtm
}

// Marshal marshals DTM + data.
func (dts DTS[T]) Marshal(t T, w muss.Writer) (n int, err error) {
	n, err = MarshalDTM(dts.dtm, w)
	if err != nil {
		return
	}
	var n1 int
	n1, err = dts.m.Marshal(t, w)
	n += n1
	return
}

// Unmarshal unmarshals DTM + data.
//
// Returns ErrWrongDTM if the unmarshalled DTM differs from the dts.DTM().
func (dts DTS[T]) Unmarshal(r muss.Reader) (t T, n int, err error) {
	dtm, n, err := UnmarshalDTM(r)
	if err != nil {
		return
	}
	if dtm != dts.dtm {
		err = ErrWrongDTM
		return
	}
	var n1 int
	t, n1, err = dts.UnmarshalData(r)
	n += n1
	return
}

// Size calculates the size of the DTM + data.
func (dts DTS[T]) Size(t T) (size int) {
	size = SizeDTM(dts.dtm)
	return size + dts.s.Size(t)
}

// Skip skips DTM + data.
//
// Returns ErrWrongDTM if the unmarshalled DTM differs from the dts.DTM().
func (dts DTS[T]) Skip(r muss.Reader) (n int, err error) {
	dtm, n, err := UnmarshalDTM(r)
	if err != nil {
		return
	}
	if dtm != dts.dtm {
		err = ErrWrongDTM
		return
	}
	n1, err := dts.SkipData(r)
	n += n1
	return
}

// UnmarshalData unmarshals only data.
func (dts DTS[T]) UnmarshalData(r muss.Reader) (t T, n int, err error) {
	return dts.u.Unmarshal(r)
}

// SkipData skips only data.
func (dts DTS[T]) SkipData(r muss.Reader) (n int, err error) {
	return dts.sk.Skip(r)
}
