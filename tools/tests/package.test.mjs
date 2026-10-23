import assert from 'node:assert/strict'
import { mkdtemp, mkdir, writeFile, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { spawnSync } from 'node:child_process'
import { test } from 'node:test'

const consumerCheck = `
    import { PlatformService } from '@schemo-app/proto/rpc/platform/v1/platform_pb.js'
    import { ErrorInfoSchema } from '@schemo-app/proto/rpc/google/rpc/error_details_pb.js'
    if (PlatformService.typeName !== 'platform.v1.PlatformService' || ErrorInfoSchema.typeName !== 'google.rpc.ErrorInfo') process.exit(1)
`

test('packed JavaScript contracts work without a producer checkout or generators', async () => {
    const root = fileURLToPath(new URL('../..', import.meta.url))
    const workDir = await mkdtemp(join(tmpdir(), 'schemo-proto-'))
    function npm(args, cwd) {
        const result = spawnSync(process.execPath, [process.env.npm_execpath, ...args],
            { cwd, encoding: 'utf8', timeout: 120000 })
        assert.equal(result.status, 0, result.stderr || result.stdout)
        return result.stdout
    }
    try {
        const packed = JSON.parse(npm(['pack', '--json', '--ignore-scripts', '--pack-destination', workDir], root))[0]
        const packedPaths = packed.files.map(file => file.path)
        assert.ok(packedPaths.includes('LICENSE'))
        assert.ok(packedPaths.includes('LICENSES/Apache-2.0.txt'))
        assert.ok(packedPaths.includes('snapshot/v1/types.d.ts'))
        assert.equal(packedPaths.some(path => /^(tools|go|node_modules)\//.test(path) || path.endsWith('.go')), false)

        const consumer = join(workDir, 'consumer')
        await mkdir(consumer)
        await writeFile(join(consumer, 'package.json'), '{"private":true,"type":"module"}')
        npm(['install', join(workDir, packed.filename), '--ignore-scripts', '--no-audit', '--no-fund'], consumer)
        const result = spawnSync(process.execPath, ['--input-type=module', '-e', consumerCheck],
            { cwd: consumer, encoding: 'utf8', timeout: 30000 })
        assert.equal(result.status, 0, result.stderr)
    } finally {
        await rm(workDir, { recursive: true, force: true })
    }
})
