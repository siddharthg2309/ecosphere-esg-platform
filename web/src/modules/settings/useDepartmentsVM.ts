import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { api } from '../../lib/apiClient'
import { queryKeys } from '../../lib/queryKeys'
import type { DepartmentInput } from '../../lib/types'

export function useDepartmentsVM(){const[page,setPage]=useState(0);const queryClient=useQueryClient();const query=useQuery({queryKey:[...queryKeys.departments,page],queryFn:()=>api.departments.list(20,page*20)});const create=useMutation({mutationFn:(input:DepartmentInput)=>api.departments.create(input),onSuccess:()=>queryClient.invalidateQueries({queryKey:queryKeys.departments})});const deactivate=useMutation({mutationFn:api.departments.deactivate,onSuccess:()=>queryClient.invalidateQueries({queryKey:queryKeys.departments})});return{rows:query.data?.items??[],total:query.data?.total??0,page,setPage,loading:query.isLoading,error:query.error,createDepartment:create.mutateAsync,creating:create.isPending,deactivateDepartment:deactivate.mutateAsync}}
