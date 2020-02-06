package tlv

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/victorgama/tlvp/mem"
	"github.com/victorgama/tlvp/meta"
	"io"
	"io/ioutil"
	"regexp"
)

type state uint8

const (
	stateTagFirstByte state = iota
	stateTagSecondByte
	stateLength
	stateExtendedLength
	stateValue
)

var (
	ErrInvalidClass      = ParserError { "invalid class", 0 }
	ErrInvalidObjectKind = ParserError { "invalid object kind", 0 }
	ErrUnexpectedEOF     = ParserError { "unexpected EOF", 0 }
	ErrInconsistentSize = ParserError{ "inconsistent size %d", 0 }
)

type Parser struct {
	reader        io.Reader
	state         state
	readBytes     uint64
	entries       []MetaEntry
	currentEntry  *MetaEntry
	isPDOL        bool
	tmpBytesTotal uint64
	tmpBytesRead  uint64
}

var binaryStripper = regexp.MustCompile("(0x|[^0-9A-Fa-f])")

func NewParser(reader io.Reader, isPDOL bool) (*Parser, error) {
	// Detect input encoding
	rawBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	convertToBin := false
	if meta.IsAsciiPrintable(rawBytes) {
		convertToBin = true
	}

	if convertToBin {
		// Here we need to convert input data into its binary representation...

		rawBytes = binaryStripper.ReplaceAll(rawBytes, []byte{})
		if len(rawBytes) % 2 != 0 {
			return nil, fmt.Errorf("input must have a power of 2 length")
		}

		rawBytes, err = hex.DecodeString(string(rawBytes))
		if err != nil {
			return nil, err
		}

	}

	reader = bytes.NewReader(rawBytes)


	return &Parser{
		reader:       reader,
		state:        stateTagFirstByte,
		entries:      []MetaEntry{},
		currentEntry: nil,
		isPDOL:       isPDOL,
	}, nil
}

func newSubParser(reader io.Reader, isPDOL bool, readBytes uint64) (*Parser, error) {
	p, err := NewParser(reader, isPDOL)
	if err != nil {
		return nil, err
	}
	p.readBytes = readBytes
	return p, nil
}

func (p *Parser) Parse() ([]*GenericEntry, error) {
	buf := make([]byte, 1)
	var parserError error

readLoop:
	for {
		n, err := p.reader.Read(buf)
		if n == 1 {
			parserError = p.feed(buf[0])
		}

		switch {
		case parserError != nil:
			return nil, parserError
		case err == io.EOF:
			break readLoop
		case err != nil:
			return nil, err
		}
	}

	if p.currentEntry != nil {
		p.entries = append(p.entries, *p.currentEntry)
	}

	if p.state != stateTagFirstByte {
		return nil, p.error(ErrUnexpectedEOF)
	}

	return p.normalize()
}

func (p *Parser) error(err error) error {
	if err, ok := err.(ParserError); ok {
		err.position = p.readBytes
		return err
	}
	return err
}

func (p *Parser) feed(b byte) (returnErr error) {
	defer func() {
		if r := recover(); r != nil {
			returnErr = p.error(r.(error))
		}
	}()

	switch p.state {
	case stateTagFirstByte:
		if b == 0x00 {
			// Null data. We can safely skip this, as this may be followed
			// by more null items. In this case, we will send all those
			// into a black hole.
			break
		}
		p.currentEntry = &MetaEntry{}
		if p.isPDOL {
			p.currentEntry.Kind = IsPDOL
		} else {
			p.currentEntry.Kind = 0
		}

		// Class
		class, err := parseClass((b | 0xc0) >> 6)
		if err != nil {
			return p.error(err)
		}
		p.currentEntry.Kind |= class

		// Object Kind
		objKind, err := parseObjectKind((b & 0x20) >> 5)
		if err != nil {
			return p.error(err)
		}
		p.currentEntry.Kind |= objKind

		p.currentEntry.TagBytes = []byte{b}
		if b&0x1F == 0x1F {
			// Extended tag.
			p.currentEntry.Kind |= HasExtendedTag
			p.state = stateTagSecondByte
		} else {
			p.currentEntry.TagBytes = []byte{b}
			p.state = stateLength
		}

	case stateTagSecondByte:
		p.currentEntry.TagBytes = append(p.currentEntry.TagBytes, b)
		if b&0x80 == 0x00 {
			// When 8th bit is set, we have another tag byte following.
			// Otherwise, we're done reading it and can just move on.
			p.state = stateLength
		}

	case stateLength:
		// Lookup tag information before moving on...
		p.currentEntry.TagInfo = meta.FindTag(p.currentEntry.TagBytes)

		if b&0x80 == 0x00 {
			// When first bit is empty, means we don't have more than a byte
			// depicting the tag length.
			p.tmpBytesTotal = uint64(b)
			p.currentEntry.ValueLength = p.tmpBytesTotal
			p.goToValueOrRestart()
			break
		}
		// First byte indicates how many bytes compose the tag.
		p.tmpBytesRead = 0
		p.tmpBytesTotal = uint64(b & 0x7F)
		p.state = stateExtendedLength
	case stateExtendedLength:
		p.currentEntry.ValueLength = (p.currentEntry.ValueLength << uint64(8)) | uint64(b)
		p.tmpBytesRead++

		if p.tmpBytesRead == p.tmpBytesTotal {
			p.goToValueOrRestart()
		}
	case stateValue:
		p.currentEntry.tmpValue[p.tmpBytesRead] = b
		p.tmpBytesRead++

		if p.tmpBytesRead == p.currentEntry.ValueLength {
			p.entries = append(p.entries, *p.currentEntry)
			p.currentEntry = nil
			p.state = stateTagFirstByte
		}
	}

	p.readBytes++
	return nil
}

func (p *Parser) goToValueOrRestart() {
	if !mem.IsValidSize(p.currentEntry.ValueLength) {
		err := ErrInconsistentSize
		err.err = fmt.Sprintf(err.err, p.currentEntry.ValueLength)
		panic(err)
	}

	p.currentEntry.tmpValue = make([]byte, p.currentEntry.ValueLength)
	if p.isPDOL {
		p.state = stateTagFirstByte
		return
	}

	p.tmpBytesRead = 0
	if p.currentEntry.ValueLength == 0 {
		p.state = stateTagFirstByte
	} else {
		p.state = stateValue
	}
}

func parseClass(class byte) (tag EntryKind, err error) {
	switch class {
	case 0x00:
		tag = ClassUniversal
	case 0x01:
		tag = ClassApplication
	case 0x02:
		tag = ClassContextSpecific
	case 0x03:
		tag = ClassPrivate
	default:
		return 0, ErrInvalidClass
	}
	err = nil
	return
}

func parseObjectKind(kind byte) (tag EntryKind, err error) {
	switch kind {
	case 0x00:
		tag = KindPrimitive
	case 0x01:
		tag = KindConstructed
	default:
		return 0, ErrInvalidObjectKind
	}
	err = nil
	return
}

func (p *Parser) normalize() ([]*GenericEntry, error) {
	var result []*GenericEntry
	for _, e := range p.entries {
		switch {
		case e.Kind & KindConstructed == KindConstructed:
			entry := ConstructedEntry{
				MetaEntry: e,
				Primitives: []PrimitiveEntry{},
				Constructed: []ConstructedEntry{},
				index: []index{},
			}
			parser, err := newSubParser(bytes.NewReader(e.tmpValue), p.isPDOL, p.readBytes)
			if err != nil {
				return nil, err
			}
			entries, err := parser.Parse()
			if err != nil {
				return nil, err
			}
			for _, je := range entries {
				if je.Kind == KindPrimitive {
					entry.Primitives = append(entry.Primitives, je.Value.(PrimitiveEntry))
					entry.index = append(entry.index, index{je.Kind, uint64(len(entry.Primitives)) - 1})
				} else {
					entry.Constructed = append(entry.Constructed, je.Value.(ConstructedEntry))
					entry.index = append(entry.index, index{je.Kind, uint64(len(entry.Constructed)) - 1})
				}
			}
			result = append(result, &GenericEntry{KindConstructed, entry})
		case e.Kind & KindPrimitive == KindPrimitive:
			result = append(result, &GenericEntry{KindPrimitive, PrimitiveEntry{
				MetaEntry:   e,
				Value:       e.tmpValue,
				ValueLength: e.ValueLength,
			}})
		default:
			panic(fmt.Errorf("found unknown class type during normalisation phase"))
		}
	}

	return result, nil
}
