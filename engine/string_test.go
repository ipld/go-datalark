package datalarkengine

func Example_string() {
	evalExample(`
		print(datalark.String("yo"))
	`, nil)

	// Output:
	// string{"yo"}
}
