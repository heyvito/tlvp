package tlv

import (
	"fmt"
	"io"

	"github.com/victorgama/tlvp/meta"
)

type EntryKind byte

const (
	KindConstructed EntryKind = 1 << iota
	KindPrimitive
	ClassUniversal
	ClassApplication
	ClassContextSpecific
	ClassPrivate
	HasExtendedTag
	IsPDOL
)

type ParserError struct {
	err      string
	position uint64
}

func (e ParserError) Error() string {
	return fmt.Sprintf("%s at position %d", e.err, e.position)
}

type index struct {
	kind  EntryKind
	index uint64
}

type GenericEntry struct {
	Kind  EntryKind
	Value interface{}
}

func (g *GenericEntry) AsPrimitive() PrimitiveEntry {
	return g.Value.(PrimitiveEntry)
}

func (g GenericEntry) AsConstructed() ConstructedEntry {
	return g.Value.(ConstructedEntry)
}

func (g GenericEntry) AsMeta() *MetaEntry {
	if g.Kind == KindConstructed {
		c := g.AsConstructed()
		return &c.MetaEntry
	} else {
		c := g.AsPrimitive()
		return &c.MetaEntry
	}
}

type MetaEntry struct {
	TagBytes    []byte
	TagInfo     *meta.TagInfo
	Kind        EntryKind
	ValueLength uint64

	tmpValue       []byte
}

type PrimitiveEntry struct {
	MetaEntry
	Value       []byte
	ValueLength uint64
}

type ConstructedEntry struct {
	MetaEntry
	Primitives  []PrimitiveEntry
	Constructed []ConstructedEntry

	index []index
	cur   int
}

func (c *ConstructedEntry) ResetCursor() {
	c.cur = 0
}

func (c *ConstructedEntry) Next() (*GenericEntry, bool, error) {
	if c.cur > len(c.index)-1 {
		return nil, false, io.EOF
	}

	item := c.index[c.cur]
	c.cur++
	last := c.cur > len(c.index)-1


	if item.kind == KindPrimitive {
		return &GenericEntry{KindPrimitive, c.Primitives[item.index]}, last, nil
	} else {
		return &GenericEntry{KindConstructed, c.Constructed[item.index]}, last, nil
	}
}
