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

Map Methods
-----------

All of the standard map (dict) methods are available

[testmark]:# (hello-maps/map-methods/script.various/run)
```python
fruits = datalark.Map(_={'a': 'apple', 'b': 'banana'})
print(fruits)
print(fruits.keys())
print(fruits.values())
print(fruits.items())
```

[testmark]:# (hello-maps/map-methods/output)
```text
map{
	string{"a"}: string{"apple"}
	string{"b"}: string{"banana"}
}
list{
	0: string{"a"}
	1: string{"b"}
}
list{
	0: string{"apple"}
	1: string{"banana"}
}
list{
	0: list{
		0: string{"a"}
		1: string{"apple"}
	}
	1: list{
		0: string{"b"}
		1: string{"banana"}
	}
}
```
