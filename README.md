go-datalark
===========

Bindings for IPLD in Starlark.

**Starlark** is a dynamic interpreted language that is easy to embed in golang programs
(and also has interpreters in Rust, and in Java -- it can probably go anywhere).
The syntax is python-like and easily approachable for developers from many backgrounds.
(Remarkably, it's also *non-turing complete* and inherently sandboxed -- it has no access to a host enviroment
(unless you intentionally give the interpreter sandbox custom functions that can do that!).)

**IPLD** is a data interchange standard, with emphasis on utility in the decentralized web.

**Marrying Starlark and IPLD** gives you a data interchange format that can talk to many kinds of APIs
and is great for building new protocols...
_and_ gives you a full-fledged (yet sandboxed, safe) langauge you can use for anything
from basic templating (e.g. for config files) all the way to being a generic data transform language.


Features
--------

- Wrap `ipld.Node` values from golang and hand them to a Starlark interpreter to read.

- Get constructor functions you can hand to a Starlark interpreter so it can make basic `ipld.Node`s (maps, strings, lists, etc etc).

- Combined with `bindnode.Wrap`: easily expose arbitrary golang structs and types to a Starlark interpreter environment to read as if they were dicts and lists and regular starlark values.

- Combined with [IPLD Schemas](https://ipld.io/docs/schemas/): get constructor functions you can hand to a Starlark interpreter to make _those types_.
	- ... so that you can access all of IPLD Schema's validation powers from within Starlark!
	- ... so that you can use kwargs initialization to create struct types with syntactic grace.
	- ... so that you can have dotted access to struct fields in the Starlark syntax, just like you'd expect.
	- ... so that, together with `bindnode`, you can fill in golang native structs from Starlark with ease!

- Does it support [ADLs](https://ipld.io/glossary/#adl)?  Of course it does!
	- ADLs are just `ipld.Node`s, so they work with datalark just like any other nodes do.  No fuss.

- Use Starlark's regular `print` function to get the IPLD debug printout format for data, for rapid development and easy debugging.

- Well-behaved as a library: go-datalark doesn't wrap the Starlark interpreter or do any other weird non-composable hacks:
  it just gives you functions and variables that you hand into the Starlark environment
  (and, a few helper functions to make that even easier to do).
  That's it.  It's totally composable.


Status
------

Datalark should currently be considered in an alpha or beta status.
It's working, but there is currently no promise of API stability.
Some features may also be missing.

The best way to increase the stability and completeness of datalark is to start using it!
We're adding features as the demand for them is found.


Contributing
------------

If you'd like to dive deep in on code, there's a [HACKME](engine/HACKME.md) guide
in the engine package which gives some pointers for how to orient yourself in the code.

If you want to get in touch with folks working on this for a chat:
this project is part of the IPLD community, so you can follow any of the suggestions
in https://github.com/ipld/ipld#finding-us -- there's a matrix chat, and a discord bridged to it,
and you should find some friendly folks there.


License
-------

SPDX-License-Identifier: Apache-2.0 OR MIT
