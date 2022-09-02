package markdown

import (
	"fmt"
	"log"
	"runtime"
	"strings"
)

// Parse parses the github markdown to construct a slack markdown representation.
func Parse(text string) (smd string, err error) {
	p := &parser{
		lex: lex(text),
	}

	defer p.recover(&err)
	p.parse()

	return strings.ReplaceAll(p.text, "\t", "\\t"), nil
}

type parser struct {
	lex       *lexer
	token     [2]item // two-token lookahead for parser.
	peekCount int
	text      string
}

// next returns the next token.
func (p *parser) next() item {
	if p.peekCount > 0 {
		p.peekCount--
	} else {
		p.token[0] = p.lex.nextItem()
	}
	log.Println(p.token[p.peekCount])
	return p.token[p.peekCount]
}

// peek returns but does not consume the next token.
func (p *parser) peek() item {
	if p.peekCount > 0 {
		return p.token[p.peekCount-1]
	}
	p.peekCount = 1
	p.token[0] = p.lex.nextItem()
	return p.token[0]
}

// backup backs the input stream up one token.
func (p *parser) backup() {
	p.peekCount++
}

// parse is the top-level parser for the github markdown.
// It runs to EOF.
func (p *parser) parse() {
	//for p.peek()
	for {
		switch n := p.next(); n.typ {
		case itemEOF:
			return
		case itemEOL:
			p.text += "\\n"
		case itemEsc:
			// TODO: There has to be a way to deal with escaping in Slack
			p.text += n.val
			return
		case itemText:
			p.text += n.val
		case itemHeader:
			p.parseHeader()

		case itemLinkTextStart:
			p.backup()
			p.parseLink()

		case itemStar:
			n = p.next()
			// double star is bold
			if n.typ != itemStar {
				p.text += "_"
				p.backup()
			} else {
				p.text += "*"
			}
		case itemUnderscore:
			n = p.next()
			// double underscore is bold
			if n.typ != itemUnderscore {
				p.text += "_"
				p.backup()
			} else {
				p.text += "*"
			}
		case itemBullet:
			p.text += "• "
		case itemCodeStart:
			p.text += n.val
		case itemCodeFinish:
			p.text += n.val
		case itemCodeLang:
			// ignore
		case itemCode:
			code := strings.ReplaceAll(n.val, "\n", "\\n")
			code = strings.ReplaceAll(code, "\r", "")
			p.text += code

		default:
			p.errorf("unexpected %s", n)
		}
	}

}

func (p *parser) parseHeader() {
	p.text += "*"
	for {
		switch n := p.next(); n.typ {
		case itemEOF:
			p.text += "*"
			p.backup()
			return
		case itemEOL:
			p.text += "*"
			p.text += "\\n"
			return
		case itemEsc:
			next := p.peek()
			_ = next
		case itemText:
			p.text += n.val
		case itemStar:
			n = p.next()
			// double star is bold - ignore in heading
			if n.typ != itemStar {
				// italic
				p.text += "_"
				p.backup()
			}
		case itemUnderscore:
			n = p.next()
			// double underscore is bold - ignore in heading
			if n.typ != itemUnderscore {
				// italic
				p.text += "_"
				p.backup()
			}

		case itemLinkTextStart:
			p.backup()
			p.parseLink()
		default:
			p.errorf("unexpected %s", n)
		}
	}

}

func (p *parser) parseLink() {
	var text, url string
	p.next() // consume opening [
	n := p.next()
	if n.typ != itemLinkText {
		p.backup()
		p.text += "["
		return
	}
	text = n.val
	p.next() // consume closing ]

	n = p.next()
	if n.typ != itemLinkURLStart {
		p.backup()
		p.text += "[" + text + "]"
		return
	}

	n = p.next()
	if n.typ != itemLinkURL {
		p.backup()
		p.text += "[" + text + "]("
		return
	}

	url = n.val
	n = p.next() // consume closing )

	p.text += fmt.Sprintf("<%s|%s>", url, text)
}

func (p *parser) parseX() {

}

// errorf formats the error and terminates processing.
func (p *parser) errorf(format string, args ...interface{}) {
	format = fmt.Sprintf("parse: %s", format)
	panic(fmt.Errorf(format, args...))
}

// expect consumes the next token and guarantees it has the required type.
func (p *parser) expect(expected ...itemType) item {
	token := p.next()
	for _, e := range expected {
		if token.typ == e {
			return token
		}
	}

	p.unexpected(token)
	return item{}
}

// unexpected complains about the token and terminates processing.
func (p *parser) unexpected(token item) {
	p.errorf("unexpected %s", token)
}

// recover is the handler that turns panics into returns from the top level of Parse.
func (p *parser) recover(errp *error) {
	e := recover()
	if e != nil {
		// propagate panics not created by us
		if _, ok := e.(runtime.Error); ok {
			panic(e)
		}
		if p != nil {
			p.lex.drain()
		}
		*errp = e.(error)
	}
}
