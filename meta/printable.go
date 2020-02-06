package meta

import (
	"bufio"
	"unicode"
)

func isPrintable(b byte) bool {
	return b <= unicode.MaxASCII && unicode.IsPrint(rune(b))
}

func IsAsciiPrintable(bytes []byte) bool {
	for _, b := range bytes {
		if !isPrintable(b) && b != 0x0a {
			return false
		}
	}
	return true
}

func IsReaderPrintable(reader *bufio.Reader) bool {
	buf := make([]byte, 1)
	for {
		_, err := reader.Read(buf)
		if err != nil {
			return true
		}
		if buf[0] == 0x0a {
			continue
		}
		if !isPrintable(buf[0]) {
			return false
		}
	}
}
