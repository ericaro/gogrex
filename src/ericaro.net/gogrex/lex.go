//package gogrex builds a regular graph based on any regular expression.
package gogrex

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

/*

STAR   : "*" >
 PLUS   : "+" >
 OPT    : "?" >
 BANG   : "&" >
 SEL    : "|" >
 SEQ    : "," >
 LEFT   : "(" >
 RIGHT  : ")" >
#LETTER : [ "_", "a"-"z", "A"-"Z"] >
#DIGIT  : [ "0"-"9"] >
 IDENTIFIER : < LETTER > (< LETTER > | < DIGIT > )* >

*/


const (
	typeError      = iota
	typeEOF        = iota
	typeStar       = iota
	typePlus       = iota
	typeOpt        = iota
	typeSel        = iota
	typeSeq        = iota
	typeLeft       = iota
	typeRight      = iota
	typeIdentifier = iota 
	typeComment    = iota
)

var(
	itemError      itemType = itemType{nature: typeError     ,operator:false , precedence: -1} // error occured
	itemEOF        itemType = itemType{nature: typeEOF       ,operator:false , precedence: -1} // error occured
	itemStar       itemType = itemType{nature: typeStar      ,operator:true  , precedence: 20} //   "*" 
	itemPlus       itemType = itemType{nature: typePlus      ,operator:true  , precedence: 20} //   "+"
	itemOpt        itemType = itemType{nature: typeOpt       ,operator:true  , precedence: 20} //   "?"
	itemSel        itemType = itemType{nature: typeSel       ,operator:true  , precedence: 10} //   "|"
	itemSeq        itemType = itemType{nature: typeSeq       ,operator:true  , precedence:  0} //   ","
	itemLeft       itemType = itemType{nature: typeLeft      ,operator:false , precedence: -1} //   "("
	itemRight      itemType = itemType{nature: typeRight     ,operator:false , precedence: -1} //   ")"
	itemIdentifier itemType = itemType{nature: typeIdentifier,operator:false , precedence: -1} //   any valid identifier 
	itemComment    itemType = itemType{nature: typeComment   ,operator:false , precedence: -1} //   any valid identifier 

)
const eof = -1

// itemType identifies the type of lex items.
type itemType struct{
	nature, precedence int
	operator bool 
	
}

// item represents a token returned from the scanner.
type item struct {
	typ itemType // Type, such as itemNumber.
	val string   // Value, such as "23.2".
}

func (i item) IsOperator()bool { return i.typ.operator}
func (i item) IsLeaf()bool { return i.typ.nature == typeIdentifier}
func (i item) IsLeftParenthesis()bool { return i.typ.nature == typeLeft}
func (i item) IsRightParenthesis()bool { return i.typ.nature == typeRight}
func (i item) IsLeftAssociative()bool { return true} // does not apply here
func (i item) Precedence()int { return i.typ.precedence} // does not apply here


//func (i item) isFunction()bool { return false} // no function operators here

type lexer struct {
	input string    // the string being scanned.
	start int       // start position of this item.
	pos   int       // current position in the input.
	width int       // width of last rune read from input.
	items chan Token // channel of scanned items.
}

// stateFn represents the state of the scanner
// as a function that returns the next state.
type stateFn func(*lexer) stateFn

//String prints any item
func (i item) String() string {
	switch i.typ {
	case itemStar:
		return "*"
	case itemPlus:
		return "+"
	case itemOpt:
		return "?"
	case itemSel:
		return "|"
	case itemSeq:
		return ","
	case itemLeft:
		return "("
	case itemRight:
		return ")"
	case itemIdentifier:
		return i.val
	case itemError:
		return i.val
	default:
		return i.val
	}
	return ""
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

// lex creates a new scanner for the input string.
func lex(input string) chan Token {
	l := &lexer{
		input: input,
		items: make(chan Token, 2), // Two items sufficient.
	}
	go l.run()
	return l.items
}
func (l *lexer) run() {
    for state := lexText; state != nil; {
        state = state(l)
    }
    close(l.items) // No more tokens will be delivered.
}
// next returns the next rune in the input.
func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}
//
//// nextItem returns the next item from the input.
//func (l *lexer) nextItem() item {
//	for {
//		select {
//		case item := <-l.items:
//			return item
//		default:
//			l.state = l.state(l)
//		}
//	}
//	panic("not reached")
//}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// backup steps back one rune.
// Can be called only once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// peek returns but does not consume
// the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// accept consumes the next rune
// if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

// error returns an error token and terminates the scan
// by passing back a nil pointer that will be the next
// state, terminating l.run.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{
		itemError,
		fmt.Sprintf(format, args...),
	}
	return nil
}

func lexText(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof:
			l.emit(itemEOF)
			return nil
		case r == '*':
			l.emit(itemStar)
		case r == '+':
			l.emit(itemPlus)
		case r == '?':
			l.emit(itemOpt)
		case r == '|':
			l.emit(itemSel)
		case r == ',':
			l.emit(itemSeq)
		case r == '(':
			l.emit(itemLeft)
		case r == ')':
			l.emit(itemRight)
		case unicode.IsSpace(r): // auto ignored
			l.ignore()
		case unicode.IsLetter(r):
			return lexIdentifier // now read an identifier
		case r == '/': // comment start
			
			switch r = l.next() ; {
			case r== '/':
			 return lexSingleLineComment 
			case r== '*':
			 return lexMultiLineComment 
			default:
				l.errorf("invalid comment start /%s", r)
			} 		
		default:
			l.errorf("Unknown character %s", r)
		}
	}
	return nil // Stop the run loop.
}


func lexIdentifier(l *lexer) stateFn {
	for r:= l.next() ; unicode.IsLetter(r) || unicode.IsDigit(r); r = l.next(){
	}
	l.backup()
	l.emit(itemIdentifier)
	return lexText
}
	
func lexSingleLineComment(l *lexer) stateFn {
	for r:= l.next() ; r != '\n'; r = l.next(){
	}
	l.emit(itemComment)
	return lexText
}
func lexMultiLineComment(l *lexer) stateFn {
	p:= l.next()
	for r:= l.next() ; p!='*' && r != '/'; r = l.next(){
	}
	
	l.emit(itemComment)
	return lexText
}

