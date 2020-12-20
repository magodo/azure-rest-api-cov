package addr

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// item represents a token or text string returned from the scanner.
type item struct {
	typ itemType // The type of this item.
	pos Pos      // The starting position, in bytes, of this item in the input string.
	val string   // The value of this item.
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
	itemPropertyDelim
	itemLeftDiscriminatorDelim
	itemRightDiscriminatorDelim
	itemPropertyName
	itemDiscriminatorValue
	itemEOF
)

const (
	PropertyDelim           rune = '/'
	leftDiscriminatorDelim       = '{'
	rightDiscriminatorDelim      = '}'
	escapeCharacter              = '\\'
)

const eof = -1

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// Pos represents a byte position in the original input text from which
// it was parsed.
type Pos int

func (p Pos) Position() Pos {
	return p
}

// lexer holds the state of the scanner.
type lexer struct {
	input string    // the string being scanned
	pos   Pos       // current position in the input
	start Pos       // start position of this item
	width Pos       // width of last rune read from input
	items chan item // channel of scanned items
}

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = Pos(w)
	l.pos += l.width
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
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

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
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

// run runs the state machine for the lexer.
func (l *lexer) run() {
	for state := lexPropertyName; state != nil; {
		state = state(l)
	}
	close(l.items)
}

// lex creates a new scanner for the input string.
func lex(input string) *lexer {
	l := &lexer{
		input: input,
		items: make(chan item),
	}
	go l.run()
	return l
}

// lexPropertyName scans until a left discriminator delim or a property deilim
func lexPropertyName(l *lexer) stateFn {
	for {
		next := l.peek()
		switch next {
		case escapeCharacter:
			l.next()
			l.next()
			continue
		case PropertyDelim:
			l.emit(itemPropertyName)
			return lexPropertyDelim
		case leftDiscriminatorDelim:
			l.emit(itemPropertyName)
			return lexLeftDiscriminatorDelim
		case eof:
			l.emit(itemPropertyName)
			l.emit(itemEOF)
			return nil
		default:
			l.next()
		}
	}
}

func lexPropertyDelim(l *lexer) stateFn {
	l.pos += Pos(len(string(PropertyDelim)))
	l.emit(itemPropertyDelim)
	l.ignore()
	return lexPropertyName
}

func lexLeftDiscriminatorDelim(l *lexer) stateFn {
	l.pos += Pos(len(string(leftDiscriminatorDelim)))
	l.emit(itemLeftDiscriminatorDelim)
	l.ignore()
	return lexDiscriminatorValue
}

// lexDiscriminatorValue scans until a right discriminator delim
func lexDiscriminatorValue(l *lexer) stateFn {
	for {
		next := l.peek()
		switch next {
		case escapeCharacter:
			l.next()
			l.next()
			continue
		case rightDiscriminatorDelim:
			l.emit(itemDiscriminatorValue)
			return lexRightDiscriminatorDelim
		case eof:
			return l.errorf("no matching %q until EOF ", rightDiscriminatorDelim)
		default:
			l.next()
		}
	}

}

func lexRightDiscriminatorDelim(l *lexer) stateFn {
	l.pos += Pos(len(string(rightDiscriminatorDelim)))
	l.emit(itemRightDiscriminatorDelim)
	l.ignore()
	next := l.peek()
	switch next {
	case PropertyDelim:
		return lexPropertyDelim
	case eof:
		l.emit(itemEOF)
		return nil
	default:
		return l.errorf("unexpected character %q after right discriminator delim", string(next))
	}
}
