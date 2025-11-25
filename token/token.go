package token

type TokenType string

type Token struct {
        Type    TokenType
        Literal string
}

const (
        ILLEGAL = "ILLEGAL"
        EOF     = "EOF"

        // Identfiers + literals
        IDENT = "IDENT"
        INT   = "INT"

        // Operators
        ASSIGN   = "="
        PLUS     = "+"
        MINUS    = "-"
        ASTERICK = "*"
        BANG     = "!"
        SLASH    = "/"

        // Delimiters
        COMMA     = ","
        SEMICOLON = ";"
        COLON     = ":"

        LPAREN   = "("
        RPAREN   = ")"
        LBRACE   = "{"
        RBRACE   = "}"
        LBRACKET = "["
        RBRACKET = "]"

        EQ     = "=="
        NOT_EQ = "!="
        LT     = "<"
        GT     = ">"

        LET      = "LET"
        FUNCTION = "FUNCTION"
        IF       = "IF"
        ELSE     = "ELSE"
        RETURN   = "RETURN"
        TRUE     = "TRUE"
        FALSE    = "FALSE"

        STRING = "STRING"
)

var keywords = map[string]TokenType{
        "let":    LET,
        "fn":     FUNCTION,
        "if":     IF,
        "else":   ELSE,
        "return": RETURN,
        "true":   TRUE,
        "false":  FALSE,
}

func LookUpIdentifer(ident string) TokenType {
        if tok, ok := keywords[ident]; ok {
                return tok
        }
        return IDENT
}
