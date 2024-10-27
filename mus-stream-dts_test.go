package dts

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"testing"

	com "github.com/mus-format/common-go"
	muss "github.com/mus-format/mus-stream-go"
	"github.com/mus-format/mus-stream-go/ord"
	muss_mock "github.com/mus-format/mus-stream-go/testdata/mock"
	"github.com/mus-format/mus-stream-go/varint"
)

const FooDTM com.DTM = 1

type Foo struct {
	num int
	str string
}

func MarshalFoo(foo Foo, w muss.Writer) (n int, err error) {
	n, err = varint.MarshalInt(foo.num, w)
	if err != nil {
		return
	}
	var n1 int
	n1, err = ord.MarshalString(foo.str, nil, w)
	n += n1
	return
}

func UnmarshalFoo(r muss.Reader) (foo Foo, n int, err error) {
	foo.num, n, err = varint.UnmarshalInt(r)
	if err != nil {
		return
	}
	var n1 int
	foo.str, n1, err = ord.UnmarshalString(nil, r)
	n += n1
	return
}

func SizeFoo(foo Foo) (size int) {
	size = varint.SizeInt(foo.num)
	return size + ord.SizeString(foo.str, nil)
}

func SkipFoo(r muss.Reader) (n int, err error) {
	n, err = varint.SkipInt(r)
	if err != nil {
		return
	}
	var n1 int
	n1, err = ord.SkipString(nil, r)
	n += n1
	return
}

func TestDTS(t *testing.T) {

	t.Run("Marshal, Unmarshal, Size, Skip methods should work correctly",
		func(t *testing.T) {
			var (
				foo    = Foo{num: 11, str: "hello world"}
				fooDTS = New[Foo](FooDTM,
					muss.MarshallerFn[Foo](MarshalFoo),
					muss.UnmarshallerFn[Foo](UnmarshalFoo),
					muss.SizerFn[Foo](SizeFoo),
					muss.SkipperFn(SkipFoo),
				)
				size = fooDTS.Size(foo)
				buf  = bytes.NewBuffer(make([]byte, 0, size))
			)
			n, err := fooDTS.Marshal(foo, buf)
			if err != nil {
				t.Fatal(err)
			}
			if n != size {
				t.Fatalf("unexpected n, want '%v' actual '%v'", size, n)
			}
			afoo, n, err := fooDTS.Unmarshal(buf)
			if err != nil {
				t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
			}
			if n != size {
				t.Errorf("unexpected n, want '%v' actual '%v'", size, n)
			}
			if !reflect.DeepEqual(afoo, foo) {
				t.Errorf("unexpected afoo, want '%v' actual '%v'", foo, afoo)
			}
			buf.Reset()
			fooDTS.Marshal(foo, buf)
			n, err = fooDTS.Skip(buf)
			if err != nil {
				t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
			}
			if n != size {
				t.Errorf("unexpected n, want '%v' actual '%v'", size, n)
			}
		})

	t.Run("Marshal, UnmarshalDTM, UnmarshalData, Size, SkipDTM, SkipData methods should work correctly",
		func(t *testing.T) {
			var (
				wantDTSize = 1
				foo        = Foo{num: 11, str: "hello world"}
				fooDTS     = New[Foo](FooDTM,
					muss.MarshallerFn[Foo](MarshalFoo),
					muss.UnmarshallerFn[Foo](UnmarshalFoo),
					muss.SizerFn[Foo](SizeFoo),
					muss.SkipperFn(SkipFoo),
				)
				size = fooDTS.Size(foo)
				buf  = bytes.NewBuffer(make([]byte, 0, size))
			)
			n, err := fooDTS.Marshal(foo, buf)
			if err != nil {
				t.Fatal(err)
			}
			if n != size {
				t.Fatalf("unexpected n, want '%v' actual '%v'", size, n)
			}
			dtm, n, err := UnmarshalDTM(buf)
			if err != nil {
				t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
			}
			if n != wantDTSize {
				t.Errorf("unexpected n, want '%v' actual '%v'", 1, n)
			}
			if dtm != FooDTM {
				t.Errorf("unexpected dtm, want '%v' actual '%v'", FooDTM, dtm)
			}
			afoo, n, err := fooDTS.UnmarshalData(buf)
			if err != nil {
				t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
			}
			if n != size-wantDTSize {
				t.Errorf("unexpected n, want '%v' actual '%v'", size, n)
			}
			if !reflect.DeepEqual(afoo, foo) {
				t.Errorf("unexpected afoo, want '%v' actual '%v'", foo, afoo)
			}
			buf.Reset()
			fooDTS.Marshal(foo, buf)
			_, err = SkipDTM(buf)
			if err != nil {
				t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
			}
			n, err = fooDTS.SkipData(buf)
			if err != nil {
				t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
			}
			if n != size-wantDTSize {
				t.Errorf("unexpected n, want '%v' actual '%v'", size, n)
			}
		})

	t.Run("DTM method should return correct DTM", func(t *testing.T) {
		var (
			fooDTS = New[Foo](FooDTM, nil, nil, nil, nil)
			dtm    = fooDTS.DTM()
		)
		if dtm != FooDTM {
			t.Errorf("unexpected dtm, want '%v' actual '%v'", FooDTM, dtm)
		}
	})

	t.Run("Unamrshal should fail with ErrWrongDTM, if meets another DTM",
		func(t *testing.T) {
			var (
				wantDTSize = 1
				r          = muss_mock.NewReader().RegisterReadByte(
					func() (b byte, err error) {
						b = byte(FooDTM) + 3
						return
					},
				)
				fooDTS      = New[Foo](FooDTM, nil, nil, nil, nil)
				foo, n, err = fooDTS.Unmarshal(r)
			)
			if err != ErrWrongDTM {
				t.Errorf("unexpected error, want '%v' actual '%v'", ErrWrongDTM, err)
			}
			if !reflect.DeepEqual(foo, Foo{}) {
				t.Errorf("unexpected foo, want '%v' actual '%v'", nil, foo)
			}
			if n != wantDTSize {
				t.Errorf("unexpected n, want '%v' actual '%v'", wantDTSize, n)
			}
		})

	t.Run("Skip should fail with ErrWrongDTM, if meets another DTM",
		func(t *testing.T) {
			var (
				wantDTSize = 1
				r          = muss_mock.NewReader().RegisterReadByte(
					func() (b byte, err error) {
						b = byte(FooDTM) + 3
						return
					},
				)
				fooDTS = New[Foo](FooDTM, nil, nil, nil, nil)
				n, err = fooDTS.Skip(r)
			)
			if err != ErrWrongDTM {
				t.Errorf("unexpected error, want '%v' actual '%v'", ErrWrongDTM, err)
			}
			if n != wantDTSize {
				t.Errorf("unexpected n, want '%v' actual '%v'", wantDTSize, n)
			}
		})

	t.Run("If MarshalDTM fails with an error, Marshal should return it",
		func(t *testing.T) {
			var (
				wantErr = errors.New("write byte error")
				w       = muss_mock.NewWriter().RegisterWriteByte(func(c byte) error {
					return wantErr
				})
				fooDTS = New[Foo](FooDTM, nil, nil, nil, nil)
			)
			_, err := fooDTS.Marshal(Foo{}, w)
			if err != wantErr {
				t.Errorf("unexpected error, want '%v' actual '%v'", wantErr, err)
			}
		})

	t.Run("If UnmarshalDTM fails with an error, Unmarshal should return it",
		func(t *testing.T) {
			var (
				wantErr = errors.New("read byte error")
				r       = muss_mock.NewReader().RegisterReadByte(
					func() (b byte, err error) {
						err = wantErr
						return
					},
				)
				fooDTS      = New[Foo](FooDTM, nil, nil, nil, nil)
				foo, n, err = fooDTS.Unmarshal(r)
			)
			if err != wantErr {
				t.Errorf("unexpected error, want '%v' actual '%v'", io.EOF, err)
			}
			if !reflect.DeepEqual(foo, Foo{}) {
				t.Errorf("unexpected foo, want '%v' actual '%v'", nil, foo)
			}
			if n != 0 {
				t.Errorf("unexpected n, want '%v' actual '%v'", 0, n)
			}
		})

	t.Run("If UnmarshalDTM fails with an error, Skip should return it",
		func(t *testing.T) {
			var (
				wantErr = errors.New("read byte error")
				r       = muss_mock.NewReader().RegisterReadByte(
					func() (b byte, err error) {
						err = wantErr
						return
					},
				)
				fooDTS = New[Foo](FooDTM, nil, nil, nil, nil)
				n, err = fooDTS.Skip(r)
			)
			if err != wantErr {
				t.Errorf("unexpected error, want '%v' actual '%v'", io.EOF, err)
			}
			if n != 0 {
				t.Errorf("unexpected n, want '%v' actual '%v'", 0, n)
			}
		})

	t.Run("If varint.UnmarshalInt fails with an error, UnmarshalDTM should return it",
		func(t *testing.T) {
			var (
				wantErr = errors.New("read byte error")
				r       = muss_mock.NewReader().RegisterReadByte(
					func() (b byte, err error) {
						err = wantErr
						return
					},
				)
				dtm, n, err = UnmarshalDTM(r)
			)
			if err != wantErr {
				t.Errorf("unexpected error, want '%v' actual '%v'", io.EOF, err)
			}
			if dtm != 0 {
				t.Errorf("unexpected dtm, want '%v' actual '%v'", 0, dtm)
			}
			if n != 0 {
				t.Errorf("unexpected n, want '%v' actual '%v'", 0, n)
			}
		})

}
