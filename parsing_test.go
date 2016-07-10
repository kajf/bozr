package main

import (
	"testing"
	"github.com/clbanning/mxj"
)

func TestXmlUnmarshalMxj(t *testing.T) {
	expectedVal := "22"
	s := `<pay_period_profiles>
		<pay_period_profile>123</pay_period_profile>
		<pay_period_profile>` + expectedVal + `</pay_period_profile>
	     </pay_period_profiles>`
	m, err := mxj.NewMapXml([]byte(s))
	if err != nil {
		t.Error(err)
	}

	got, err := getByPath(m.Old(), "pay_period_profiles", "pay_period_profile", "1")
	if got != expectedVal || err != nil {
		t.Error(
			"expected", expectedVal,
			"got", got,
			"err", err,
		)
	}
}
