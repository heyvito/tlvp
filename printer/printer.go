package printer

import (
	"fmt"
	"github.com/gookit/color"
	"io"
	"strings"

	"github.com/victorgama/tlvp/meta"
	"github.com/victorgama/tlvp/tlv"
)

type itemLine struct {
	text  string
	color color.Color
}

type printItem struct {
	data     []itemLine
	title    itemLine
	parent   *printItem
	children []*printItem
}

func (item *printItem) addLine(line string, color color.Color) {
	item.data = append(item.data, itemLine{line, color})
}

func (item *printItem) setTitle(line string) {
	item.title = itemLine{line, color.Normal}
}

func (item *printItem) isLastChild() bool {
	if item.parent == nil {
		return true
	}
	return item.parent.children[len(item.parent.children)-1] == item
}

func Print(entries []*tlv.GenericEntry, describe bool) {
	var result []*printItem
	for _, e := range entries {
		func(entry *tlv.GenericEntry) {
			processEntry(entry, describe, nil, &result)
		}(e)
	}
	printItems(result, describe)
}

func printItems(entries []*printItem, describe bool) {
	for _, e := range entries {
		printEntry(e, describe)
	}
}

func autoWrap(text string, wrapAt int) []string {
	var result []string
	var elements = strings.Split(text, " ")
	var currentLine []string

	for {
		if len(elements) == 0 {
			if len(currentLine) > 0 {
				result = append(result, strings.Join(currentLine, " "))
			}
			return result
		}

		// Check whether we can add a new item to the current line
		newArr := append(currentLine, elements[0])
		if len(strings.Join(newArr, " ")) > wrapAt {
			// Too long. Flush current line and start a new one.
			result = append(result, strings.Join(currentLine, " "))
			currentLine = []string{elements[0]}
		} else {
			currentLine = newArr
		}
		elements = elements[1:]
	}
}

func printEntry(e *printItem, describe bool) {
	margin := printMargin(e, true)
	fmt.Printf("%s%s\n", margin, e.title.text)

	margin = printMargin(e, false)
	marginLen := len(margin)

	for _, s := range e.data {
		for _, txt := range autoWrap(s.text, 80-marginLen) {
			fmt.Printf("%s%s\n", margin, s.color.Sprint(txt))
		}
	}

	if describe {
		fmt.Printf("%s\n", margin)
	}

	printItems(e.children, describe)
}

func printMargin(entry *printItem, isTitle bool) string {
	var tree []*printItem
	e := entry
	for {
		if e == nil {
			break
		}
		tree = append(tree, e)
		e = e.parent
	}
	tree = reverseTree(tree)
	return printTree(tree, isTitle, "")
}

func printTree(tree []*printItem, isTitle bool, margin string) string {
	if len(tree) == 0 {
		return margin
	}

	item := tree[0]

	if isTitle && item.parent != nil {
		if len(tree) == 1 {
			margin += "  "
		} else if len(tree) > 1 {
			margin += "      "
		}

		if item.parent != nil && item.isLastChild() {
			if len(tree) == 1 {
				margin += "└"
			}
		} else if item.parent != nil && !item.isLastChild() {
			margin += "├"
		}
	}

	if !isTitle {
		if len(tree) == 1 {
			if item.isLastChild() && item.parent != nil {
				margin += "      "
			}
			if !item.isLastChild() || item.parent == nil || !item.parent.isLastChild() || len(item.children) > 0 {
				margin += "  │   "
			}
		} else if item.parent != nil {
			if item.isLastChild() {
				margin += "      "
			} else {
				margin += "    "
			}
		}
	}

	return printTree(tree[1:], isTitle, margin)
}

func reverseTree(input []*printItem) []*printItem {
	if len(input) == 0 {
		return input
	}
	return append(reverseTree(input[1:]), input[0])
}

func processEntry(entry *tlv.GenericEntry, describe bool, parent *printItem, list *[]*printItem) {
	item := printItem{
		parent: parent,
	}
	item.setTitle(printTitle(entry, &item))
	if describe {
		if metadata := entry.AsMeta(); metadata != nil {
			if info := metadata.TagInfo; info != nil {
				item.addLine(info.Description, color.Gray)
			}
		}
	}
	item.addLine(printSize(entry))

	if entry.Kind == tlv.KindConstructed {
		constructed := entry.AsConstructed()
		for {
			e, _, err := constructed.Next()
			if err == io.EOF {
				break
			}

			processEntry(e, describe, &item, &item.children)
		}
	} else {
		item.addLine(printValue(entry.AsPrimitive().Value))
	}

	newList := append(*list, &item)
	*list = newList
}

func printSize(ent *tlv.GenericEntry) (string, color.Color) {
	size := ent.AsMeta().ValueLength
	plural := ""
	if size > 1 {
		plural = "s"
	}
	return fmt.Sprintf("Size: %02d byte%s", size, plural), color.Normal
}

func printTitle(ent *tlv.GenericEntry, item *printItem) string {
	var (
		bytes string
		title string
	)
	if ent.Kind == tlv.KindConstructed {
		e := ent.AsConstructed()
		bytes = printBytes(e.TagBytes, false)
		if tagMeta := e.TagInfo; tagMeta != nil {
			title = tagMeta.Name
		} else {
			title = "Unknown"
		}
	} else {
		e := ent.AsPrimitive()
		bytes = printBytes(e.TagBytes, false)
		if tagMeta := e.TagInfo; tagMeta != nil {
			title = tagMeta.Name
		} else {
			title = "Unknown"
		}
	}

	rail := ""
	if item.parent != nil {
		rail = "── "
	}

	return fmt.Sprintf("%s[%s] %s", rail, color.Magenta.Text(bytes), color.Light.Sprint(title))
}

func printBytes(buf []byte, pretty bool) string {
	var format string
	separator := ""
	if pretty {
		format = "0x%02X"
		separator = " "
	} else {
		format = "%02X"
	}

	data := make([]string, len(buf))
	for i, b := range buf {
		data[i] = fmt.Sprintf(format, b)
	}

	return strings.Join(data, separator)
}

func printValue(buf []byte) (string, color.Color) {
	isPrintable := meta.IsAsciiPrintable(buf)
	var result string
	if isPrintable {
		result = string(buf)
	} else {
		result = printBytes(buf, true)
	}

	return result, color.Cyan
}
