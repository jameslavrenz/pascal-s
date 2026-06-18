program whiletest(output);

var
  n, sum: integer;

begin
  n := 1;
  sum := 0;
  while n <= 5 do
  begin
    sum := sum + n;
    n := n + 1
  end;
  writeln(sum)
end.
