import assert from 'node:assert/strict'
import test from 'node:test'
import { formatDuration, formatInterval, intervalToPosition, minuteToPercent } from '../src/lib/timeline.ts'

test('maps and clamps timeline minutes', () => {
  assert.equal(minuteToPercent(0, 90), 0)
  assert.equal(minuteToPercent(90, 90), 100)
  assert.equal(minuteToPercent(-5, 90), 0)
  assert.equal(minuteToPercent(100, 90), 100)
  assert.equal(minuteToPercent(10, 0), 0)
})
test('positions the canonical half-open interval', () => assert.deepEqual(intervalToPosition(15, 60, 90), { left: '16.666667%', width: '50%' }))
test('formats durations and bounded or ongoing intervals', () => {
  assert.equal(formatDuration(1), '1 minute'); assert.equal(formatDuration(45), '45 minutes')
  assert.equal(formatInterval(15, 60), '[15,60)'); assert.equal(formatInterval(15, 90, true), '[15,90) — continues to simulation horizon')
})
