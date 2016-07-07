package main

import (
	"testing"
	"encoding/xml"
)

func TestXmlUnmarshal(t *testing.T) {
	s := `<pay_period_profiles></pay_period_profiles>`
	var m map[string]interface{}
	err := xml.Unmarshal([]byte(s), &m)
	if err != nil {
		t.Error(err)
	}

	//if got != "417857" {
	//	t.Error(
	//		"expected", "417857",
	//		"got", got,
	//	)
	//}
}
