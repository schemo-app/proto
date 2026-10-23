// Code generated from schema/snapshot.v1.json by quicktype. DO NOT EDIT.

/**
 * Current v1 writer model. Readers preserve compatibility with earlier v1 documents by
 * applying zero defaults for omitted fields. Arrays may be null in Go JSON output.
 */
export interface Snapshot {
    format:       string;
    version:      number;
    generator:    Generator;
    locale:       string;
    collection:   Collection;
    environment:  Environment;
    environments: Environment[] | null;
    prose:        Prose;
    changes:      Release[] | null;
    options:      Options;
}

export interface Release {
    at:       string;
    items:    Item[] | null;
    revision: number;
}

export interface Item {
    sign:   string;
    method: string;
    name:   string;
    fields: string[] | null;
}

export interface Collection {
    id:          string;
    name:        string;
    description: string;
    groups:      Group[] | null;
    variables:   Variable[] | null;
}

export interface Group {
    id:          string;
    name:        string;
    description: string;
    endpoints:   Endpoint[] | null;
}

export interface Endpoint {
    id:          string;
    name:        string;
    method:      string;
    path:        string;
    description: string;
    parameters:  Parameter[] | null;
    body:        Body | null;
    responses:   Response[] | null;
}

export interface Body {
    fields:  Field[] | null;
    example: string;
}

export interface Field {
    name:        string;
    type:        string;
    required:    boolean;
    description: string;
}

export interface Parameter {
    name:        string;
    in:          string;
    type:        string;
    required:    boolean;
    description: string;
    example:     string;
}

export interface Response {
    status: number;
    name:   string;
    body:   string;
}

export interface Variable {
    key:   string;
    value: string;
}

export interface Environment {
    name:    string;
    baseUrl: string;
}

export interface Generator {
    name:    string;
    version: string;
}

export interface Options {
    docs:     boolean;
    examples: boolean;
    tryIt:    boolean;
}

export interface Prose {
    lead:   string;
    auth:   string;
    errors: string;
    limits: string;
}
