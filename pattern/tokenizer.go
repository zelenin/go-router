package pattern

import (
    "bufio"
    "io"
)

const eof = 0

func Lex(reader io.Reader) Tokenizer {
    lexer := &lexer{
        reader: bufio.NewReader(reader),
        tokens: make(chan Token, 10),
    }

    go lexer.run()

    return lexer
}

type Tokenizer interface {
    NextToken() Token
}

type stateFn func(*lexer) stateFn

type lexer struct {
    reader *bufio.Reader
    state  stateFn
    tokens chan Token
}

func (lexer *lexer) run() {
    for lexer.state = lexPath; lexer.state != nil; {
        lexer.state = lexer.state(lexer)
    }
    close(lexer.tokens)
}

func (lexer *lexer) readRune() rune {
    r, _, err := lexer.reader.ReadRune()
    if err != nil {
        return eof
    }

    return r
}

func (lexer *lexer) nextRune() rune {
    r, _, err := lexer.reader.ReadRune()
    if err != nil {
        return eof
    }

    lexer.unreadRune()
    return r
}

func (lexer *lexer) unreadRune() error {
    return lexer.reader.UnreadRune()
}

func (lexer *lexer) NextToken() Token {
    return <-lexer.tokens
}

func (lexer *lexer) emit(token Token) {
    lexer.tokens <- token
}
