package sashay_test

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

var _ = Describe("Data typing", func() {
	Describe("BuiltinDataTyperFor", func() {
		It("uses the noop typer for a non-builtin type", func() {
			type T struct{}
			dt := sashay.BuiltinDataTyperFor(T{})
			of := sashay.ObjectFields{}
			dt(sashay.Field{}, of)
			Expect(of).To(BeEmpty())
		})
	})
})
