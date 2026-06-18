program conditional(output);

var
  x: integer;

begin
  x := 7;
  if x > 5 then
    writeln('greater')
  else
    writeln('not greater');
  if odd(x) then
    writeln('odd')
end.
