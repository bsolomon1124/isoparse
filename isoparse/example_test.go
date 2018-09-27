package isoparse

import "fmt"

func ExampleParseISODatetime() {
	fmt.Printf("%25v\t%25v\t%30v\n", "format", "string", "result")
	fmt.Printf("%25v\t%25v\t%30v\n", "------", "------", "------")
	for datetime, c := range allFormats {
		dt, _ := ParseISODatetime(datetime)
		fmt.Printf("%25v\t%25v\t%30v\n", c.f, datetime, dt)
	}
}
