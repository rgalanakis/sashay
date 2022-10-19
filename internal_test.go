package sashay

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"reflect"
)

var _ = Describe("Swagger internals test", func() {
	Describe("jsonName", func() {
		type Tester struct {
			Dash       int `json:"-"`
			None       int
			NoneMapped int `json:",omitempty"`
			Mapped     int `json:"mapped"`
		}
		testerType := reflect.TypeOf(Tester{})

		It("returns empty string for json tag value of -", func() {
			f, _ := testerType.FieldByName("Dash")
			Expect(jsonName(f)).To(Equal(""))
		})

		It("returns empty string for no JSON tag", func() {
			f, _ := testerType.FieldByName("None")
			Expect(jsonName(f)).To(Equal(""))
		})

		It("returns the field name if there is no JSON name mapped (',omitempty')", func() {
			f, _ := testerType.FieldByName("NoneMapped")
			Expect(jsonName(f)).To(Equal("NoneMapped"))
		})

		It("returns mapped json name", func() {
			f, _ := testerType.FieldByName("Mapped")
			Expect(jsonName(f)).To(Equal("mapped"))
		})
	})

	Describe("isExportedField", func() {
		It("is true for anonymous fields", func() {
			type T struct {
				int
			}
			Expect(isExportedField(reflect.TypeOf(T{}).Field(0))).To(BeTrue())
		})

		It("is true for unnamed fields (not sure if happens in real world)", func() {
			type T struct {
				field int
			}
			f := reflect.TypeOf(T{}).Field(0)
			f.Name = ""
			Expect(isExportedField(f)).To(BeTrue())
		})

		It("is true for fields with an exported name", func() {
			type T struct {
				Field int `json:"field"`
			}
			Expect(isExportedField(reflect.TypeOf(T{}).Field(0))).To(BeTrue())
		})

		It("is false for fields with an unexported name", func() {
			//noinspection GoStructTag
			type T struct {
				field int `path:"field"`
			}
			Expect(isExportedField(reflect.TypeOf(T{}).Field(0))).To(BeFalse())
		})
	})

	Describe("isExportedName", func() {
		It("panics if used with an empty string", func() {
			Expect(func() {
				isExportedName("")
			}).To(Panic())
		})
	})
})
