package sashay_test

import (
	"fmt"
	"github.com/rgalanakis/sashay"
)

func ExampleSimpleDataTyper() {
	dt := sashay.SimpleDataTyper("string", "date-time")
	fields := dt(sashay.NewField("abc"))
	fmt.Println("Type:", fields["type"], "Format:", fields["format"])
	// Output:
	// Type: string Format: date-time
}

func ExampleChainDataTyper() {
	dt := sashay.ChainDataTyper(
		sashay.SimpleDataTyper("string", "format1"),
		func(tvp sashay.Field) sashay.ObjectFields {
			return sashay.ObjectFields{"format": "format2"}
		})
	fields := dt(sashay.NewField("abc"))
	fmt.Println("Type:", fields["type"], "Format:", fields["format"])
	// Output:
	// Type: string Format: format2
}
