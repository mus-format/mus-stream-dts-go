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

func MarshalFooMUS(foo Foo, w muss.Writer) (n int, err error) {
	n, err = varint.MarshalInt(foo.num, w)
	if err != nil {
		return
	}
	var n1 int
	n1, err = ord.MarshalString(foo.str, w)
	n += n1
	return
}

func UnmarshalFooMUS(r muss.Reader) (foo Foo, n int, err error) {
	foo.num, n, err = varint.UnmarshalInt(r)
	if err != nil {
		return
	}
	var n1 int
	foo.str, n1, err = ord.UnmarshalString(r)
	n += n1
	return
}

func SizeFooMUS(foo Foo) (size int) {
	size = varint.SizeInt(foo.num)
	return size + ord.SizeString(foo.str)
}

func TestDTS(t *testing.T) {

	t.Run("MarshalMUS, UnmarshalMUS, SizeMUS methods should work correctly",
		func(t *testing.T) {
			var (
				foo    = Foo{num: 11, str: "hello world"}
				fooDTS = New[Foo](FooDTM,
					muss.MarshallerFn[Foo](MarshalFooMUS),
					muss.UnmarshallerFn[Foo](UnmarshalFooMUS),
					muss.SizerFn[Foo](SizeFooMUS),
				)
				size = fooDTS.SizeMUS(foo)
				buf  = bytes.NewBuffer(make([]byte, 0, size))
			)
			n, err := fooDTS.MarshalMUS(foo, buf)
			if err != nil {
				t.Fatal(err)
			}
			if n != size {
				t.Fatalf("unexpected n, want '%v' actual '%v'", size, n)
			}
			afoo, n, err := fooDTS.UnmarshalMUS(buf)
			if err != nil {
				t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
			}
			if n != size {
				t.Errorf("unexpected n, want '%v' actual '%v'", size, n)
			}
			if !reflect.DeepEqual(afoo, foo) {
				t.Errorf("unexpected afoo, want '%v' actual '%v'", foo, afoo)
			}
		})

	t.Run("MarshalMUS, UnmarshalDTMUS, UnmarshalDataMUS, SizeMUS methods should work correctly",
		func(t *testing.T) {
			var (
				wantDTSize = 1
				foo        = Foo{num: 11, str: "hello world"}
				fooDTS     = New[Foo](FooDTM,
					muss.MarshallerFn[Foo](MarshalFooMUS),
					muss.UnmarshallerFn[Foo](UnmarshalFooMUS),
					muss.SizerFn[Foo](SizeFooMUS),
				)
				size = fooDTS.SizeMUS(foo)
				buf  = bytes.NewBuffer(make([]byte, 0, size))
			)
			n, err := fooDTS.MarshalMUS(foo, buf)
			if err != nil {
				t.Fatal(err)
			}
			if n != size {
				t.Fatalf("unexpected n, want '%v' actual '%v'", size, n)
			}
			dtm, n, err := UnmarshalDTMUS(buf)
			if err != nil {
				t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
			}
			if n != wantDTSize {
				t.Errorf("unexpected n, want '%v' actual '%v'", 1, n)
			}
			if dtm != FooDTM {
				t.Errorf("unexpected dtm, want '%v' actual '%v'", FooDTM, dtm)
			}
			afoo, n, err := fooDTS.UnmarshalDataMUS(buf)
			if err != nil {
				t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
			}
			if n != size-wantDTSize {
				t.Errorf("unexpected n, want '%v' actual '%v'", size, n)
			}
			if !reflect.DeepEqual(afoo, foo) {
				t.Errorf("unexpected afoo, want '%v' actual '%v'", foo, afoo)
			}
		})

	t.Run("DTM method should return correct DTM", func(t *testing.T) {
		var (
			fooDTS = New[Foo](FooDTM, nil, nil, nil)
			dtm    = fooDTS.DTM()
		)
		if dtm != FooDTM {
			t.Errorf("unexpected dtm, want '%v' actual '%v'", FooDTM, dtm)
		}
	})

	t.Run("UnamrshalMUS should fail with ErrWrongDTM, if meets another DTM",
		func(t *testing.T) {
			var (
				wantDTSize = 1
				r          = muss_mock.NewReader().RegisterReadByte(
					func() (b byte, err error) {
						b = byte(FooDTM) + 3
						return
					},
				)
				fooDTS      = New[Foo](FooDTM, nil, nil, nil)
				foo, n, err = fooDTS.UnmarshalMUS(r)
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

	t.Run("If MarshalDTMUS fails with an error, MarshalMUS should return it",
		func(t *testing.T) {
			var (
				wantErr = errors.New("write byte error")
				w       = muss_mock.NewWriter().RegisterWriteByte(func(c byte) error {
					return wantErr
				})
				fooDTS = New[Foo](FooDTM, nil, nil, nil)
			)
			_, err := fooDTS.MarshalMUS(Foo{}, w)
			if err != wantErr {
				t.Errorf("unexpected error, want '%v' actual '%v'", wantErr, err)
			}
		})

	t.Run("If UnmarshalDTMUS fails with an error, UnmarshalMUS should return it",
		func(t *testing.T) {
			var (
				wantErr = errors.New("read byte error")
				r       = muss_mock.NewReader().RegisterReadByte(
					func() (b byte, err error) {
						err = wantErr
						return
					},
				)
				fooDTS      = New[Foo](FooDTM, nil, nil, nil)
				foo, n, err = fooDTS.UnmarshalMUS(r)
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

	t.Run("If varint.UnmarshalInt fails with an error, UnmarshalDTMUS should return it",
		func(t *testing.T) {
			var (
				wantErr = errors.New("read byte error")
				r       = muss_mock.NewReader().RegisterReadByte(
					func() (b byte, err error) {
						err = wantErr
						return
					},
				)
				dtm, n, err = UnmarshalDTMUS(r)
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
