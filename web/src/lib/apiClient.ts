import type { ApiError, AuthResponse, Department, DepartmentInput, ESGConfig, NotificationPreference, PageResult, User } from './types'

const API_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

export class RequestError extends Error { constructor(public status:number, public body:ApiError){super(body.message)} }

export async function request<T>(path:string, init:RequestInit={}) {
  const token = localStorage.getItem('ecosphere.accessToken')
  const response = await fetch(`${API_URL}${path}`, { ...init, headers:{'Content-Type':'application/json',...(token?{Authorization:`Bearer ${token}`}:{ }),...init.headers} })
  if (!response.ok) { const body = await response.json().catch(()=>({code:'request_failed',message:'Request failed'})) as ApiError; throw new RequestError(response.status,body) }
  if (response.status === 204) return undefined as T
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
  master:{
    list:<T>(entity:string,params='')=>request<PageResult<T>>(`/${entity}?limit=100&offset=0${params}`),
    create:<T>(entity:string,input:unknown)=>request<T>(`/${entity}`,{method:'POST',body:JSON.stringify(input)}),
    update:<T>(entity:string,id:string,input:unknown)=>request<T>(`/${entity}/${id}`,{method:'PUT',body:JSON.stringify(input)}),
    remove:(entity:string,id:string)=>request<void>(`/${entity}/${id}`,{method:'DELETE'}),
  },
  settings:{
    config:()=>request<ESGConfig>('/settings/esg-config'),
    saveConfig:(input:ESGConfig)=>request<ESGConfig>('/settings/esg-config',{method:'PUT',body:JSON.stringify(input)}),
    preferences:()=>request<NotificationPreference[]>('/settings/notification-preferences'),
    savePreferences:(input:NotificationPreference[])=>request<NotificationPreference[]>('/settings/notification-preferences',{method:'PUT',body:JSON.stringify(input)}),
  },
}
