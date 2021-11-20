package datalarkengine

func Example_map() {
	evalExample(nil, `
		x = {"bz": "zoo"}
		print(datalark.Map(hey="hai", zonk="wot", **x))
		print(datalark.Map({datalark.String("fun"): "heeey"}))
	`)

	// Output:
	// map{
	// 	string{"hey"}: string{"hai"}
	// 	string{"zonk"}: string{"wot"}
	// 	string{"bz"}: string{"zoo"}
	// }
	// map{
	// 	string{"fun"}: string{"heeey"}
	// }
}
