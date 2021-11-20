package datalarkengine

func Example_string() {
	evalExample(nil, `
		print(datalark.String("yo"))
	`)

	// Output:
	// string{"yo"}
}
