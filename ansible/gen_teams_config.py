tmpl = "  {{name => '{}', network => '10.118.{}.0/24', host => '10.118.{}.0', team_id => '{}', token => '{}', logo => '/data/logos/{}'}},"

lines = []

import random
x = '0123456789abcdef'

names = [
'Команда Лучкиных Вячеславов',
'C4T BuT S4D',
'SPRUSH',
'♿️🅵🅰️🅺🅰️🅿️🅿️🅰️♿️',
]



for i in range(101, 101 + len(names)):
    t = ''.join(random.choice(x) for i in range(20))
    s = tmpl.format(names[i-101], i, i, i, t, i)
    lines.append(s)

print("\n".join(lines))
