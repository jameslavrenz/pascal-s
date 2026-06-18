package runtime_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JimLavrenz/pascal-s/internal/parser"
	"github.com/JimLavrenz/pascal-s/internal/runtime"
)

func runSource(t *testing.T, source string) string {
	t.Helper()
	var out bytes.Buffer
	rt := runtime.New()
	rt.SetOutput(&out)
	prog, err := parser.ParseProgram(source)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := rt.RunProgram(prog); err != nil {
		t.Fatalf("run: %v", err)
	}
	return out.String()
}

func runFile(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return runSource(t, string(data))
}

func TestHello(t *testing.T) {
	out := runFile(t, "hello.pas")
	if !strings.Contains(out, "Hello, Pascal-S!") {
		t.Fatalf("got %q", out)
	}
}

func TestArithmetic(t *testing.T) {
	out := runFile(t, "arithmetic.pas")
	lines := strings.Fields(strings.TrimSpace(out))
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %q", out)
	}
	if lines[0] != "16" || lines[1] != "3" || lines[2] != "1" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestFactorial(t *testing.T) {
	out := runFile(t, "factorial.pas")
	if strings.TrimSpace(out) != "120" {
		t.Fatalf("got %q", out)
	}
}

func TestConditional(t *testing.T) {
	out := runFile(t, "conditional.pas")
	if !strings.Contains(out, "greater") || !strings.Contains(out, "odd") {
		t.Fatalf("got %q", out)
	}
}

func TestWhileLoop(t *testing.T) {
	out := runFile(t, "while_test.pas")
	if strings.TrimSpace(out) != "15" {
		t.Fatalf("got %q", out)
	}
}

func TestForLoop(t *testing.T) {
	out := runFile(t, "for_test.pas")
	if strings.TrimSpace(out) != "55" {
		t.Fatalf("got %q", out)
	}
}

func TestArray(t *testing.T) {
	out := runFile(t, "array_test.pas")
	if strings.TrimSpace(out) != "30" {
		t.Fatalf("got %q", out)
	}
}

func TestRepeatUntil(t *testing.T) {
	out := runFile(t, "repeat_test.pas")
	if strings.TrimSpace(out) != "4" {
		t.Fatalf("got %q", out)
	}
}

func TestConstAndTypes(t *testing.T) {
	source := `program t(output);
const max = 10;
var n: integer;
begin
  n := max;
  writeln(n)
end.`
	out := runSource(t, source)
	if strings.TrimSpace(out) != "10" {
		t.Fatalf("got %q", out)
	}
}

func TestRealArithmetic(t *testing.T) {
	source := `program t(output);
var x, y: real;
begin
  x := 3.5;
  y := 2.0;
  writeln(x / y)
end.`
	out := runSource(t, source)
	if !strings.HasPrefix(strings.TrimSpace(out), "1.75") {
		t.Fatalf("got %q", out)
	}
}

func TestCaseStatement(t *testing.T) {
	source := `program t(output);
var x: integer;
begin
  x := 2;
  case x of
    1: writeln('one');
    2: writeln('two')
  end
end.`
	out := runSource(t, source)
	if strings.TrimSpace(out) != "two" {
		t.Fatalf("got %q", out)
	}
}

func TestREPLStatements(t *testing.T) {
	var out bytes.Buffer
	rt := runtime.New()
	rt.SetOutput(&out)

	block, err := parser.ParseREPL("writeln(42)")
	if err != nil {
		t.Fatal(err)
	}
	if err := rt.RunBlock(block); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out.String()) != "42" {
		t.Fatalf("got %q", out.String())
	}
}
