package meta

import "bytes"

type TagInfo struct {
	Tag         []byte
	Name        string
	Description string
}

var Tags []TagInfo

func FindTag(tag []byte) *TagInfo {
	size := len(tag)
	var candidates []*TagInfo

	for _, t := range Tags {
		(func(tag TagInfo) {
			if len(tag.Tag) == size {
				candidates = append(candidates, &tag)
			}
		})(t)
	}

	for _, t := range candidates {
		if bytes.Compare(t.Tag, tag) == 0 {
			return t
		}
	}

	return nil
}
