import { describe, it, expect } from 'vitest'
import {
  localTimeToUTC,
  utcTimeToLocal,
  requestToUTC,
  responseToLocal,
} from '../timezone'

describe('localTimeToUTC', () => {
  it('should convert local time to UTC time string', () => {
    const result = localTimeToUTC('12:00')
    expect(result).toMatch(/^\d{2}:\d{2}$/)
  })

  it('should handle midnight', () => {
    const result = localTimeToUTC('00:00')
    expect(result).toMatch(/^\d{2}:\d{2}$/)
  })

  it('should handle edge time 23:59', () => {
    const result = localTimeToUTC('23:59')
    expect(result).toMatch(/^\d{2}:\d{2}$/)
  })
})

describe('utcTimeToLocal', () => {
  it('should convert UTC time to local time string', () => {
    const result = utcTimeToLocal('12:00')
    expect(result).toMatch(/^\d{2}:\d{2}$/)
  })

  it('should round-trip through midnight correctly', () => {
    const original = '00:00'
    const utc = localTimeToUTC(original)
    const local = utcTimeToLocal(utc)
    // The minutes should be preserved through the round trip
    const [, utcMin] = utc.split(':').map(Number)
    const [, localMin] = local.split(':').map(Number)
    expect(utcMin).toBe(localMin)
  })
})

describe('requestToUTC', () => {
  it('should return null as-is', () => {
    expect(requestToUTC(null)).toBeNull()
  })

  it('should return primitives as-is', () => {
    expect(requestToUTC('hello')).toBe('hello')
    expect(requestToUTC(42)).toBe(42)
  })

  it('should return arrays with each element converted', () => {
    const result = requestToUTC([{ name: 'test' }, { name: 'test2' }])
    expect(result).toEqual([{ name: 'test' }, { name: 'test2' }])
  })

  it('should convert time in schedule_data from local to UTC', () => {
    const input = {
      task: {
        schedule_data: { time: '09:00', days: [1, 3, 5] },
      },
    }
    const result = requestToUTC(input) as any
    const sd = result.task.schedule_data
    expect(sd.time).toMatch(/^\d{2}:\d{2}$/)
    expect(sd.days).toEqual([1, 3, 5])
  })

  it('should ignore schedule_data without time field', () => {
    const input = { schedule_data: { days: [1] } }
    const result = requestToUTC(input) as any
    expect(result.schedule_data.days).toEqual([1])
    expect(result.schedule_data.time).toBeUndefined()
  })

  it('should convert date+time for one-shot schedule_data', () => {
    const input = {
      schedule_data: { date: '2026-08-15', time: '10:00' },
    }
    const result = requestToUTC(input) as any
    expect(result.schedule_data.date).toMatch(/^\d{4}-\d{2}-\d{2}$/)
    expect(result.schedule_data.time).toMatch(/^\d{2}:\d{2}$/)
  })
})

describe('responseToLocal', () => {
  it('should return null as-is', () => {
    expect(responseToLocal(null)).toBeNull()
  })

  it('should return primitives as-is', () => {
    expect(responseToLocal('hello')).toBe('hello')
    expect(responseToLocal(123)).toBe(123)
  })

  it('should convert time in schedule_data from UTC to local', () => {
    const input = {
      schedule_data: { time: '01:00', days: [2, 4] },
    }
    const result = responseToLocal(input) as any
    expect(result.schedule_data.time).toMatch(/^\d{2}:\d{2}$/)
    expect(result.schedule_data.days).toEqual([2, 4])
  })

  it('should handle nested objects with multiple schedule_data', () => {
    const input = [
      { schedule_data: { time: '01:00' } },
      { schedule_data: { time: '02:00' } },
    ]
    const result = responseToLocal(input) as any
    expect(result[0].schedule_data.time).toMatch(/^\d{2}:\d{2}$/)
    expect(result[1].schedule_data.time).toMatch(/^\d{2}:\d{2}$/)
  })
})
