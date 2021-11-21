datalark engine hackme guide
============================

contributing & status
---------------------

This whole shebang is early alpha.  Not all features are complete yet.

We would like to have complete support for basically every feature of IPLD schemas.
We're adding roughly one typekind at a time.
Help adding support for more typekinds would be welcome.

We're kinda "feeling it out" on ergonomics as we go along.
(E.g., what should it "feel like" and when should a user expect kwargs vs positional args, etc.)
In other words: the current code isn't "strongly held opinions"; it can still change.
If you think something is inconsistent or surprising, please open an issue and let's think about it together!

The public APIs of this package aren't super thrilling yet.
(e.g. InjectGlobals and such feel a little weird.)
It's an "engine" internal-ish package, so they don't have to be thrilling.
But if you can think of ways to improve the smoothness,
either here or in the top-level package (which *should* be smooth),
then feedback, suggestions, or even PRs are welcomed.

And if you'd be so kind as to contribute docs or tests if you discover them missing --
please!  Thank you!  You're a hero!  Testing and docs are _so_ important.

Lastly, if you just need to find the folks working on this for a chat:
this project is part of the IPLD community, so you can follow any of the suggestions
in https://github.com/ipld/ipld#finding-us -- there's a matrix chat, and a discord bridged to it,
and you should find some friendly folks there.



testing
-------

Please write tests by writing testmark files in the docs directory.

You should see plenty of examples of how to write those files,
and how to rig them up (it's one function call!) in this package.

You may have to read `testmark.go` to see what the magic labels are;
or, you can probably figure it out just by copying existing docs and following the pattern.

### fixture regeneration

When you're writing a new test, you can just write the schema and your script,
and leave a blank placeholder code block for the output -- we can fill it in automatically!
Just run the tests with this additional argument: `go test ./... -testmark.regen`.

(Run the tests again without that argument afterwards -- that flag means the tests
will almost always pass, because it's assuming the new output is "correct";
you should doublecheck that the result is deterministic before commiting.)

The same goes when updating the expected results for any existing tests, if you've made changes:
run `go test ./... -testmark.regen`, and fixtures will be updated automatically.

Make sure to review all the changes carefully when committing!  `git diff` is your friend.
This command regenerates *all* test fixtures, so a key part of reviewing your change is
to make sure only the fixtures that you _expected_ to change have actually changed.



function and type name conventions
----------------------------------

So far, we've got:

- For each IPLD Schema typekind, there's a struct named that.  Exported.
  (There's not a strong reason for this to be exported, but "We're all consenting adults here".)
  - The names of methods on those types is dictated by the starlark APIs.
  - They also have a `Node` method, for us to unwrap things with.

- For each of the above structs, there's a function named `Construct{typekind}`.
  - See the section below about routers for where you'll need to wire this up.

- For each of the above structs, there's a function named `Wrap{typekind}`.
  - ... maybe.  This hasn't actually seemed very important yet.  Perhaps let's not add more of these until more excuses to use it appear.
  - These are not highly demanded because you can do the same thing with the `Wrap` function.

... and that's about it.

Everything else gets wired up through the router functions,
which are described in the next section.



the big routers
---------------

If you're adding support for a whole new typekind,
you'll need to hook it into these places:

- 1: `prototype.go`: the `Prototype.CallInternal` method is where all the constructors calls cross in from starlark.
  Accordingly, there's a giant demux switch in there to pick out the right constructor function.
  When adding support for a new typekind: add the new constructor functions here.

- 2. `engine.go`: the `Wrap` function is where data going from golang to starlark for the first time all moves through.
  Accordingly, there's a giant demux switch in there to pick out the wrapper type with the correct methods for the starlark side.
  When adding support for a new typekind: add the new wrapper types here.
