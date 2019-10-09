package main

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestResponseBodyOnce(t *testing.T) {
	resp := Response{
		http: &http.Response{
			Header: map[string][]string{"Content-Type": {"application/json"}},
		},
		body: []byte(`{"key":true}`),
	}

	resp.Body() // first call to parse valid body
	resp.body = []byte(`# set invalid body so it fails if parsed`)

	parsedBody, err := resp.Body()
	if parsedBody == nil || err != nil {
		t.Error("body", parsedBody, "err", err)
	}
}
func TestParseEmptyResponse(t *testing.T) {
	resp := Response{
		body: make([]byte, 0),
		http: &http.Response{
			Header: map[string][]string{"Content-Type": {"application/json"}},
		},
	}

	data, err := resp.Body()
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	if data != nil {
		t.Error("Unexpected data.")
	}
}

func TestTimeFrameExtendStart(t *testing.T) {
	// given
	//   [----] tf
	// [------] tf2
	t1 := time.Now()
	t2 := t1.Add(time.Minute)

	tf := TimeFrame{Start: t1, End: t2}

	tf2 := TimeFrame{Start: t1.Add(-time.Second), End: t2}

	// when
	tf.Extend(tf2)

	// then
	if tf.Start != tf2.Start {
		t.Error("wrong start extension", tf.Start)
	}
}

func TestTimeFrameExtendEnd(t *testing.T) {
	// given
	// [----] tf
	// [------] tf2
	t1 := time.Now()
	t2 := t1.Add(time.Minute)

	tf := TimeFrame{Start: t1, End: t2}
	tf2 := TimeFrame{Start: t1, End: t2.Add(time.Second)}

	// when
	tf.Extend(tf2)

	// then
	if tf.End != tf2.End {
		t.Error("wrong end extension", tf.End)
	}
}

func TestTimeFrameExtendNoExtension(t *testing.T) {
	// given
	// [----] tf
	//  [-] tf2
	t1 := time.Now()
	t2 := t1.Add(time.Minute)

	tf := TimeFrame{Start: t1, End: t2}
	tf2 := TimeFrame{Start: t1.Add(time.Second), End: t2.Add(-time.Second)}

	// when
	tf.Extend(tf2)

	// then
	if tf.Start != t1 || tf.End != t2 {
		t.Error("wrong extension", tf.Start, tf.End)
	}
}

func TestTimeFrameExtendNoIntersection(t *testing.T) {
	// given
	// [--]       tf
	//       [--] tf2
	t1 := time.Now()
	t2 := t1.Add(time.Second)

	tf := TimeFrame{Start: t1, End: t2}
	tf2 := TimeFrame{Start: t1.Add(time.Minute), End: t2.Add(time.Minute)}

	// when
	tf.Extend(tf2)

	// then
	if tf.Start != t1 || tf.End != tf2.End {
		t.Error("wrong intersection extension", tf.Start, tf.End)
	}
}

func TestThrottleFixedSize(t *testing.T) {
	requestLimit := 2
	tr := NewThrottle(requestLimit, 50*time.Millisecond)

	tr.RunOrPause()
	tr.RunOrPause()
	tr.RunOrPause() // should pause on this one

	if len(tr.queue) != requestLimit {
		t.Error("unexpected length " + string(len(tr.queue)))
	}
}

func TestThrottleCleanOld(t *testing.T) {
	tr := NewThrottle(3, time.Second)
	now := time.Now()

	// fake queue
	q := make([]time.Time, 0)
	q = append(q, now.Add(-9*time.Second))
	q = append(q, now.Add(-7*time.Second))
	q = append(q, now)

	tr.queue = q
	// --

	tr.cleanOld()

	if len(tr.queue) != 1 {
		t.Error("unexpected length ", len(tr.queue))
	}
}

func TestThrottleFirstCall(t *testing.T) {
	tr := NewThrottle(3, time.Second)

	tr.RunOrPause()

	if len(tr.queue) != 1 {
		t.Error("unexpected length ", len(tr.queue))
	}
}

func TestThrottleZeroLimit(t *testing.T) {
	tr := NewThrottle(0, time.Second)

	tr.RunOrPause()

	// no NPE, no timeout
}

func TestVarsApplyTo(t *testing.T) {
	token := "test_token"

	vars := NewVars("")
	vars.AddAll(map[string]interface{}{"savedToken": token})

	got := vars.ApplyTo("bearer {savedToken}")

	if got != "bearer "+token {
		t.Error(
			"expected", "bearer "+token,
			"got", got,
		)
	}
}

func TestVarsApplyToMultiple(t *testing.T) {
	token := "test_token"
	second := "second"

	vars := NewVars("")
	vars.AddAll(map[string]interface{}{"savedToken": token, "aSecond": second})

	got := vars.ApplyTo("prefix {savedToken} middle {aSecond} postfix")

	expected := "prefix " + token + " middle " + second + " postfix"
	if got != expected {
		t.Error(
			"expected[", expected,
			"got[", got,
		)
	}
}

func TestVarsApplyToNestedReference(t *testing.T) {
	vars := NewVars("")
	vars.AddAll(map[string]interface{}{"id": "4256", "username": "RU{id}", "key": "{username}"})

	got := vars.ApplyTo("{key}")

	if got != "RU4256" {
		t.Error(
			"expected", "RU4256",
			"got", got,
		)
	}
}

func TestVarsApplyToNestedTemplate(t *testing.T) {
	vars := NewVars("")
	vars.AddAll(map[string]interface{}{"id": "{{ .Base64 `BOZR` }}", "username": "RU{id}", "key": "{{ .SHA1 `{username}` }}"})

	got := vars.ApplyTo("{key}")

	if got != "5365cfcc94c3b65eda62adcc1d6b743d867a4625" {
		t.Error(
			"expected", "5365cfcc94c3b65eda62adcc1d6b743d867a4625",
			"got", got,
		)
	}
}

func TestVarsNotInitializedWithRecoursiveReferences(t *testing.T) {
	vars := NewVars("")

	err := vars.AddAll(map[string]interface{}{
		"username": "RU{username}",
	})

	if err != nil {
		t.Error("Unexpected error", err.Error())
		return
	}

	got := vars.ApplyTo("{username}")

	if got != "RU{username}" {
		t.Error("Expected", "RU{username}", "got", got)
	}
}

func TestVarsIgnoreAddedRecoursiveReference(t *testing.T) {
	vars := NewVars("")

	err := vars.Add("username", "BY{username}")

	if err != nil {
		t.Error("Unexpected error", err.Error())
		return
	}

	got := vars.ApplyTo("{username}")

	if got != "BY{username}" {
		t.Error("Expected", "BY{username}", "got", got)
	}
}

func TestExpectPopulateWithNoChange(t *testing.T) {
	path := "items.id"

	expect := &Expect{BPath: map[string]interface{}{path: "xyz"}}
	vars := NewVars("")
	vars.Add("savedId", "abc")

	expect.populateWith(vars)

	if expect.BodyPath()[path] != "xyz" || len(expect.BodyPath()) != 1 {
		t.Errorf("body was modified, body %v", expect.BodyPath())
	}
}

func TestExpectPopulateWithHeaders(t *testing.T) {

	header := "Key"
	val := "myId"

	expect := &Expect{Headers: map[string]string{header: "{savedId}"}}
	vars := NewVars("")
	vars.Add("savedId", val)

	expect.populateWith(vars)

	if expect.Headers[header] != val {
		t.Errorf("header does not contain val '%s', headers %v", val, expect.Headers)
	}
}

func TestExpectPopulateWithBody(t *testing.T) {

	path := "items.id"
	expect := &Expect{BPath: map[string]interface{}{path: "{savedId}"}}

	val := "myId"
	vars := NewVars("")
	vars.Add("savedId", val)

	expect.populateWith(vars)

	if expect.BodyPath()[path] != val {
		t.Errorf("body does not contain var '%s', body %v", val, expect.BodyPath())
	}
}

func TestExpectPopulateWithBodyArray(t *testing.T) {

	path := "items.id"
	expect := &Expect{BPath: map[string]interface{}{path: []string{"{savedId}", "abc", "{nextId}"}}}

	val := "myId"
	vars := NewVars("")
	vars.AddAll(map[string]interface{}{"savedId": val, "nextId": 3})

	expect.populateWith(vars)

	arr := expect.BodyPath()[path].([]string)
	if arr[0] != "myId" || arr[1] != "abc" || arr[2] != "3" {
		t.Errorf("body does not contain var '%s', body %v", val, expect.BodyPath())
	}
}

func TestExpectPopulateWithBodyInt(t *testing.T) {
	expect := &Expect{BPath: map[string]interface{}{"items.id": 12}}
	vars := NewVars("")
	vars.Add("savedId", "someId")

	expect.populateWith(vars)

	if expect.BodyPath()["items.id"] != 12 || len(expect.BodyPath()) != 1 {
		t.Errorf("body was modified, body %v", expect.BodyPath())
	}
}

func TestOnBodyContentRemovesStartEndDoubleQuotes(t *testing.T) {
	on := &On{Body: []byte("\"abc\"")}

	s, err := on.BodyContent("")

	if err != nil || strings.HasPrefix(s, "\"") || strings.HasSuffix(s, "\"") {
		t.Error("Double quotes was not removed", s, err)
	}
}

func TestOnBodyContentKeepsMiddleDoubleQuotes(t *testing.T) {
	initialString := "abc \"middle\" ending"
	on := &On{Body: []byte(initialString)}

	s, err := on.BodyContent("")

	if err != nil || s != initialString {
		t.Error(s, err)
	}
}

func TestOnBodyContentKeepsSingleQuotes(t *testing.T) {
	initialString := "'abc'"
	on := &On{Body: []byte(initialString)}

	s, err := on.BodyContent("")

	if err != nil || s != initialString {
		t.Error(s, err)
	}
}

func TestPopulateProperty_Map(t *testing.T) {
	vars := NewVars("")
	_ = vars.Add("username", "dpfg")

	tmpl := NewTemplateContext(vars)
	body := map[string]interface{}{
		"username": "{username}",
	}

	result := populateProperty(tmpl, body).(map[string]interface{})["username"]

	if result != "dpfg" {
		t.Errorf("Unexpected populated value: %s", result)
	}
}

func TestPopulateProperty_ArrayOfStrings(t *testing.T) {
	vars := NewVars("")
	_ = vars.Add("username", "dpfg")

	tmpl := NewTemplateContext(vars)
	body := []string{"{username}", "abc123"}

	result := populateProperty(tmpl, body).([]string)

	if result[0] != "dpfg" {
		t.Errorf("Unexpected populated value: %s", result[0])
	}

	if result[1] != "abc123" {
		t.Errorf("Unexpected populated value: %s", result[1])
	}
}

func TestPopulateProperty_ArrayOfInt(t *testing.T) {
	vars := NewVars("")
	_ = vars.Add("username", "dpfg")

	tmpl := NewTemplateContext(vars)
	body := []int{12, 3}

	result := populateProperty(tmpl, body).([]int)

	if result[0] != 12 || result[1] != 3 {
		t.Errorf("Unexpected populated value: %d", result[0])
	}

}

func TestVarsApplyToWithContextBaseUrl(t *testing.T) {
	baseUrl := "http://127.0.0.1/abc"
	vars := NewVars(baseUrl)

	got := vars.ApplyTo(`{ctx:base_url}/my-resource`)

	if got != baseUrl+"/my-resource" {
		t.Error(
			"expected", baseUrl+"/my-resource",
			"got", got,
		)
	}
}

func TestVarsApplyToWithEmptyContextBaseUrl(t *testing.T) {
	vars := NewVars("")

	got := vars.ApplyTo(`{ctx:base_url}/my-resource`)

	if got != "/my-resource" {
		t.Error(
			"expected", "/my-resource",
			"got", got,
		)
	}
}

func TestVarsUnused_EnvVar_NotReported(t *testing.T) {

	vars := NewVars("")

	vars.ApplyTo("{}")

	unused := vars.Unused()
	if len(unused) != 0 {
		t.Error("Unexpected", unused)
	}
}

func TestVarsUnused_CtxVar_NotReported(t *testing.T) {

	vars := NewVars("http://127.0.0.1/abc")

	vars.ApplyTo("{}")

	unused := vars.Unused()
	if len(unused) != 0 {
		t.Error("Unexpected", unused)
	}
}

func TestVarsAdd_CtxVarOverride_Err(t *testing.T) {

	initialCtxValue := "http://127.0.0.1/abc"
	vars := NewVars(initialCtxValue)

	ctxNameUsed := ctxVarPrefix + varPrefixSeparator + "base_url"
	err := vars.Add(ctxNameUsed, "abc")

	if err == nil || vars.items[ctxNameUsed] != initialCtxValue {
		t.Errorf("Expected (var override) error not thrown [%v] or value was overridden %#v. Expected %#v", err, vars.items[ctxNameUsed], initialCtxValue)
	}
}

func TestVarsAdd_VarOverride_Err(t *testing.T) {
	vars := NewVars("")
	duplicateName := "first"
	initialVal := "a"
	errInit := vars.Add(duplicateName, initialVal)
	if errInit != nil {
		t.Error("initialization error", errInit)
	}

	err := vars.Add(duplicateName, "b")

	if err == nil || vars.items[duplicateName] != initialVal {
		t.Errorf("Expected (var override) error not thrown [%v] or value was overridden %#v. Expected %#v", err, vars.items[duplicateName], initialVal)
	}
}

func TestVarsAddAll_VarOverride_Err(t *testing.T) {
	vars := NewVars("")
	duplicateName := "second"
	initialVal := "b"
	errInit := vars.AddAll(map[string]interface{}{"first": 1, duplicateName: initialVal})
	if errInit != nil {
		t.Error("initialization error", errInit)
	}

	err := vars.AddAll(map[string]interface{}{duplicateName: 2, "third": 3})

	if err == nil || vars.items[duplicateName] != initialVal {
		t.Errorf("Expected (var override) error not thrown [%v] or value was overridden %#v. Expected %#v", err, vars.items[duplicateName], initialVal)
	}
}

func TestVarsUnused(t *testing.T) {

	vars := NewVars("")
	unusedVarName := "unusedVar"
	_ = vars.Add("myVar", "abc")
	_ = vars.Add(unusedVarName, "xyz")

	vars.ApplyTo("{myVar}")

	unused := vars.Unused()
	if unused[0] != unusedVarName || len(unused) != 1 {
		t.Error("Unexpected", unused, "should be [", unusedVarName, "]")
	}
}
