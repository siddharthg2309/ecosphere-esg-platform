import type { ESGConfig } from '../../lib/types'

export function weightsValid(value: Pick<ESGConfig, 'weightEnv' | 'weightSocial' | 'weightGov'>) {
  return value.weightEnv >= 0 && value.weightSocial >= 0 && value.weightGov >= 0 && value.weightEnv + value.weightSocial + value.weightGov === 100
}
