import assert from 'node:assert/strict'
import { test } from 'node:test'
import { create, fromJson, toBinary, fromBinary, getOption } from '@bufbuild/protobuf'
import { PlatformService, GetInfoResponseSchema } from '@schemo-app/proto/rpc/platform/v1/platform_pb.js'
import { ErrorInfoSchema } from '@schemo-app/proto/rpc/google/rpc/error_details_pb.js'
import { Access, access } from '@schemo-app/proto/rpc/common/v1/access_pb.js'

test('public package exports usable service descriptors and binary messages', () => {
    assert.equal(PlatformService.typeName, 'platform.v1.PlatformService')
    const message = fromJson(GetInfoResponseSchema, { server: { name: 'Example', version: '1' } })
    const received = fromBinary(GetInfoResponseSchema, toBinary(GetInfoResponseSchema, message))
    assert.equal(received.server.name, 'Example')
})

test('standard error details use the public protobuf runtime', () => {
    const details = create(ErrorInfoSchema, { reason: 'denied', domain: 'example.test' })
    assert.equal(fromBinary(ErrorInfoSchema, toBinary(ErrorInfoSchema, details)).reason, 'denied')
})

test('method access options are available through the public descriptors', () => {
    assert.equal(getOption(PlatformService.method.getInfo, access), Access.PUBLIC)
})
