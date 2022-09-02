package markdown

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestParse(t *testing.T) {
	c := qt.New(t)

	parseTest := func(input string, expected string, errStr string) func(c *qt.C) {
		return func(c *qt.C) {
			smd, err := Parse(input)
			c.Log(smd, err)
			if errStr == "" {
				c.Assert(err, qt.IsNil)
				c.Assert(smd, qt.DeepEquals, expected)
			} else {
				c.Assert(err, qt.ErrorMatches, errStr)
			}
		}
	}

	c.Run("Single line of text",
		parseTest(`A single line of text`,
			`A single line of text`, ""))

	c.Run("Multiple lines of text",
		parseTest(`A single line of text
followed by another.`,
			`A single line of text\nfollowed by another.`, ""))

	c.Run("Multiple lines of text (CRLF)",
		parseTest("A single line of text\r\nfollowed by another.",
			`A single line of text\nfollowed by another.`, ""))

	c.Run("Initial Header",
		parseTest("# A Header\nLine of text",
			`*A Header*\nLine of text`, ""))

	c.Run("Initial Header with underscore italic",
		parseTest("# An _Italic_ Header\nLine of text",
			`*An _Italic_ Header*\nLine of text`, ""))

	c.Run("Initial Header with star italic",
		parseTest("# An *Italic* Header\nLine of text",
			`*An _Italic_ Header*\nLine of text`, ""))

	c.Run("Initial Header with underscore bold",
		parseTest("# A __Bold__ Header\nLine of text",
			`*A Bold Header*\nLine of text`, ""))

	c.Run("Initial Header with star bold",
		parseTest("# A **Bold** Header\nLine of text",
			`*A Bold Header*\nLine of text`, ""))

	c.Run("Simple link",
		parseTest("Please [click me](http://here.com) for cake",
			`Please <http://here.com|click me> for cake`, ""))

	c.Run("Missing link",
		parseTest("Please [click me] for cake",
			`Please [click me] for cake`, ""))

	c.Run("Simple line with underscore italic",
		parseTest("An _Italic_ Line\nLine of text",
			`An _Italic_ Line\nLine of text`, ""))

	c.Run("Simple line with star italic",
		parseTest("An *Italic* Line\nLine of text",
			`An _Italic_ Line\nLine of text`, ""))

	c.Run("Simple line with underscore bold",
		parseTest("A __Bold__ Line\nLine of text",
			`A *Bold* Line\nLine of text`, ""))

	c.Run("Simple line with star bold",
		parseTest("A **Bold** Line\nLine of text",
			`A *Bold* Line\nLine of text`, ""))

	c.Run("Multi line with underscore italic",
		parseTest("An _Italic Line\nLine_ of text",
			`An _Italic Line\nLine_ of text`, ""))

	c.Run("Multi line with star italic",
		parseTest("An *Italic Line\nLine* of text",
			`An _Italic Line\nLine_ of text`, ""))

	c.Run("Multi line with underscore bold",
		parseTest("A __Bold Line\nLine__ of text",
			`A *Bold Line\nLine* of text`, ""))

	c.Run("Multi line with star bold",
		parseTest("A **Bold Line\nLine** of text",
			`A *Bold Line\nLine* of text`, ""))

	c.Run("Underscore Italic Nested within Bold Stars",
		parseTest("A **bold _Italic_** Line\nLine of text",
			`A *bold _Italic_* Line\nLine of text`, ""))

	c.Run("Star Italic Nested within Bold Underscores",
		parseTest("A __bold *Italic*__ Line\nLine of text",
			`A *bold _Italic_* Line\nLine of text`, ""))

	c.Run("Bullet points",
		parseTest("A Line before\n- Option A\n- Option B\nOut of options",
			`A Line before\n• Option A\n• Option B\nOut of options`, ""))

	c.Run("Numbered list",
		parseTest("A Line before\n1. Option A\n2. Option B\nOut of options",
			`A Line before\n1. Option A\n2. Option B\nOut of options`, ""))

	c.Run("Inline Code",
		parseTest("A `code variable` in a line",
			"A `code variable` in a line", ""))

	c.Run("Basic code block",
		parseTest("```\ntype Eater interface{\n  Eat()\n}\n```",
			"```\\ntype Eater interface{\\n  Eat()\\n}\\n```", ""))

	c.Run("Text surrounded code block",
		parseTest("Use this:\n```go\ntype Eater interface{\n  Eat()\n}\n```\ninstead.",
			"Use this:\\n```\\ntype Eater interface{\\n  Eat()\\n}\\n```\\ninstead.", ""))

	c.Run("Code block containing markdown controls",
		parseTest("```\n# A Heading\nThis **is not** bold\n```",
			"```\\n# A Heading\\nThis **is not** bold\\n```", ""))

	c.Run("Tabs replaced",
		parseTest("Use\n\tthis:\n```go\ntype Eater interface{\n\tEat()\n}\n```\ninstead.",
			"Use\\n\\tthis:\\n```\\ntype Eater interface{\\n\\tEat()\\n}\\n```\\ninstead.", ""))

	c.Run("With icons",
		parseTest("## 💬 What does this PR do and why is this needed?\r\nThis PR eats all the custard\r\n\r\n## 📝 Describe the important code changes\r\n\r\n## ❔ Questions or remarks",
			"**💬 What does this PR do and why is this needed?**\\nThis PR eats all the custard\\n\\n**📝 Describe the important code changes**\\n\\n**❔ Questions or remarks**", ""))

	//
}
