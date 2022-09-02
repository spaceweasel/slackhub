package markdown

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// item represents a token returned from the scanner.
type item struct {
	typ itemType // Type, such as itemNumber.
	pos int      // The starting position, in bytes, of this item in the input string.
	val string   // Value, such as "23.2".
}

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return i.val
	case len(i.val) > 10:
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

// itemType identifies the type of lex items.
type itemType int

const (
	itemError itemType = iota // error occurred; value is text of error
	itemEOF
	itemEOL // '\n' or '\r\n'
	itemText
	// itemBold   // slack only supports single asterisk
	// itemItalic // slack only supports single underscore
	itemHeader

	itemCodeStart
	itemCode
	itemCodeFinish
	itemCodeLang

	itemLinkTextStart
	itemLinkText
	itemLinkTextFinish
	itemLinkURLStart
	itemLinkURL
	itemLinkURLFinish

	itemBlockQuote
	itemBullet
	itemStar
	itemUnderscore

	itemEsc
)

const eof = -1

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	input string    // the string being scanned
	pos   int       // current position in the input
	start int       // start position of this item
	width int       // width of last rune read from input
	items chan item // channel of scanned items
}

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += l.width
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.start, l.input[l.start:l.pos]}
	l.start = l.pos
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, l.start, fmt.Sprintf(format, args...)}
	return nil
}

// nextItem returns the next item from the input.
// Called by the parser, not in the lexing goroutine.
func (l *lexer) nextItem() item {
	return <-l.items
}

// drain drains the output so the lexing goroutine will exit.
// Called by the parser, not in the lexing goroutine.
func (l *lexer) drain() {
	for range l.items {
	}
}

// lex creates a new scanner for the input string.
func lex(input string) *lexer {
	l := &lexer{
		input: input,
		items: make(chan item),
	}
	// TODO: check if we can avoid this 35:55
	go l.run()
	return l
}

// run runs the state machine for the lexer.
func (l *lexer) run() {
	for state := lexLine; state != nil; {
		state = state(l)
	}
	close(l.items)
}

// state functions

const (
	codeBlock = "```"
	bullet    = "- "
)

func lexLine(l *lexer) stateFn {
	if strings.HasPrefix(l.input[l.pos:], "#") {
		return lexHeader
	}
	if strings.HasPrefix(l.input[l.pos:], codeBlock) {
		return lexCodeStart
	}
	if strings.HasPrefix(l.input[l.pos:], ">") {
		return lexQuoteStart
	}
	if strings.HasPrefix(l.input[l.pos:], bullet) {
		return lexBullet
	}

	// Number List Item
	return lexText
}

func lexText(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == '\r' || r == '\n':
			l.backup()
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexEOL

		case r == '[':
			l.backup()
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexLinkTextStart

		case r == '*':
			l.backup()
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexStar

		case r == '_':
			l.backup()
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexUnderscore

		case r == '\\':
			l.backup()
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexEscape

		case r == eof:
			if l.pos > l.start {
				l.emit(itemText)
			}
			l.emit(itemEOF)
			return nil
		}
	}
}

func lexStar(l *lexer) stateFn {
	l.pos += len("*")
	l.emit(itemStar)
	return lexText
}

func lexUnderscore(l *lexer) stateFn {
	l.pos += len("_")
	l.emit(itemUnderscore)
	return lexText
}

func lexEscape(l *lexer) stateFn {
	l.pos += len("\\")
	l.emit(itemEsc)
	return lexText
}

func lexBullet(l *lexer) stateFn {
	l.pos += len(bullet)
	l.emit(itemBullet)
	return lexText
}

func lexQuoteStart(l *lexer) stateFn {
	l.pos += len(">")
	l.emit(itemBlockQuote)
	return lexText
}

func lexLinkTextStart(l *lexer) stateFn {
	l.pos += len("[")
	l.emit(itemLinkTextStart)
	return lexLinkText
}

func lexLinkText(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof:
			return l.errorf("no closing brace for link text")
		case r == ']':
			l.backup()
			l.emit(itemLinkText)
			return lexLinkTextFinish
		}
	}
}

func lexLinkTextFinish(l *lexer) stateFn {
	l.pos += len("]")
	l.emit(itemLinkTextFinish)
	return lexLinkURLStart
}

func lexLinkURLStart(l *lexer) stateFn {
	if l.accept("(") {
		l.emit(itemLinkURLStart)
		return lexLinkURL
	}

	return lexText
}

func lexLinkURL(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof:
			return l.errorf("no closing brace for link URL")
		case r == ')':
			l.backup()
			l.emit(itemLinkURL)
			return lexLinkURLFinish
		}
	}
}

func lexLinkURLFinish(l *lexer) stateFn {
	l.pos += len(")")
	l.emit(itemLinkURLFinish)
	return lexText
}

func lexEOL(l *lexer) stateFn {
	r := l.next()
	if r == '\r' {
		l.accept("\n")
	}

	l.emit(itemEOL)
	return lexLine
}

func lexHeader(l *lexer) stateFn {
	var h int
	for {
		if !l.accept("#") {
			break
		}
		h++
	}
	if h <= 6 && l.accept(" ") {
		l.emit(itemHeader)
	}
	return lexText
}

func lexCodeStart(l *lexer) stateFn {
	l.pos += len(codeBlock)
	l.emit(itemCodeStart)

loop:
	for {
		switch r := l.next(); {
		case r == eof:
			return l.errorf("unclosed code block")
		case r == '\r' || r == '\n':
			l.backup()
			l.emit(itemCodeLang)
			break loop
		}
	}

	r := l.next()
	if r == '\r' {
		l.accept("\n")
	}

	return lexCode
}

func lexCode(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == '\n':
			if strings.HasPrefix(l.input[l.pos:], codeBlock) {
				l.emit(itemCode)
				return lexCodeFinish
			}
		case r == eof:
			return l.errorf("unclosed code block")
		}
	}
}

func lexCodeFinish(l *lexer) stateFn {
	l.pos += len(codeBlock)
	l.emit(itemCodeFinish)
	return lexText
}
