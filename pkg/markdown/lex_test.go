package markdown

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

const (
	CR   = "\r"
	LF   = "\n"
	CRLF = CR + LF
)

func TestLex(t *testing.T) {
	c := qt.New(t)

	lexTest := func(input string, expected []item, checkPos bool) func(c *qt.C) {
		return func(c *qt.C) {
			var items []item
			l := lex(input)
			for {
				item := l.nextItem()
				items = append(items, item)
				if item.typ == itemEOF || item.typ == itemError {
					break
				}
			}
			c.Log("got:\n", items)
			c.Log("want:\n", expected)
			for i, item := range items {
				c.Assert(item.typ, qt.Equals, expected[i].typ)
				c.Assert(item.val, qt.Equals, expected[i].val)
				if checkPos {
					c.Assert(item.pos, qt.Equals, expected[i].pos)
				}
			}
		}
	}

	c.Run("Single Line of plain text",
		lexTest("A line of plain text", []item{
			newItem(itemText, "A line of plain text"),
			testEOF,
		}, false))

	c.Run("Multiple lines of plain text (LF)",
		lexTest("A line of text\nAnother line of text", []item{
			newItem(itemText, "A line of text"),
			newItem(itemEOL, LF),
			newItem(itemText, "Another line of text"),
			testEOF,
		}, false))

	c.Run("Multiple lines of plain text (CRLF)",
		lexTest("A line of text\r\nAnother line of text", []item{
			newItem(itemText, "A line of text"),
			newItem(itemEOL, CRLF),
			newItem(itemText, "Another line of text"),
			testEOF,
		}, false))

	c.Run("Initial Header (LF)",
		lexTest("# A Header\nLine of text", []item{
			newItem(itemHeader, "# "),
			newItem(itemText, "A Header"),
			newItem(itemEOL, LF),
			newItem(itemText, "Line of text"),
			testEOF,
		}, false))

	c.Run("Initial Header (CRLF)",
		lexTest("# A Header\r\nLine of text", []item{
			newItem(itemHeader, "# "),
			newItem(itemText, "A Header"),
			newItem(itemEOL, CRLF),
			newItem(itemText, "Line of text"),
			testEOF,
		}, false))

	c.Run("Multiple Headers (LF)",
		lexTest("# Header\nLine of text\n## Sub Header\nMore text...", []item{
			newItem(itemHeader, "# "),
			newItem(itemText, "Header"),
			newItem(itemEOL, LF),
			newItem(itemText, "Line of text"),
			newItem(itemEOL, LF),
			newItem(itemHeader, "## "),
			newItem(itemText, "Sub Header"),
			newItem(itemEOL, LF),
			newItem(itemText, "More text..."),
			testEOF,
		}, false))

	c.Run("Multiple Headers (CRLF)",
		lexTest("# Header\r\nLine of text\r\n## Sub Header\r\nMore text...", []item{
			newItem(itemHeader, "# "),
			newItem(itemText, "Header"),
			newItem(itemEOL, CRLF),
			newItem(itemText, "Line of text"),
			newItem(itemEOL, CRLF),
			newItem(itemHeader, "## "),
			newItem(itemText, "Sub Header"),
			newItem(itemEOL, CRLF),
			newItem(itemText, "More text..."),
			testEOF,
		}, false))

	c.Run("Invalid Header, no space (LF)",
		lexTest("#A Header\nLine of text", []item{
			newItem(itemText, "#A Header"),
			newItem(itemEOL, LF),
			newItem(itemText, "Line of text"),
			testEOF,
		}, false))

	c.Run("Invalid Header, no space (CRLF)",
		lexTest("#A Header\r\nLine of text", []item{
			newItem(itemText, "#A Header"),
			newItem(itemEOL, CRLF),
			newItem(itemText, "Line of text"),
			testEOF,
		}, false))

	c.Run("Invalid Header, too many # (LF)",
		lexTest("####### A Header\nLine of text", []item{
			newItem(itemText, "####### A Header"),
			newItem(itemEOL, LF),
			newItem(itemText, "Line of text"),
			testEOF,
		}, false))

	c.Run("Invalid Header, too many # (CRLF)",
		lexTest("####### A Header\r\nLine of text", []item{
			newItem(itemText, "####### A Header"),
			newItem(itemEOL, CRLF),
			newItem(itemText, "Line of text"),
			testEOF,
		}, false))

	c.Run("Inline code",
		lexTest("A code `term` in line of text", []item{
			newItem(itemText, "A code `term` in line of text"),
			testEOF,
		}, false))

	c.Run("Standalone Code Block (LF)",
		lexTest("```\n{\n\tx = y\n}\n```", []item{
			newItem(itemCodeStart, "```"),
			newItem(itemCodeLang, ""),
			newItem(itemCode, "\n{\n\tx = y\n}\n"),
			newItem(itemCodeFinish, "```"),
			testEOF,
		}, false))

	c.Run("Standalone Code Block (CRLF)",
		lexTest("```\r\n{\r\n\tx = y\r\n}\r\n```", []item{
			newItem(itemCodeStart, "```"),
			newItem(itemCodeLang, ""),
			newItem(itemCode, "\r\n{\r\n\tx = y\r\n}\r\n"),
			newItem(itemCodeFinish, "```"),
			testEOF,
		}, false))

	c.Run("Code Block with Language (LF)",
		lexTest("```banana\n{\n\tx = y\n}\n```", []item{
			newItem(itemCodeStart, "```"),
			newItem(itemCodeLang, "banana"),
			newItem(itemCode, "\n{\n\tx = y\n}\n"),
			newItem(itemCodeFinish, "```"),
			testEOF,
		}, false))

	c.Run("Code Block with Language (CRLF)",
		lexTest("```banana\r\n{\r\n\tx = y\r\n}\r\n```", []item{
			newItem(itemCodeStart, "```"),
			newItem(itemCodeLang, "banana"),
			newItem(itemCode, "\r\n{\r\n\tx = y\r\n}\r\n"),
			newItem(itemCodeFinish, "```"),
			testEOF,
		}, false))

	c.Run("Code Block Surrounded by Textblocks (LF)",
		lexTest("Suggestion:\n```\n{\n\tx = y\n}\n```\nShould fix this", []item{
			newItem(itemText, "Suggestion:"),
			newItem(itemEOL, LF),
			newItem(itemCodeStart, "```"),
			newItem(itemCodeLang, ""),
			newItem(itemCode, "\n{\n\tx = y\n}\n"),
			newItem(itemCodeFinish, "```"),
			newItem(itemEOL, LF),
			newItem(itemText, "Should fix this"),
			testEOF,
		}, false))

	c.Run("Code Block Surrounded by Textblocks (CRLF)",
		lexTest("Suggestion:\r\n```\r\n{\r\n\tx = y\r\n}\r\n```\r\nShould fix this", []item{
			newItem(itemText, "Suggestion:"),
			newItem(itemEOL, CRLF),
			newItem(itemCodeStart, "```"),
			newItem(itemCodeLang, ""),
			newItem(itemCode, "\r\n{\r\n\tx = y\r\n}\r\n"),
			newItem(itemCodeFinish, "```"),
			newItem(itemEOL, CRLF),
			newItem(itemText, "Should fix this"),
			testEOF,
		}, false))

	c.Run("Simple URL",
		lexTest("A line with http://somewhere link", []item{
			newItem(itemText, "A line with http://somewhere link"),
			testEOF,
		}, false))

	c.Run("URL with link text",
		lexTest("A line with [click me](http://somewhere) link", []item{
			newItem(itemText, "A line with "),
			newItem(itemLinkTextStart, "["),
			newItem(itemLinkText, "click me"),
			newItem(itemLinkTextFinish, "]"),
			newItem(itemLinkURLStart, "("),
			newItem(itemLinkURL, "http://somewhere"),
			newItem(itemLinkURLFinish, ")"),
			newItem(itemText, " link"),
			testEOF,
		}, false))

	c.Run("BlockQuote",
		lexTest(">A quotes line", []item{
			newItem(itemBlockQuote, ">"),
			newItem(itemText, "A quotes line"),
			testEOF,
		}, false))

	c.Run("BlockQuote With Trailing Spaces",
		lexTest(">   A quotes line", []item{
			newItem(itemBlockQuote, ">"),
			newItem(itemText, "   A quotes line"),
			testEOF,
		}, false))

	c.Run("Escaped BlockQuote",
		lexTest(`\>A quotes line`, []item{
			newItem(itemEsc, `\`),
			newItem(itemText, ">A quotes line"),
			testEOF,
		}, false))

	c.Run("Escaped URL with link text",
		lexTest(`A line with \[click me](http://somewhere) link`, []item{
			newItem(itemText, "A line with "),
			newItem(itemEsc, `\`),
			newItem(itemLinkTextStart, "["),
			newItem(itemLinkText, "click me"),
			newItem(itemLinkTextFinish, "]"),
			newItem(itemLinkURLStart, "("),
			newItem(itemLinkURL, "http://somewhere"),
			newItem(itemLinkURLFinish, ")"),
			newItem(itemText, " link"),
			testEOF,
		}, false))

	c.Run("Single Bullet",
		lexTest("- A bullet line", []item{
			newItem(itemBullet, "- "),
			newItem(itemText, "A bullet line"),
			testEOF,
		}, false))

	c.Run("Leading Hyphen with no spaces",
		lexTest("-A non bullet line", []item{
			newItem(itemText, "-A non bullet line"),
			testEOF,
		}, false))

	c.Run("Escaped Bullet",
		lexTest(`\- An escaped bullet line`, []item{
			newItem(itemEsc, `\`),
			newItem(itemText, "- An escaped bullet line"),
			testEOF,
		}, false))

	c.Run("Double Escaped Sequence",
		lexTest(`This is a path: \\server\\ping for biscuits`, []item{
			newItem(itemText, "This is a path: "),
			newItem(itemEsc, `\`),
			newItem(itemEsc, `\`),
			newItem(itemText, "server"),
			newItem(itemEsc, `\`),
			newItem(itemEsc, `\`),
			newItem(itemText, "ping for biscuits"),
			testEOF,
		}, false))

	c.Run("Emphasis and strong with * and _",
		lexTest(`This _is_ the *best* day for **eating** cake and __drinking__ beer!`, []item{
			newItem(itemText, "This "),
			newItem(itemUnderscore, `_`),
			newItem(itemText, "is"),
			newItem(itemUnderscore, `_`),
			newItem(itemText, " the "),
			newItem(itemStar, `*`),
			newItem(itemText, "best"),
			newItem(itemStar, `*`),
			newItem(itemText, " day for "),
			newItem(itemStar, `*`),
			newItem(itemStar, `*`),
			newItem(itemText, "eating"),
			newItem(itemStar, `*`),
			newItem(itemStar, `*`),
			newItem(itemText, " cake and "),
			newItem(itemUnderscore, `_`),
			newItem(itemUnderscore, `_`),
			newItem(itemText, "drinking"),
			newItem(itemUnderscore, `_`),
			newItem(itemUnderscore, `_`),
			newItem(itemText, " beer!"),
			testEOF,
		}, false))
}

func newItem(typ itemType, val string) item {
	return item{
		typ: typ,
		val: val,
	}
}

func newItemWithPos(typ itemType, pos int, val string) item {
	return item{
		typ: typ,
		pos: pos,
		val: val,
	}
}

var testEOF = newItem(itemEOF, "")
