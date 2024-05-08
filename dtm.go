package dts

import (
	com "github.com/mus-format/common-go"
	muss "github.com/mus-format/mus-stream-go"
	"github.com/mus-format/mus-stream-go/varint"
)

func MarshalDTM(dtm com.DTM, w muss.Writer) (n int, err error) {
	return varint.MarshalInt(int(dtm), w)
}

func UnmarshalDTM(r muss.Reader) (dtm com.DTM, n int, err error) {
	num, n, err := varint.UnmarshalInt(r)
	if err != nil {
		return
	}
	dtm = com.DTM(num)
	return
}

func SizeDTM(dtm com.DTM) (size int) {
	return varint.SizeInt(int(dtm))
}
