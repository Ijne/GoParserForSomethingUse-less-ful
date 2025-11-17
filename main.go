package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

const (
	TOKEN_EOF = iota
	TOKEN_IDENT
	TOKEN_NUMBER
	TOKEN_STRING
	TOKEN_LPAREN
	TOKEN_RPAREN
	TOKEN_LBRACKET
	TOKEN_RBRACKET
	TOKEN_SEMICOLON
	TOKEN_ASSIGN
	TOKEN_HASH
	TOKEN_BANG
	TOKEN_DEFINE
	TOKEN_BEGIN
	TOKEN_END
	TOKEN_PLUS
	TOKEN_STAR
)

type Token struct {
	Type  int
	Value string
	Line  int
	Pos   int
}

type Parser struct {
	tokens []Token
	pos    int
	defs   map[string]interface{}
}

type Value interface{}

func lex(input string) ([]Token, error) {
	var tokens []Token
	lines := strings.Split(input, "\n")

	for lineNum, line := range lines {
		lineNum++
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}
		if strings.HasPrefix(line, "=begin") {
			for i := lineNum; i < len(lines); i++ {
				if strings.TrimSpace(lines[i-1]) == "=cut" {
					break
				}
			}
			continue
		}

		pos := 0
		for pos < len(line) {
			if unicode.IsSpace(rune(line[pos])) {
				pos++
				continue
			}

			if strings.HasPrefix(line[pos:], "--") {
				break
			}

			if line[pos] == '\'' {
				end := strings.Index(line[pos+1:], "'")
				if end == -1 {
					return nil, fmt.Errorf("unclosed string at line %d", lineNum)
				}
				tokens = append(tokens, Token{TOKEN_STRING, line[pos+1 : pos+1+end], lineNum, pos})
				pos += end + 2
				continue
			}

			if match, _ := regexp.MatchString(`^[+-]?\d+\.?\d*[eE][+-]?\d+`, line[pos:]); match {
				num := ""
				start := pos
				for pos < len(line) && (unicode.IsDigit(rune(line[pos])) || line[pos] == '.' ||
					line[pos] == 'e' || line[pos] == 'E' || line[pos] == '+' || line[pos] == '-') {
					num += string(line[pos])
					pos++
				}
				tokens = append(tokens, Token{TOKEN_NUMBER, num, lineNum, start})
				continue
			}

			if unicode.IsDigit(rune(line[pos])) ||
				((line[pos] == '-' || line[pos] == '+') && pos+1 < len(line) && unicode.IsDigit(rune(line[pos+1]))) {
				num := ""
				start := pos
				if line[pos] == '-' || line[pos] == '+' {
					num += string(line[pos])
					pos++
				}
				for pos < len(line) && unicode.IsDigit(rune(line[pos])) {
					num += string(line[pos])
					pos++
				}
				if pos < len(line) && line[pos] == '.' {
					num += string(line[pos])
					pos++
					for pos < len(line) && unicode.IsDigit(rune(line[pos])) {
						num += string(line[pos])
						pos++
					}
				}
				tokens = append(tokens, Token{TOKEN_NUMBER, num, lineNum, start})
				continue
			}

			if unicode.IsLetter(rune(line[pos])) || line[pos] == '_' {
				ident := ""
				start := pos
				for pos < len(line) && (unicode.IsLetter(rune(line[pos])) || unicode.IsDigit(rune(line[pos])) || line[pos] == '_') {
					ident += string(line[pos])
					pos++
				}

				switch ident {
				case "define":
					tokens = append(tokens, Token{TOKEN_DEFINE, ident, lineNum, start})
				case "begin":
					tokens = append(tokens, Token{TOKEN_BEGIN, ident, lineNum, start})
				case "end":
					tokens = append(tokens, Token{TOKEN_END, ident, lineNum, start})
				default:
					tokens = append(tokens, Token{TOKEN_IDENT, ident, lineNum, start})
				}
				continue
			}

			switch line[pos] {
			case '(':
				tokens = append(tokens, Token{TOKEN_LPAREN, "(", lineNum, pos})
			case ')':
				tokens = append(tokens, Token{TOKEN_RPAREN, ")", lineNum, pos})
			case '[':
				tokens = append(tokens, Token{TOKEN_LBRACKET, "[", lineNum, pos})
			case ']':
				tokens = append(tokens, Token{TOKEN_RBRACKET, "]", lineNum, pos})
			case ';':
				tokens = append(tokens, Token{TOKEN_SEMICOLON, ";", lineNum, pos})
			case ':':
				if pos+1 < len(line) && line[pos+1] == '=' {
					tokens = append(tokens, Token{TOKEN_ASSIGN, ":=", lineNum, pos})
					pos++
				} else {
					return nil, fmt.Errorf("unexpected character ':' at line %d, position %d", lineNum, pos)
				}
			case '#':
				tokens = append(tokens, Token{TOKEN_HASH, "#", lineNum, pos})
			case '!':
				tokens = append(tokens, Token{TOKEN_BANG, "!", lineNum, pos})
			case '+':
				tokens = append(tokens, Token{TOKEN_PLUS, "+", lineNum, pos})
			case '*':
				tokens = append(tokens, Token{TOKEN_STAR, "*", lineNum, pos})
			default:
				return nil, fmt.Errorf("unexpected character '%c' at line %d, position %d", line[pos], lineNum, pos)
			}
			pos++
		}
	}

	tokens = append(tokens, Token{TOKEN_EOF, "", len(lines) + 1, 0})
	return tokens, nil
}

func (p *Parser) current() Token {
	if p.pos >= len(p.tokens) {
		return Token{TOKEN_EOF, "", 0, 0}
	}
	return p.tokens[p.pos]
}

func (p *Parser) advance() {
	p.pos++
}

func (p *Parser) expect(tokenType int) error {
	if p.current().Type != tokenType {
		return fmt.Errorf("expected token %d, got %d at line %d", tokenType, p.current().Type, p.current().Line)
	}
	p.advance()
	return nil
}

func (p *Parser) parseValue() (Value, error) {
	token := p.current()

	switch token.Type {
	case TOKEN_NUMBER:
		p.advance()
		if strings.Contains(token.Value, ".") || strings.Contains(token.Value, "e") || strings.Contains(token.Value, "E") {
			val, err := strconv.ParseFloat(token.Value, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid number format: %s at line %d", token.Value, token.Line)
			}
			return val, nil
		}
		val, err := strconv.Atoi(token.Value)
		if err != nil {
			return nil, fmt.Errorf("invalid number format: %s at line %d", token.Value, token.Line)
		}
		return val, nil

	case TOKEN_STRING:
		p.advance()
		return token.Value, nil

	case TOKEN_HASH:
		return p.parseArray()

	case TOKEN_BEGIN:
		return p.parseDictionary()

	case TOKEN_BANG:
		return p.parseConstantExpression()

	case TOKEN_IDENT:
		if val, exists := p.defs[token.Value]; exists {
			p.advance()
			return val, nil
		}
		p.advance()
		return token.Value, nil

	default:
		return nil, fmt.Errorf("unexpected token in value: %v at line %d", token, token.Line)
	}
}

func (p *Parser) parseArray() (Value, error) {
	if err := p.expect(TOKEN_HASH); err != nil {
		return nil, err
	}
	if err := p.expect(TOKEN_LPAREN); err != nil {
		return nil, err
	}

	var array []Value
	for p.current().Type != TOKEN_RPAREN && p.current().Type != TOKEN_EOF {
		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		array = append(array, val)
	}

	if err := p.expect(TOKEN_RPAREN); err != nil {
		return nil, err
	}

	return array, nil
}

func (p *Parser) parseDictionary() (Value, error) {
	if err := p.expect(TOKEN_BEGIN); err != nil {
		return nil, err
	}

	dict := make(map[string]Value)

	for p.current().Type != TOKEN_END && p.current().Type != TOKEN_EOF {
		if p.current().Type != TOKEN_IDENT {
			return nil, fmt.Errorf("expected identifier in dictionary, got %v at line %d", p.current(), p.current().Line)
		}

		key := p.current().Value
		p.advance()

		if err := p.expect(TOKEN_ASSIGN); err != nil {
			return nil, err
		}

		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}

		dict[key] = val

		if err := p.expect(TOKEN_SEMICOLON); err != nil {
			return nil, err
		}
	}

	if err := p.expect(TOKEN_END); err != nil {
		return nil, err
	}

	return dict, nil
}

func (p *Parser) parseConstantExpression() (Value, error) {
	if err := p.expect(TOKEN_BANG); err != nil {
		return nil, err
	}
	if err := p.expect(TOKEN_LBRACKET); err != nil {
		return nil, err
	}

	left, err := p.parseValue()
	if err != nil {
		return nil, err
	}

	for p.current().Type != TOKEN_RBRACKET && p.current().Type != TOKEN_EOF {
		var operator string
		switch p.current().Type {
		case TOKEN_PLUS:
			operator = "+"
			p.advance()
		case TOKEN_STAR:
			operator = "*"
			p.advance()
		case TOKEN_IDENT:
			if p.current().Value == "+" {
				operator = "+"
				p.advance()
			} else if p.current().Value == "*" {
				operator = "*"
				p.advance()
			} else {
				return nil, fmt.Errorf("unexpected identifier in expression: %s at line %d", p.current().Value, p.current().Line)
			}
		default:
			return nil, fmt.Errorf("unexpected token in expression: %v at line %d", p.current(), p.current().Line)
		}

		right, err := p.parseValue()
		if err != nil {
			return nil, err
		}

		switch operator {
		case "+":
			if leftNum, rightNum, ok := asInts(left, right); ok {
				left = leftNum + rightNum
			} else if leftNum, rightNum, ok := asFloats(left, right); ok {
				left = leftNum + rightNum
			} else {
				return nil, fmt.Errorf("invalid operands for + operation at line %d", p.current().Line)
			}
		case "*":
			if leftNum, rightNum, ok := asInts(left, right); ok {
				left = leftNum * rightNum
			} else if leftNum, rightNum, ok := asFloats(left, right); ok {
				left = leftNum * rightNum
			} else {
				return nil, fmt.Errorf("invalid operands for * operation at line %d", p.current().Line)
			}
		}
	}

	if err := p.expect(TOKEN_RBRACKET); err != nil {
		return nil, err
	}

	return left, nil
}

func (p *Parser) parseDefine() error {
	if err := p.expect(TOKEN_LPAREN); err != nil {
		return err
	}
	if err := p.expect(TOKEN_DEFINE); err != nil {
		return err
	}

	if p.current().Type != TOKEN_IDENT {
		return fmt.Errorf("expected identifier after define at line %d", p.current().Line)
	}

	name := p.current().Value
	p.advance()

	val, err := p.parseValue()
	if err != nil {
		return err
	}

	p.defs[name] = val

	if err := p.expect(TOKEN_RPAREN); err != nil {
		return err
	}
	if err := p.expect(TOKEN_SEMICOLON); err != nil {
		return err
	}

	return nil
}

func (p *Parser) parse() (map[string]Value, error) {
	result := make(map[string]Value)
	p.defs = make(map[string]interface{})

	for p.current().Type != TOKEN_EOF {
		switch p.current().Type {
		case TOKEN_LPAREN:
			if err := p.parseDefine(); err != nil {
				return nil, err
			}
		case TOKEN_IDENT:
			key := p.current().Value
			p.advance()

			if err := p.expect(TOKEN_ASSIGN); err != nil {
				return nil, err
			}

			val, err := p.parseValue()
			if err != nil {
				return nil, err
			}

			result[key] = val

			if err := p.expect(TOKEN_SEMICOLON); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unexpected token at top level: %v at line %d", p.current(), p.current().Line)
		}
	}

	return result, nil
}

func asInts(a, b Value) (int, int, bool) {
	numA, okA := a.(int)
	numB, okB := b.(int)
	return numA, numB, okA && okB
}

func asFloats(a, b Value) (float64, float64, bool) {
	var numA, numB float64
	var okA, okB bool

	if numA, okA = a.(float64); !okA {
		if i, ok := a.(int); ok {
			numA = float64(i)
			okA = true
		}
	}

	if numB, okB = b.(float64); !okB {
		if i, ok := b.(int); ok {
			numB = float64(i)
			okB = true
		}
	}

	return numA, numB, okA && okB
}

func toTOML(data map[string]Value, indent string) string {
	var result strings.Builder

	for key, value := range data {
		result.WriteString(key)
		result.WriteString(" = ")
		result.WriteString(valueToTOML(value))
		result.WriteString("\n")
	}

	return result.String()
}

func valueToTOML(value Value) string {
	switch v := value.(type) {
	case int:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%.1f", v)
	case string:
		return fmt.Sprintf("\"%s\"", v)
	case []Value:
		var elements []string
		for _, elem := range v {
			elements = append(elements, valueToTOML(elem))
		}
		return "[" + strings.Join(elements, ", ") + "]"
	case map[string]Value:
		var result strings.Builder
		result.WriteString("{ ")
		first := true
		for k, val := range v {
			if !first {
				result.WriteString(", ")
			}
			result.WriteString(k)
			result.WriteString(" = ")
			result.WriteString(valueToTOML(val))
			first = false
		}
		result.WriteString(" }")
		return result.String()
	default:
		return "null"
	}
}

func main() {
	inputFile := flag.String("input", "", "Path to input file")
	flag.Parse()

	if *inputFile == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s -input <file>\n", os.Args[0])
		os.Exit(1)
	}

	content, err := os.ReadFile(*inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	tokens, err := lex(string(content))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Lexer error: %v\n", err)
		os.Exit(1)
	}

	parser := &Parser{tokens: tokens}
	result, err := parser.parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parser error: %v\n", err)
		os.Exit(1)
	}

	tomlOutput := toTOML(result, "")
	fmt.Print(tomlOutput)
}
