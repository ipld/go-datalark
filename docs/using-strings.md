Using Strings with Datalark
==========================

Datalark has its own string type

TODO: testmark requires a schema, it should be changed to be optional, for
tests like this.

[testmark]:# (hello-strings/schema)
```ipldsch
type RemoveMe {String:String}
```

String Methods
--------------

All of the standard string methods are available

[testmark]:# (hello-strings/string-methods/script.various/run)
```python
text = datalark.String('Hello There')
print(text)
print(text.upper())
print(text.lower())
print(text.count('l'))
print(len(text))
```

[testmark]:# (hello-strings/string-methods/output)
```text
string{"Hello There"}
string{"HELLO THERE"}
string{"hello there"}
int{2}
11
```
