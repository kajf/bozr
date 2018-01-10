package main

import (
	"net/http"
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
	rememberMap := map[string]interface{}{"savedToken": token}
	vars := &Vars{items: rememberMap}

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
	rememberMap := map[string]interface{}{"savedToken": token, "aSecond": second}
	vars := &Vars{items: rememberMap}

	got := vars.ApplyTo("prefix {savedToken} middle {aSecond} postfix")

	expected := "prefix " + token + " middle " + second + " postfix"
	if got != expected {
		t.Error(
			"expected[", expected,
			"got[", got,
		)
	}
}

func TestExpectPopulateWithNoChange(t *testing.T) {
	path := "items.id"

	expect := &Expect{Body: map[string]interface{}{path: "xyz"}}
	vars := Vars{items: map[string]interface{}{"savedId": "abc"}}

	expect.populateWith(vars)

	if expect.Body[path] != "xyz" || len(expect.Body) != 1 {
		t.Errorf("body was modified, body %v", expect.Body)
	}
}

func TestExpectPopulateWithHeaders(t *testing.T) {

	header := "Key"
	val := "myId"

	expect := &Expect{Headers: map[string]string{header: "{savedId}"}}
	vars := Vars{items: map[string]interface{}{"savedId": val}}

	expect.populateWith(vars)

	if expect.Headers[header] != val {
		t.Errorf("header does not contain val '%s', headers %v", val, expect.Headers)
	}
}

func TestExpectPopulateWithBody(t *testing.T) {

	path := "items.id"
	expect := &Expect{Body: map[string]interface{}{path: "{savedId}"}}

	val := "myId"
	vars := Vars{items: map[string]interface{}{"savedId": val}}

	expect.populateWith(vars)

	if expect.Body[path] != val {
		t.Errorf("body does not contain var '%s', body %v", val, expect.Body)
	}
}

func TestExpectPopulateWithBodyArray(t *testing.T) {

	path := "items.id"
	expect := &Expect{Body: map[string]interface{}{path: []string{"{savedId}", "abc", "{nextId}"}}}

	val := "myId"
	vars := Vars{items: map[string]interface{}{"savedId": val, "nextId": 3}}

	expect.populateWith(vars)

	arr := expect.Body[path].([]string)
	if arr[0] != "myId" || arr[1] != "abc" || arr[2] != "3" {
		t.Errorf("body does not contain var '%s', body %v", val, expect.Body)
	}
}

func TestExpectPopulateWithBodyInt(t *testing.T) {
	expect := &Expect{Body: map[string]interface{}{"items.id": 12}}
	vars := Vars{items: map[string]interface{}{"savedId": "someId"}}

	expect.populateWith(vars)

	if expect.Body["items.id"] != 12 || len(expect.Body) != 1 {
		t.Errorf("body was modified, body %v", expect.Body)
	}
}
