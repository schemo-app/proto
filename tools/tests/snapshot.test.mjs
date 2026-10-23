import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { test } from 'node:test'
import Ajv from 'ajv'
import addFormats from 'ajv-formats'

const schema = JSON.parse(await readFile(new URL('../../schema/snapshot.v1.json', import.meta.url), 'utf8'))
const fixture = JSON.parse(await readFile(new URL('../../schema/fixtures/v1.json', import.meta.url), 'utf8'))
const validator = new Ajv({ strict: true })
validator.addKeyword({ keyword: 'x-format', schemaType: 'string' })
validator.addKeyword({ keyword: 'x-version', schemaType: 'number' })
addFormats(validator)
const validate = validator.compile(schema)

test('current writer fixture conforms to the versioned schema, including nullable lists', () => {
    assert.equal(validate(fixture), true, JSON.stringify(validate.errors))
    assert.equal(fixture.format, schema['x-format'])
    assert.equal(fixture.version, schema['x-version'])
    assert.equal(fixture.environments, null)
})

test('schema rejects malformed structures and accidental secret fields in writer documents', () => {
    for (const mutate of [
        document => { document.collection.groups = 'not a list' },
        document => { document.collection.groups[0].endpoints[0].body.fields[0].required = 'true' },
        document => { document.changes[0].at = 'not a timestamp' },
        document => { document.collection.variables[0].secret = 'must not be copied' },
    ]) {
        const changed = structuredClone(fixture)
        mutate(changed)
        assert.equal(validate(changed), false)
    }
})
