package datalarkengine

func Example_string() {
	mustExecExample(nil,
		"mytypes",
		`
		print(datalark.String("yo"))
	`)

	// Output:
	// string{"yo"}
}
