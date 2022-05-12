package datalarkengine

func Example_string() {
	mustExecExample(nil, nil,
		"mytypes",
		`
		print(datalark.String("yo"))
	`)

	// Output:
	// string{"yo"}
}
