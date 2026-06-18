program repeat_test(output);

var
  n: integer;

begin
  n := 0;
  repeat
    n := n + 1
  until n >= 4;
  writeln(n)
end.
