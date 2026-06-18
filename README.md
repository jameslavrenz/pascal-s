# Pascal-S Interpreter

An interpreter for [Pascal-S](https://standardpascal.org/pascals.html), the educational subset of Pascal defined by Niklaus Wirth (1975). Pascal-S includes integer, real, boolean, and char types; arrays and records; procedures and functions; and structured control flow (`if`, `while`, `repeat`, `for`, `case`).

## Build

```bash
make
```

Or:

```bash
go build -o build/bin/pascal-s ./cmd/pascal-s
```

## Run a `.pas` file

```bash
./build/bin/pascal-s testdata/hello.pas
```

Or:

```bash
make run -- testdata/factorial.pas   # or invoke the binary directly
./build/bin/pascal-s testdata/factorial.pas
```

## REPL

Start the interactive REPL:

```bash
make run
```

The REPL accepts:

- Full programs (`program name(output); ... end.`)
- Declaration fragments (`var x: integer;`)
- Statement sequences (`writeln(42); x := x + 1`)

Multi-line program input continues until a line ending with `.` is entered.

## Example program

```pascal
program hello(output);

begin
  writeln('Hello, Pascal-S!')
end.
```

## Test suite

```bash
make test
```

Test programs live in `testdata/` and cover:

- `writeln` / I/O
- Integer arithmetic (`+`, `-`, `*`, `div`, `mod`)
- Functions and recursion
- `if`-`then`-`else`
- `while`, `for`, `repeat`-`until`
- Arrays
- `case` statements
- Real arithmetic

## Pascal-S features supported

| Feature | Status |
|---------|--------|
| `program` / `begin` / `end` | yes |
| `const`, `var`, `type` | yes |
| `integer`, `real`, `boolean`, `char` | yes |
| Arrays, records | yes |
| Procedures and functions | yes |
| `if`, `while`, `repeat`, `for`, `case` | yes |
| `read`, `readln`, `write`, `writeln` | yes |
| Built-ins: `abs`, `sqr`, `odd`, `chr`, `ord`, `succ`, `pred`, `sqrt`, `sin`, `cos`, etc. | yes |

Not implemented (omitted from Pascal-S or out of scope): sets, pointers, file types beyond standard I/O, `goto`, `with`, procedure parameters.
