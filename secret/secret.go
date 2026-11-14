// Package secret reads and writes the secret paths of a sync entity's JSON data. Client and server
// share it so both sides agree on what a path names.
//
// A path is a JSON pointer (RFC 6901) with one addition: a segment "@<id>" names the element of an
// array whose "id" field is <id>.
package secret

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
)

// Doc is parsed JSON data.
type Doc struct{ root map[string]any }

// Parse reads data, which must be a JSON object. Empty data is an empty object.
func Parse(data []byte) (*Doc, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return &Doc{root: map[string]any{}}, nil
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var root map[string]any
	if err := dec.Decode(&root); err != nil {
		return nil, fmt.Errorf("data is not JSON: %w", err)
	}
	if root == nil {
		return nil, errors.New("data is not a JSON object")
	}
	if err := dec.Decode(new(json.RawMessage)); err != io.EOF {
		return nil, errors.New("trailing data after JSON object")
	}
	return &Doc{root: root}, nil
}

// Bytes writes the document back.
func (d *Doc) Bytes() ([]byte, error) { return json.Marshal(d.root) }

// Get is the string at path, empty when there is none.
func (d *Doc) Get(path string) string {
	parent, last, ok := d.walk(path)
	if !ok {
		return ""
	}
	value, _ := parent[last].(string)
	return value
}

// Set writes a string at path. The path's parent must exist; false when it does not.
func (d *Doc) Set(path, value string) bool {
	parent, last, ok := d.walk(path)
	if !ok {
		return false
	}
	parent[last] = value
	return true
}

// Valid reports whether path is well formed and its parent exists.
func (d *Doc) Valid(path string) bool {
	_, _, ok := d.walk(path)
	return ok
}

// walk finds the object holding the path's last segment.
func (d *Doc) walk(path string) (map[string]any, string, bool) {
	if !strings.HasPrefix(path, "/") {
		return nil, "", false
	}
	segments := strings.Split(path[1:], "/")
	var at any = d.root
	for i, raw := range segments {
		segment := unescape(raw)
		last := i == len(segments)-1
		switch node := at.(type) {
		case map[string]any:
			if last {
				return node, segment, true
			}
			at = node[segment]
		case []any:
			if last {
				return nil, "", false
			}
			at = element(node, segment)
		default:
			return nil, "", false
		}
		if at == nil {
			return nil, "", false
		}
	}
	return nil, "", false
}

func element(list []any, segment string) any {
	if id, ok := strings.CutPrefix(segment, "@"); ok {
		for _, item := range list {
			if obj, ok := item.(map[string]any); ok && obj["id"] == id {
				return obj
			}
		}
		return nil
	}
	i, err := strconv.Atoi(segment)
	if err != nil || i < 0 || i >= len(list) {
		return nil
	}
	return list[i]
}

func unescape(segment string) string {
	return strings.ReplaceAll(strings.ReplaceAll(segment, "~1", "/"), "~0", "~")
}

// Escape makes a key safe to use as one segment of a path.
func Escape(key string) string {
	return strings.ReplaceAll(strings.ReplaceAll(key, "~", "~0"), "/", "~1")
}

// Withhold empties every secret path that holds a value and names those paths.
func Withhold(data []byte, secrets []string) ([]byte, []string, error) {
	doc, err := Parse(data)
	if err != nil {
		return nil, nil, err
	}
	var withheld []string
	for _, path := range secrets {
		if doc.Get(path) != "" {
			doc.Set(path, "")
			withheld = append(withheld, path)
		}
	}
	if len(withheld) == 0 {
		return data, nil, nil
	}
	slices.Sort(withheld)
	out, err := doc.Bytes()
	return out, withheld, err
}

// Restore fills each withheld path that is empty in data with the value held has there.
func Restore(data []byte, withheld []string, held []byte) ([]byte, error) {
	return restore(data, held, withheld, nil)
}

// Keep restores withheld values. For a sender without read access, it only restores
// paths still marked secret and also considers previously stored secret paths.
func Keep(data []byte, secrets, withheld []string, held []byte, heldSecrets []string, hidden bool) ([]byte, error) {
	if !hidden {
		return Restore(data, withheld, held)
	}
	return restore(data, held, slices.Concat(withheld, secrets), func(path string) bool {
		return slices.Contains(secrets, path) && (slices.Contains(withheld, path) || slices.Contains(heldSecrets, path))
	})
}

func restore(data, held []byte, paths []string, allowed func(string) bool) ([]byte, error) {
	if len(paths) == 0 || len(held) == 0 {
		return data, nil
	}
	doc, err := Parse(data)
	if err != nil {
		return nil, err
	}
	old, err := Parse(held)
	if err != nil {
		return data, nil
	}
	changed := false
	for _, path := range paths {
		if allowed != nil && !allowed(path) {
			continue
		}
		if value := old.Get(path); value != "" && doc.Get(path) == "" && doc.Set(path, value) {
			changed = true
		}
	}
	if !changed {
		return data, nil
	}
	return doc.Bytes()
}
