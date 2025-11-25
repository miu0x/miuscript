package repl

import (
        "bufio"
        "fmt"
        "io"

        "github.com/miu0x/miuscript/evaluator"
        "github.com/miu0x/miuscript/lexer"
        "github.com/miu0x/miuscript/object"
        "github.com/miu0x/miuscript/parser"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
        scanner := bufio.NewScanner(in)
        env := object.NewEnvironment()

        for {
                fmt.Fprintf(out, PROMPT)
                scanned := scanner.Scan()
                if !scanned {
                        return
                }

                line := scanner.Text()
                l := lexer.New(line)
                p := parser.New(l)

                program := p.ParseProgram()
                if len(p.Errors()) != 0 {
                        printParserErrors(out, p.Errors())
                        continue
                }

                evaluated := evaluator.Eval(program, env)
                if evaluated != nil {
                        io.WriteString(out, evaluated.Inspect())
                        io.WriteString(out, "\n")
                }

        }
}

const MIU_FACE = `
        /\_/\  
   /\  ( o.o )   < miuscript parser error >
  /  \  > ^ <     neko can’t parse this… try again.
`

func printParserErrors(out io.Writer, errors []string) {
        io.WriteString(out, MIU_FACE)
        io.WriteString(out, "parser errors:\n")
        for _, msg := range errors {
                io.WriteString(out, "\t"+msg+"\n")
        }
}
