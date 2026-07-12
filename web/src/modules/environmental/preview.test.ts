import { describe,expect,it } from 'vitest'
import { co2Preview } from './useEnvironmentalVM'
describe('CO2 preview',()=>{it('updates deterministically when quantity changes',()=>{expect(co2Preview('268','2.68')).toBe('718.240');expect(co2Preview('300','2.68')).toBe('804.000')})})
