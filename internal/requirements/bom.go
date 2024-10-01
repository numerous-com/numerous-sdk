package requirements

import (
	"bytes"
	"reflect"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
)

type bom []byte

func (b bom) Detect(v []byte) bool {
	return bytes.HasPrefix(v, b)
}

func (b bom) Equal(o bom) bool {
	return reflect.DeepEqual(b, o)
}

var (
	bomUTF8    = bom{0xef, 0xbb, 0xbf}
	bomUTF16LE = bom{0xff, 0xfe}
	bomUTF16BE = bom{0xfe, 0xff}
	bomUTF32LE = bom{0xff, 0xfe, 0x00, 0x00}
	bomUTF32BE = bom{0x00, 0x00, 0xfe, 0xff}
)

var boms = []bom{
	bomUTF8,
	bomUTF16LE,
	bomUTF16BE,
	bomUTF32LE,
	bomUTF32BE,
}

func (b bom) Encoding() encoding.Encoding {
	switch {
	case b.Equal(bomUTF8):
		return unicode.UTF8BOM
	case b.Equal(bomUTF16LE):
		return unicode.UTF16(unicode.LittleEndian, unicode.UseBOM)
	case b.Equal(bomUTF16BE):
		return unicode.UTF16(unicode.BigEndian, unicode.UseBOM)
	default:
		return nil
	}
}
