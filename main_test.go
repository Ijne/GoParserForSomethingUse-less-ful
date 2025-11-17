package main

import (
	"testing"
)

func TestLexer(t *testing.T) {
	input := `server_name := 'test'; port := 8080;`
	tokens, err := lex(input)
	if err != nil {
		t.Fatalf("Lexer error: %v", err)
	}

	if len(tokens) < 7 {
		t.Errorf("Expected 7 tokens, got %d", len(tokens))
	}
}

func TestParser(t *testing.T) {
	input := `name := 'test'; value := 42;`
	tokens, _ := lex(input)
	parser := &Parser{tokens: tokens}
	result, err := parser.parse()

	if err != nil {
		t.Fatalf("Parser error: %v", err)
	}

	if result["name"] != "test" {
		t.Errorf("Expected name='test', got %v", result["name"])
	}
}

func TestArray(t *testing.T) {
	input := `items := #( 1 2 3 );`
	tokens, _ := lex(input)
	parser := &Parser{tokens: tokens}
	result, err := parser.parse()

	if err != nil {
		t.Fatalf("Array parsing error: %v", err)
	}

	arr, ok := result["items"].([]Value)
	if !ok || len(arr) != 3 {
		t.Errorf("Expected 3 elements, got %v", result["items"])
	}
}

func TestDictionary(t *testing.T) {
	input := `config := begin key := 'value'; num := 42; end;`
	tokens, _ := lex(input)
	parser := &Parser{tokens: tokens}
	result, err := parser.parse()

	if err != nil {
		t.Fatalf("Dict parsing error: %v", err)
	}

	dict, ok := result["config"].(map[string]Value)
	if !ok || dict["key"] != "value" {
		t.Errorf("Expected key='value', got %v", result["config"])
	}
}
