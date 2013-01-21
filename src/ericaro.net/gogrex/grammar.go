package gogrex

import (
	"errors"
	"fmt"
)

//grammar for the regexp
// we are gonna use the shunting yard algorithm


type Token interface {
	Precedence() int
	IsOperator() bool
	IsLeaf() bool
	IsLeftParenthesis() bool
	IsRightParenthesis() bool
	IsLeftAssociative() bool
	//	Error() error
}

//type errorToken string
//func (err * errorToken) Precedence() int         {return 0}
//func (err * errorToken) IsOperator() bool        {return false}
//func (err * errorToken) IsLeaf() bool            {return false}
//func (err * errorToken) IsLeftParenthesis() bool {return false}
//func (err * errorToken) IsRightParenthesis() bool{return false}
//func (err * errorToken) IsLeftAssociative() bool {return false}
//func (err * errorToken) Error() error            {return errors.New(string(*err))}

type itemStack []Token

func (stack *itemStack) Pop() (i Token, err error) {
	i, err = stack.Peek()
	if err == nil {
		*stack = (*stack)[:len(*stack)-1]
	}
	return
}

func (stack *itemStack) Push(x Token) {
	*stack = append(*stack, x)
}

func (stack *itemStack) Peek() (i Token, err error) {
	if len(*stack) == 0 {
		return i, errors.New("Empty Stack")
	}
	return (*stack)[len(*stack)-1], nil
}

//shunting Yard
func shunting(tokens chan Token) (output chan Token, err chan error) {
	output = make(chan Token)
	err = make(chan error)
	go run(tokens, output, err)
	return
}


//run execute a partial the shunting yard algorithm ( http://en.wikipedia.org/wiki/Shunting-yard_algorithm ) (no function support)
func run(tokens chan Token, output chan Token, errchan chan error) {
	stack := itemStack(make([]Token, 0, 10))
	for token := range tokens {
		switch {
		case token.IsLeaf(): // usually a number in shunting yard, or an identifier
			output <- token
		//case token.IsFunction(): stack.Push(token) // ignored for now, I don't need to support function call
		//If the token is an operator, o1, then:
		case token.IsOperator():
			o2, err := stack.Peek()
			for err == nil && (( // while there is an operator token, o2,at the top of the stack, and
			//o1 is left-associative and its precedence is less than or equal to that of o2,
			token.IsLeftAssociative() && token.Precedence() <= o2.Precedence()) || (
			//o1 has precedence less than that of o2,
			token.Precedence() < o2.Precedence())) {
				stack.Pop()
				output <- o2
				o2, err = stack.Peek()

				fmt.Printf("        err %v , o2 %v\n", err, o2)
			}
			stack.Push(token)
		//If the token is a left parenthesis, then push it onto the stack.
		case token.IsLeftParenthesis():
			//fmt.Printf("is left\n")
			stack.Push(token)
		//If the token is a right parenthesis:
		case token.IsRightParenthesis():
			o2, err := stack.Peek()
			for err == nil && !o2.IsLeftParenthesis() { //Until the token at the top of the stack is a left parenthesis, 
				//pop operators off the stack onto the output queue.
				stack.Pop()
				output <- o2
				o2, err = stack.Peek()
			}
			if !o2.IsLeftParenthesis() {
				errchan <- errors.New("parenthesis mismatch")
			}
			stack.Pop()

			//		case token.typ.nature == typeEOF:
			//			close(output)
		}
	}
	for len(stack) > 0 {
		pop, err := stack.Pop()
		if err != nil || pop.IsLeftParenthesis() || pop.IsRightParenthesis() {
			errchan <- errors.New("parenthesis mismatch")
		} // this is an error
		output <- pop
	}
	close(output)
	close(errchan)
	return
}


