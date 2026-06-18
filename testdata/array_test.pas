program arraytest(output);

var
  a: array[1..5] of integer;
  i: integer;

begin
  for i := 1 to 5 do
    a[i] := i * 10;
  writeln(a[3])
end.
