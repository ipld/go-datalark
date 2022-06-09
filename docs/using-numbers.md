Using Numbers with Datalark
==========================

Datalark has both ints and floats

TODO: testmark requires a schema, it should be changed to be optional, for
tests like this.

[testmark]:# (hello-numbers/schema)
```ipldsch
type RemoveMe {String:String}
```

[testmark]:# (hello-numbers/hello-numbers/int/script.various/kwargs)
```python
a = datalark.Int(3)
b = datalark.Int(4)
c = a + b
d = c * 7
e = d - a
f = e / 2
print(f)
```

[testmark]:# (hello-numbers/hello-numbers/int/output)
```text
int{23}
```

[testmark]:# (hello-numbers/hello-numbers/float/script.various/kwargs)
```python
a = datalark.Float(3.1)
b = datalark.Float(5.2)
c = a + b
d = c * 7.8
e = d - a
f = e / 2.6
print(f)
```

TODO: testmark should have a almost-equals comparator for floats

[testmark]:# (hello-numbers/hello-numbers/float/output)
```text
float{23.70769230769231}
```
