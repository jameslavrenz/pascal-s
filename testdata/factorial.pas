program factorial(output);

var
  n, i, result: integer;

function fact(x: integer): integer;
begin
  if x <= 1 then
    fact := 1
  else
    fact := x * fact(x - 1)
end;

begin
  n := 5;
  result := fact(n);
  writeln(result)
end.
