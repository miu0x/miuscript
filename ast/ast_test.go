package ast

import (
        "testing"

        "github.com/miu0x/miuscript/token"
)

func TestString(t *testing.T) {
        // let myvar = anotherVar
        program := []Program{
                &LetStatement{
                        Token: token.Token{Token: token.LET, Literal: "let"},
                        Name: &Identifier{
                                Token:token.Token{ Token: token.IDENT, Literal: "myvar"},
                                Value: "myvar"
                        },
                        Value: &Identifier {
                                Token:token.Token{ Token: token.IDENT, Literal: "anotherVar"},
                                Expression: "anotherVar"
                        }        
                }
        }
        if program.String() != "let myvar = anotherVar;" {
                t.Errof("program.String() wron, got=%q", program.String())
        }

}
