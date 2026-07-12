import { describe,expect,it } from 'vitest'
import { weightsValid } from './settingsValidation'
describe('weightsValid',()=>{it('rejects weights that do not total 100',()=>expect(weightsValid({weightEnv:40,weightSocial:30,weightGov:40})).toBe(false));it('accepts 40/30/30',()=>expect(weightsValid({weightEnv:40,weightSocial:30,weightGov:30})).toBe(true))})
