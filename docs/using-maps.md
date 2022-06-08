Using Maps with Datalark
==========================

Map types may be defined in IPLD Schemas, like this:

[testmark]:# (hello-maps/schema)
```ipldsch
type FruitColors {String:String}
```

This is a map that uses both String keys and values.

Creating Maps Values
---------------------

### Creating maps

Maps can be created with kwargs, and indexed using strings:

[testmark]:# (hello-maps/hello-maps/create-kwargs/script.various/kwargs)
```python
m = mytypes.FruitColors(apple="red")
print(m)
print(m['apple'])
```

Here is the output for this code:

[testmark]:# (hello-maps/hello-maps/create-kwargs/output)
```text
map<FruitColors>{
	string<String>{"apple"}: string<String>{"red"}
}
string<String>{"red"}
```
