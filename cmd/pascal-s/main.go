package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/JimLavrenz/pascal-s/internal/parser"
	"github.com/JimLavrenz/pascal-s/internal/runtime"
)

const prompt = "pascal-s>>> "

func main() {
	if len(os.Args) > 1 {
		if err := runFile(os.Args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}
	runREPL()
}

func runFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	source := string(data)
	rt := runtime.New()
	prog, err := parser.ParseProgram(source)
	if err != nil {
		return err
	}
	return rt.RunProgram(prog)
}

func runREPL() {
	fmt.Println("Pascal-S interpreter")
	fmt.Print("Type Pascal-S statements or a full program. Exit with Ctrl-D.\n")
	fmt.Print(prompt)

	rt := runtime.New()
	scanner := bufio.NewScanner(os.Stdin)
	var pending strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		if pending.Len() > 0 {
			pending.WriteByte('\n')
		}
		pending.WriteString(line)

		source := pending.String()
		if !replInputComplete(source) {
			fmt.Print("...     ")
			continue
		}

		source = pending.String()
		pending.Reset()

		if strings.TrimSpace(source) == "" {
			fmt.Print(prompt)
			continue
		}

		if err := runREPLInput(rt, source); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
		fmt.Print(prompt)
	}
	fmt.Println()
}

func replInputComplete(source string) bool {
	trimmed := strings.TrimSpace(strings.ToLower(source))
	if strings.HasPrefix(trimmed, "program ") {
		return strings.HasSuffix(trimmed, ".")
	}
	if strings.HasPrefix(trimmed, "var ") || strings.HasPrefix(trimmed, "const ") ||
		strings.HasPrefix(trimmed, "type ") {
		return strings.HasSuffix(trimmed, ";")
	}
	return true
}

func runREPLInput(rt *runtime.Runtime, source string) error {
	trimmed := strings.TrimSpace(strings.ToLower(source))
	if strings.HasPrefix(trimmed, "program ") {
		prog, err := parser.ParseProgram(source)
		if err != nil {
			return err
		}
		return rt.RunProgram(prog)
	}

	block, err := parser.ParseREPL(source)
	if err != nil {
		return err
	}
	if block == nil {
		return nil
	}
	return rt.RunBlock(block)
}
