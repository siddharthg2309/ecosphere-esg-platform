import type { ApiError, AuthResponse, Department, DepartmentInput, PageResult, User } from './types'

const API_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

export class RequestError extends Error { constructor(public status:number, public body:ApiError){super(body.message)} }

async function request<T>(path:string, init:RequestInit={}) {
  const token = localStorage.getItem('ecosphere.accessToken')
  const response = await fetch(`${API_URL}${path}`, { ...init, headers:{'Content-Type':'application/json',...(token?{Authorization:`Bearer ${token}`}:{ }),...init.headers} })
  if (!response.ok) { const body = await response.json().catch(()=>({code:'request_failed',message:'Request failed'})) as ApiError; throw new RequestError(response.status,body) }
  return response.json() as Promise<T>
}

export const api = {
  auth:{ login:(email:string,password:string)=>request<AuthResponse>('/auth/login',{method:'POST',body:JSON.stringify({email,password})}), me:()=>request<User>('/me') },
  departments:{
    list:(limit=20,offset=0)=>request<PageResult<Department>>(`/departments?limit=${limit}&offset=${offset}`),
    create:(input:DepartmentInput)=>request<Department>('/departments',{method:'POST',body:JSON.stringify(input)}),
    update:(id:string,input:DepartmentInput)=>request<Department>(`/departments/${id}`,{method:'PUT',body:JSON.stringify(input)}),
    deactivate:(id:string)=>request<Department>(`/departments/${id}`,{method:'DELETE'}),
  },
}
