package datalarkengine

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/node/bindnode"
	"github.com/ipld/go-ipld-prime/schema"
	"github.com/warpfork/go-testmark"
)

// n.b. yes there's a reason it uses filename and not a stream; it's so the patcher can work.
func testFixture(t *testing.T, filename string) {
	// TODO(dustmop): Accessor on testmark.Document that gets the basename (for better errors)
	sourceName := filepath.Base(filename)
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
		// If there's no "schema" hunk, it's not one of the test styles we recognize here.  Skip.
		if dir.Children["schema"] == nil {
			continue
		}

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
			testFixtureHelper(t, dir, doc, sourceName, &patches, npts)
		})
	}
}

func testFixtureHelper(t *testing.T, dir *testmark.DirEnt, doc *testmark.Document, sourceName string, patches *testmark.PatchAccumulator, defines []schema.TypedPrototype) {
	// There should be one of:
	// - a "script" hunk (with an "output" sibling);
	// - or a "script.various" hunk, with multiple children (with an "output" sibling);
	// - or if there's anything else, the above two rules apply within it.
	//    (Technically, you can recurse, too, but I don't see why you'd want to.)
	switch {
	case dir.Children["script"] != nil:
		scriptHunk := dir.Children["script"].Hunk
		if scriptHunk == nil {
			t.Fatal("empty hunk found")
		}
		output, err := runScript(defines, "mytypes", string(scriptHunk.Body))
		if err != nil {
			t.Fatal(makeTestmarkError(doc, sourceName, scriptHunk, err))
		}

		if *testmark.Regen {
			patches.AppendPatchIfBodyDiffers(*dir.Children["output"].Hunk, []byte(output))
		} else {
			qt.Assert(t, output, qt.Equals, string(dir.Children["output"].Hunk.Body))
		}
	case dir.Children["script.various"] != nil:
		// This is more different than "run each as if it was called 'script'"
		//  because we also have to figure out what to do if we're in regen mode and their outputs don't match.
		var seen []string
		for _, script := range dir.Children["script.various"].ChildrenList {
			t.Run(script.Name, func(t *testing.T) {
				output, err := runScript(defines, "mytypes", string(script.Hunk.Body))
				if err != nil {
					t.Fatal(makeTestmarkError(doc, sourceName, script.Hunk, err))
				}

				if *testmark.Regen {
					var repeat bool
					for _, sawIt := range seen {
						if sawIt == output {
							repeat = true
						}
					}
					if !repeat {
						seen = append(seen, output)
					}
				} else {
					qt.Assert(t, output, qt.Equals, string(dir.Children["output"].Hunk.Body))
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
				testFixtureHelper(t, dir, doc, sourceName, patches, defines)
			})
		}
	}
}

func makeTestmarkError(doc *testmark.Document, sourceName string, scriptHunk *testmark.Hunk, err error) error {
	dh, ok := doc.HunksByName[scriptHunk.Name]
	if !ok {
		fmt.Errorf("error %s:<unknown>: %w", sourceName, err)
	}
	// NOTE: LineStart+1 because the doc counts from 0, while text editors start at 1
	return fmt.Errorf("error %s:%d: %w", sourceName, dh.LineStart+1, err)
}
