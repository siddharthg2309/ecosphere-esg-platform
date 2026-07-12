import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { api } from '../../lib/apiClient'
import { queryKeys } from '../../lib/queryKeys'
import type { DepartmentInput } from '../../lib/types'

export function useDepartmentsVM(){const[page,setPage]=useState(0);const queryClient=useQueryClient();const query=useQuery({queryKey:[...queryKeys.departments,page],queryFn:()=>api.departments.list(20,page*20)});const refresh=()=>queryClient.invalidateQueries({queryKey:queryKeys.departments});const create=useMutation({mutationFn:(input:DepartmentInput)=>api.departments.create(input),onSuccess:refresh});const update=useMutation({mutationFn:({id,input}:{id:string;input:DepartmentInput})=>api.departments.update(id,input),onSuccess:refresh});const deactivate=useMutation({mutationFn:api.departments.deactivate,onSuccess:refresh});return{rows:query.data?.items??[],total:query.data?.total??0,page,setPage,loading:query.isLoading,error:query.error,createDepartment:create.mutateAsync,updateDepartment:(id:string,input:DepartmentInput)=>update.mutateAsync({id,input}),saving:create.isPending||update.isPending,deactivateDepartment:deactivate.mutateAsync}}
