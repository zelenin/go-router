package pattern

type tokenType int

const (
    TOKEN_UNKNOWN tokenType = (iota - 1)
    TOKEN_EOF
    TOKEN_ERROR

    TOKEN_PATH_PART
    TOKEN_PATH_SEPARATOR

    TOKEN_LEFT_CURLY_BRACKET
    TOKEN_RIGHT_CURLY_BRACKET

    TOKEN_COLON

    TOKEN_PARAM_NAME
    TOKEN_PARAM_PATTERN
    TOKEN_PARAM_SEPARATOR
)

type Token struct {
    Type  tokenType
    Value []byte
}

func UnknownToken() Token {
    return Token{TOKEN_UNKNOWN, nil}
}

func EofToken() Token {
    return Token{TOKEN_EOF, nil}
}

func ErrorToken(value []byte) Token {
    return Token{TOKEN_ERROR, value}
}

func PathPartToken(value []byte) Token {
    return Token{TOKEN_PATH_PART, value}
}

func PathSeparatorToken() Token {
    return Token{TOKEN_PATH_SEPARATOR, []byte(string(PATH_SEPARATOR))}
}

func LeftCurlyBracketToken() Token {
    return Token{TOKEN_LEFT_CURLY_BRACKET, []byte(string(LEFT_CURLY_BRACKET))}
}

func RightCurlyBracketToken() Token {
    return Token{TOKEN_RIGHT_CURLY_BRACKET, []byte(string(RIGHT_CURLY_BRACKET))}
}

func ColonToken() Token {
    return Token{TOKEN_COLON, []byte(string(COLON))}
}

func ParamNameToken(value []byte) Token {
    return Token{TOKEN_PARAM_NAME, value}
}

func ParamPatternToken(value []byte) Token {
    return Token{TOKEN_PARAM_PATTERN, value}
}

func ParamSeparatorToken() Token {
    return Token{TOKEN_PARAM_SEPARATOR, []byte(string(PARAM_SEPARATOR))}
}

func (token Token) Bytes() []byte {
    return token.Value
}

func (token Token) String() string {
    return string(token.Value)
}

const (
    PATH_SEPARATOR = '/'

    LEFT_CURLY_BRACKET  = '{'
    RIGHT_CURLY_BRACKET = '}'

    COLON = ':'

    PARAM_SEPARATOR = ':'
)
