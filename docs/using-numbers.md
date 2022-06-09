Using Numbers with Datalark
==========================

Datalark has both ints and floats

TODO: testmark requires a schema, it should be changed to be optional, for
tests like this.

[testmark]:# (hello-numbers/schema)
```ipldsch
type FruitColors {String:String}
```

[testmark]:# (hello-numbers/hello-numbers/create/script.various/kwargs)
```python
a = datalark.Int(3)
b = datalark.Int(4)
c = a + b
d = c * 7
e = d - a
f = e / 2
print(f)
```

[testmark]:# (hello-numbers/hello-numbers/create/output)
```text
int{23}
```
