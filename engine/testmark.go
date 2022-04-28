package datalarkengine

import (
	"bytes"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/node/bindnode"
	"github.com/ipld/go-ipld-prime/schema"
	"github.com/warpfork/go-testmark"
)

func testFixture(t *testing.T, filename string) {
	doc, err := testmark.ReadFile(filename)
	if err != nil {
		t.Fatalf("spec file parse failed?!: %s", err)
	}
	var patches testmark.PatchAccumulator
	defer func() {
		if *testmark.Regen {
			patches.WriteFileWithPatches(doc, filename)
		}
	}()

	// Data hunks should be in "directories" of a test scenario each.
	doc.BuildDirIndex()
	for _, dir := range doc.DirEnt.ChildrenList {
		t.Run(dir.Name, func(t *testing.T) {
			// There should be a "schema" hunk, containing DSL.  Parse it.
			typesystem, err := ipld.LoadSchema("<noname>", bytes.NewReader(dir.Children["schema"].Hunk.Body))
			if err != nil {
				t.Fatalf("invalid schema: %s", err)
			}

			// Produce the prototypes that we'll inject to starlark using datalark.
			var npts []schema.TypedPrototype
			for _, typeInfo := range typesystem.GetTypes() {
				npts = append(npts, bindnode.Prototype(nil, typeInfo))
			}

			// The rest of the processing is in a helper function because it may be recursive (all using the same schema).
			testFixtureHelper(t, dir, &patches, npts)
		})
	}
}

func testFixtureHelper(t *testing.T, dir testmark.DirEnt, patches *testmark.PatchAccumulator, npts []schema.TypedPrototype) {
	// There should be one of:
	// - a "script" hunk (with an "output" sibling);
	// - or a "script.various" hunk, with multiple children (with an "output" sibling);
	// - or if there's anything else, the above two rules apply within it.
	//    (Technically, you can recurse, too, but I don't see why you'd want to.)
	switch {
	case dir.Children["script"] != nil:
		var buf bytes.Buffer
		err := eval_helper(&buf, npts, string(dir.Children["script"].Hunk.Body))
		if err != nil {
			t.Fatalf("script eval failed: %s", err) // TODO probably actually just append this to the buffer for diffing
		}

		if *testmark.Regen {
			patches.AppendPatchIfBodyDiffers(*dir.Children["output"].Hunk, buf.Bytes())
		} else {
			qt.Assert(t, buf.String(), qt.Equals, string(dir.Children["output"].Hunk.Body))
		}
	case dir.Children["script.various"] != nil:
		// This is more different than "run each as if it was called 'script'"
		//  because we also have to figure out what to do if we're in regen mode and their outputs don't match.
		var seen []string
		for _, script := range dir.Children["script.various"].ChildrenList {
			t.Run(script.Name, func(t *testing.T) {
				var buf bytes.Buffer
				err := eval_helper(&buf, npts, string(script.Hunk.Body))
				if err != nil {
					t.Fatalf("script eval failed: %s", err) // TODO probably actually just append this to the buffer for diffing
				}

				if *testmark.Regen {
					var repeat bool
					for _, sawIt := range seen {
						if sawIt == buf.String() {
							repeat = true
						}
					}
					if !repeat {
						seen = append(seen, buf.String())
					}
				} else {
					qt.Assert(t, buf.String(), qt.Equals, string(dir.Children["output"].Hunk.Body))
				}
			})
			if *testmark.Regen {
				qt.Assert(t, seen, qt.HasLen, 1)
				report := strings.Join(seen, "\n--- OR ---\n") // If this join ever actually does something, surely no one is happy, but the diff in future non-regen rounds should be entertaining.
				patches.AppendPatchIfBodyDiffers(*dir.Children["output"].Hunk, []byte(report))
			}
		}
	default:
		for _, dir := range dir.ChildrenList {
			t.Run(dir.Name, func(t *testing.T) {
				testFixtureHelper(t, dir, patches, npts)
			})
		}
	}
}