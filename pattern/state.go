package pattern

import (
    "bytes"
)

func lexPath(lexer *lexer) stateFn {
    r := lexer.readRune()

    if r == eof {
        lexer.emit(EofToken())

        return nil
    }

    switch {
    case r == LEFT_CURLY_BRACKET:
        lexer.emit(LeftCurlyBracketToken())

        return lexCurlyParam(lexer)

    case r == COLON:
        lexer.emit(ColonToken())

        return lexColonParam(lexer)

    case r == PATH_SEPARATOR:
        lexer.emit(PathSeparatorToken())

        return lexPath(lexer)

    default:
        lexer.unreadRune()

        return lexPathPart(lexer)
    }
}

func lexCurlyParam(lexer *lexer) stateFn {
    ctxBuffer := bytes.NewBuffer(nil)

    for {
        r := lexer.readRune()

        switch {
        case r == RIGHT_CURLY_BRACKET:
            lexer.emit(ParamNameToken(ctxBuffer.Bytes()))
            lexer.emit(RightCurlyBracketToken())

            return lexPath(lexer)

        case r == PARAM_SEPARATOR:
            lexer.emit(ParamNameToken(ctxBuffer.Bytes()))
            lexer.emit(ParamSeparatorToken())

            return lexCurlyParamPattern(lexer)

        default:
            ctxBuffer.WriteRune(r)
        }
    }
}

func lexCurlyParamPattern(lexer *lexer) stateFn {
    ctxBuffer := bytes.NewBuffer(nil)

    for {
        r := lexer.readRune()

        if r != RIGHT_CURLY_BRACKET {
            ctxBuffer.WriteRune(r)
        } else {
            lexer.emit(ParamPatternToken(ctxBuffer.Bytes()))
            lexer.emit(RightCurlyBracketToken())

            return lexPath(lexer)
        }
    }
}

func lexColonParam(lexer *lexer) stateFn {
    ctxBuffer := bytes.NewBuffer(nil)

    for {
        r := lexer.readRune()

        if r != PATH_SEPARATOR && r != eof {
            ctxBuffer.WriteRune(r)
        } else {
            lexer.unreadRune()

            lexer.emit(ParamNameToken(ctxBuffer.Bytes()))

            return lexPath(lexer)
        }
    }
}

func lexPathPart(lexer *lexer) stateFn {
    ctxBuffer := bytes.NewBuffer(nil)

    for {
        r := lexer.readRune()

        if r == LEFT_CURLY_BRACKET || r == RIGHT_CURLY_BRACKET || r == COLON || r == PATH_SEPARATOR || r == eof {
            lexer.unreadRune()

            lexer.emit(PathPartToken(ctxBuffer.Bytes()))

            return lexPath(lexer)
        }

        ctxBuffer.WriteRune(r)
    }
}
