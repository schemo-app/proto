import { mkdir, readFile, writeFile } from 'node:fs/promises'
import { fileURLToPath } from 'node:url'
import { dirname, resolve } from 'node:path'
import { spawnSync } from 'node:child_process'
import { InputData, JSONSchemaInput, TypeScriptTargetLanguage, quicktype } from 'quicktype-core'

const root = fileURLToPath(new URL('..', import.meta.url))
const header = 'Code generated from schema/snapshot.v1.json by quicktype. DO NOT EDIT.'
const checkOnly = process.argv.includes('--check')

// An empty mapping keeps date-times as the strings JSON carries instead of turning them into Date.
class JSONTypeScript extends TypeScriptTargetLanguage {
    get stringTypeMapping() { return new Map() }
}

async function readSchema() {
    const schema = JSON.parse(await readFile(resolve(root, 'schema/snapshot.v1.json'), 'utf8'))
    if (typeof schema.properties.format.const !== 'string' || !Number.isSafeInteger(schema.properties.version.const)) {
        throw new Error('Snapshot format/version metadata is invalid')
    }
    // Field order is part of existing endpoint fingerprints, so quicktype keeps the declared order.
    for (const definition of [schema, ...Object.values(schema.definitions)]) {
        definition.quicktypePropertyOrder = Object.keys(definition.properties)
    }
    return schema
}

async function quicktypeInput(schema) {
    const model = structuredClone(schema)
    // Keep the published field types while the JSON Schema enforces the envelope literals.
    delete model.properties.format.const
    delete model.properties.version.const
    const input = new JSONSchemaInput(undefined)
    await input.addSource({ name: 'Snapshot', schema: JSON.stringify(model) })
    const inputData = new InputData()
    inputData.addInput(input)
    return inputData
}

function formatGo(generatedLines, schema) {
    const packageLine = generatedLines.findIndex(line => line.startsWith('package '))
    if (packageLine < 0) throw new Error('Generated Go package is missing')
    const { format, version } = schema.properties
    const constants = `\nconst Format = ${JSON.stringify(format.const)}\nconst Version = ${version.const}\n`
    const source = `// ${header}\n` + generatedLines.slice(packageLine).join('\n') + '\n' + constants
    const formatted = spawnSync('gofmt', { input: source, encoding: 'utf8' })
    if (formatted.error) throw formatted.error
    if (formatted.status !== 0) throw new Error(formatted.stderr || 'gofmt failed')
    return formatted.stdout
}

async function generate(schema) {
    const inputData = await quicktypeInput(schema)
    const [go, typescript] = await Promise.all([
        quicktype({ inputData, lang: 'go', rendererOptions: { package: 'snapshot' } }),
        quicktype({ inputData, lang: new JSONTypeScript(), leadingComments: [header],
            rendererOptions: { 'just-types': true } }),
    ])
    return new Map([
        [resolve(root, 'snapshot/v1/types.go'), formatGo(go.lines, schema)],
        [resolve(root, 'snapshot/v1/types.d.ts'), typescript.lines.join('\n').trimEnd() + '\n'],
    ])
}

async function emit(path, expected) {
    if (!checkOnly) {
        await mkdir(dirname(path), { recursive: true })
        await writeFile(path, expected)
        return
    }
    const actual = await readFile(path, 'utf8')
    if (actual.replaceAll('\r\n', '\n') !== expected) throw new Error(`Snapshot generation drift: ${path}`)
}

for (const [path, expected] of await generate(await readSchema())) {
    await emit(path, expected)
}
