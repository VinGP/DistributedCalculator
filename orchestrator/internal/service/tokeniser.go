package service

import (
	"errors"
	"regexp"
)

type TokenType int

const (
	Number TokenType = iota
	Plus
	Minus
	Multiply
	Divide
	Power
	LeftBracket
	RightBracket
)

type Token struct {
	Type  TokenType
	Value string
}

type Tokenizer struct {
}

func NewTokenizer() *Tokenizer {
	return &Tokenizer{}
}

func (t *Tokenizer) Tokenize(expression string) ([]Token, error) {
	var tokens []Token

	re := regexp.MustCompile(`[+\-*^/()]|[0-9.]+`)

	matches := re.FindAllString(expression, -1)

	for _, match := range matches {
		token := Token{
			Value: match,
		}

		switch match {
		case "+":
			token.Type = Plus
		case "-":
			if len(tokens) != 0 && tokens[len(tokens)-1].Type == LeftBracket {
				tokens = append(tokens, Token{
					Type:  Number,
					Value: "0"})

			}
			token.Type = Minus
		case "*":
			token.Type = Multiply
		case "/":
			token.Type = Divide
		case "^":
			token.Type = Power
		case "(":
			token.Type = LeftBracket
		case ")":
			token.Type = RightBracket
		default:
			token.Type = Number
			if len(tokens) == 1 && tokens[0].Type == Minus {
				token.Value = "-" + token.Value
				tokens = tokens[1:]

			}

			if len(tokens) >= 2 && tokens[len(tokens)-1].Type == Minus && tokens[len(tokens)-2].Type == LeftBracket {
				tokens[len(tokens)-1].Value = "-" + token.Value
				tokens[len(tokens)-1].Type = Number

				continue
			}

			if len(tokens) >= 2 && tokens[len(tokens)-1].Type == Minus && (tokens[len(tokens)-2].Type != Number && tokens[len(tokens)-2].Type != RightBracket && tokens[len(tokens)-2].Type != LeftBracket) {
				return []Token{}, errors.New("invalid expression")
			}
		}
		tokens = append(tokens, token)
	}

	return tokens, nil
}
