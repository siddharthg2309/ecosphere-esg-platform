export type Role = 'employee' | 'dept_head' | 'auditor' | 'admin'
export type Status = 'active' | 'inactive'

export interface User { id:string; name:string; email:string; role:Role; departmentId?:string }
export interface AuthResponse { accessToken:string; refreshToken:string; user:User }
export interface PageResult<T> { items:T[]; total:number }
export interface ApiError { code:string; message:string; fields?:Record<string,string> }
export interface Department { id:string; name:string; code:string; headId?:string; parentId?:string; employeeCount:number; status:Status; createdAt:string; updatedAt:string }
export type DepartmentInput = Pick<Department,'name'|'code'|'employeeCount'|'status'> & {headId?:string;parentId?:string}
