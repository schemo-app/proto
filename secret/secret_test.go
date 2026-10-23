package secret

import (
	"reflect"
	"testing"
)

func TestWithholdAndRestoreFollowAnElementByID(t *testing.T) {
	data := []byte(`{"auth":{"type":"bearer","fields":{"token":"t0k"}},` +
		`"variables":[{"id":"a","value":"one"},{"id":"b","value":"two"}]}`)
	secrets := []string{"/auth/fields/token", "/variables/@b/value", "/variables/@gone/value"}

	sent, withheld, err := Withhold(data, secrets)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"/auth/fields/token", "/variables/@b/value"}; !reflect.DeepEqual(withheld, want) {
		t.Fatalf("withheld = %v, want %v", withheld, want)
	}
	doc, _ := Parse(sent)
	if doc.Get("/auth/fields/token") != "" || doc.Get("/variables/@b/value") != "" ||
		doc.Get("/variables/@a/value") != "one" {
		t.Fatalf("sent = %s", sent)
	}

	// The other side reordered the list; the secret still lands on its own element.
	reordered := []byte(`{"auth":{"type":"bearer","fields":{"token":""}},` +
		`"variables":[{"id":"b","value":""},{"id":"a","value":"one"}]}`)
	back, err := Restore(reordered, withheld, data)
	if err != nil {
		t.Fatal(err)
	}
	doc, _ = Parse(back)
	if doc.Get("/auth/fields/token") != "t0k" || doc.Get("/variables/@b/value") != "two" {
		t.Errorf("restored = %s", back)
	}
}

func TestAPathOutsideTheDocumentIsNotValid(t *testing.T) {
	doc, _ := Parse([]byte(`{"a":{"b":"c"},"n":1}`))
	for path, want := range map[string]bool{
		"/a/b": true, "/a/x": true, "/x/y": false, "a/b": false, "/n/x": false,
	} {
		if got := doc.Valid(path); got != want {
			t.Errorf("Valid(%q) = %v, want %v", path, got, want)
		}
	}
	if _, err := Parse([]byte(`[1]`)); err == nil {
		t.Error("an array was taken as data")
	}
}

func TestNumbersSurviveARoundTrip(t *testing.T) {
	out, _, err := Withhold([]byte(`{"expiresAt":1767225600123,"value":"x"}`), []string{"/value"})
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != `{"expiresAt":1767225600123,"value":""}` {
		t.Errorf("out = %s", out)
	}
}

// What a sender withheld keeps the held value; what it cleared on purpose does not.
func TestKeepPutsBackOnlyWhatWasNotGiven(t *testing.T) {
	held := []byte(`{"value":"held"}`)
	in := []byte(`{"value":""}`)
	secrets := []string{"/value"}
	value := func(data []byte) string {
		doc, _ := Parse(data)
		return doc.Get("/value")
	}
	if got, _ := Keep(in, secrets, secrets, held, secrets, false); value(got) != "held" {
		t.Errorf("a withheld value came back as %s", got)
	}
	if got, _ := Keep(in, secrets, nil, held, secrets, false); value(got) != "" {
		t.Errorf("a value cleared by a reader of secrets came back as %s", got)
	}
	if got, _ := Keep(in, secrets, nil, held, secrets, true); value(got) != "held" {
		t.Errorf("a secret never handed to its sender was lost: %s", got)
	}
	if got, _ := Keep(in, nil, nil, held, secrets, true); value(got) != "" {
		t.Errorf("a slot made readable was refilled with the secret: %s", got)
	}
}
