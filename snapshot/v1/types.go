// Code generated from schema/snapshot.v1.json by quicktype. DO NOT EDIT.
package snapshot

import "time"

import "encoding/json"

func UnmarshalSnapshot(data []byte) (Snapshot, error) {
	var r Snapshot
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Snapshot) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// Current v1 writer model. Readers preserve compatibility with earlier v1 documents by
// applying zero defaults for omitted fields. Arrays may be null in Go JSON output.
type Snapshot struct {
	Format       string        `json:"format"`
	Version      int64         `json:"version"`
	Generator    Generator     `json:"generator"`
	Locale       string        `json:"locale"`
	Collection   Collection    `json:"collection"`
	Environment  Environment   `json:"environment"`
	Environments []Environment `json:"environments"`
	Prose        Prose         `json:"prose"`
	Changes      []Release     `json:"changes"`
	Options      Options       `json:"options"`
}

type Release struct {
	At       time.Time `json:"at"`
	Items    []Item    `json:"items"`
	Revision int64     `json:"revision"`
}

type Item struct {
	Sign   string   `json:"sign"`
	Method string   `json:"method"`
	Name   string   `json:"name"`
	Fields []string `json:"fields"`
}

type Collection struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Groups      []Group    `json:"groups"`
	Variables   []Variable `json:"variables"`
}

type Group struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Endpoints   []Endpoint `json:"endpoints"`
}

type Endpoint struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Method      string      `json:"method"`
	Path        string      `json:"path"`
	Description string      `json:"description"`
	Parameters  []Parameter `json:"parameters"`
	Body        *Body       `json:"body"`
	Responses   []Response  `json:"responses"`
}

type Body struct {
	Fields  []Field `json:"fields"`
	Example string  `json:"example"`
}

type Field struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

type Parameter struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Example     string `json:"example"`
}

type Response struct {
	Status int64  `json:"status"`
	Name   string `json:"name"`
	Body   string `json:"body"`
}

type Variable struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Environment struct {
	Name    string `json:"name"`
	BaseURL string `json:"baseUrl"`
}

type Generator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Options struct {
	Docs     bool `json:"docs"`
	Examples bool `json:"examples"`
	TryIt    bool `json:"tryIt"`
}

type Prose struct {
	Lead   string `json:"lead"`
	Auth   string `json:"auth"`
	Errors string `json:"errors"`
	Limits string `json:"limits"`
}

const Format = "schemo.collection-snapshot"
const Version = 1
