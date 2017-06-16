package main

import (
	"net/http"
	"testing"
	"time"
)

func TestResponseBodyOnce(t *testing.T) {
	resp := Response{
		http: http.Response{
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
		http: http.Response{
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
