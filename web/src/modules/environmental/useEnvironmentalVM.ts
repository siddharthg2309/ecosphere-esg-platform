import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useAuthStore } from '../../app/authStore'
import { api } from '../../lib/apiClient'
import { queryKeys } from '../../lib/queryKeys'
import type { CarbonSuggestion, CarbonTransactionInput, EmissionFactor, EnvironmentalGoalInput } from '../../lib/types'

export type IngestState = 'idle' | 'uploading' | 'suggested' | 'submitting' | 'draftCreated'

export function co2Preview(quantity: string | number, factor: string | number) {
  const value = Number(quantity) * Number(factor)
  return Number.isFinite(value) ? value.toFixed(3) : '0.000'
}

export function useEnvironmentalVM() {
  const user = useAuthStore((s) => s.user)
  const qc = useQueryClient()
  const [state, setState] = useState<IngestState>('idle')
  const [suggestion, setSuggestion] = useState<CarbonSuggestion>()

  const transactions = useQuery({ queryKey: queryKeys.carbon, queryFn: () => api.carbon.list() })
  const goals = useQuery({ queryKey: queryKeys.goals, queryFn: () => api.goals.list() })
  const summary = useQuery({ queryKey: queryKeys.carbonSummary, queryFn: () => api.carbon.summary() })
  const factors = useQuery({
    queryKey: ['master', 'emission-factors', 'environmental'],
    queryFn: () => api.master.list<EmissionFactor>('emission-factors'),
  })
  const departments = useQuery({
    queryKey: [...queryKeys.departments, 'environmental'],
    queryFn: () => api.departments.list(100, 0),
  })

  const refresh = async () =>
    Promise.all([
      qc.invalidateQueries({ queryKey: queryKeys.carbon }),
      qc.invalidateQueries({ queryKey: queryKeys.goals }),
      qc.invalidateQueries({ queryKey: queryKeys.carbonSummary }),
    ])

  const ingest = useMutation({
    mutationFn: api.carbon.ingest,
    onMutate: () => setState('uploading'),
    onSuccess: (value) => {
      setSuggestion(value)
      setState('suggested')
    },
    onError: () => setState('idle'),
  })

  const record = useMutation({
    mutationFn: (input: CarbonTransactionInput) => api.carbon.create(input),
    onMutate: () => setState('submitting'),
    onSuccess: async () => {
      setState('draftCreated')
      await refresh()
    },
    onError: () => setState('suggested'),
  })

  const verify = useMutation({
    mutationFn: api.carbon.verify,
    onSuccess: () => void refresh(),
  })

  const createGoal = useMutation({
    mutationFn: (input: EnvironmentalGoalInput) => api.goals.create(input),
    onSuccess: () => void qc.invalidateQueries({ queryKey: queryKeys.goals }),
  })

  return {
    user,
    state,
    suggestion,
    transactions: transactions.data?.items ?? [],
    goals: goals.data?.items ?? [],
    summary: summary.data,
    factors: factors.data?.items.filter((v) => v.status === 'active') ?? [],
    departments: departments.data?.items ?? [],
    loading: transactions.isLoading || goals.isLoading,
    ingestFile: ingest.mutateAsync,
    createDraft: record.mutateAsync,
    verify: verify.mutateAsync,
    createGoal: createGoal.mutateAsync,
    error: ingest.error ?? record.error ?? verify.error ?? createGoal.error,
    resetIngest: () => {
      setState('idle')
      setSuggestion(undefined)
    },
  }
}
