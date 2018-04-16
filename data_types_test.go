package sashay_test

import (
	"fmt"
	"github.com/rgalanakis/sashay"
)

func ExampleSimpleDataTyper() {
	dt := sashay.SimpleDataTyper("string", "date-time")
	fields := sashay.ObjectFields{}
	dt(sashay.NewField("abc"), fields)
	fmt.Println("Type:", fields["type"], "Format:", fields["format"])
	// Output:
	// Type: string Format: date-time
}

func ExampleChainDataTyper() {
	dt := sashay.ChainDataTyper(
		sashay.SimpleDataTyper("string", "format1"),
		func(_ sashay.Field, of sashay.ObjectFields) {
			of["format"] = "format2"
		})
	fields := sashay.ObjectFields{}
	dt(sashay.NewField("abc"), fields)
	fmt.Println("Type:", fields["type"], "Format:", fields["format"])
	// Output:
	// Type: string Format: format2
}
