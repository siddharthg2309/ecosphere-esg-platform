import { create } from 'zustand'
import { api } from '../lib/apiClient'
import type { User } from '../lib/types'

interface AuthState { user?:User; token?:string; initialized:boolean; login(email:string,password:string):Promise<void>; restore():Promise<void>; logout():void }

export const useAuthStore = create<AuthState>((set)=>({
  token:localStorage.getItem('ecosphere.accessToken') ?? undefined, initialized:false,
  async login(email,password){ const result=await api.auth.login(email,password);localStorage.setItem('ecosphere.accessToken',result.accessToken);localStorage.setItem('ecosphere.refreshToken',result.refreshToken);set({token:result.accessToken,user:result.user,initialized:true}) },
  async restore(){ if(!localStorage.getItem('ecosphere.accessToken')){set({initialized:true});return} try{const user=await api.auth.me();set({user,initialized:true})}catch{localStorage.removeItem('ecosphere.accessToken');localStorage.removeItem('ecosphere.refreshToken');set({token:undefined,user:undefined,initialized:true})} },
  logout(){localStorage.removeItem('ecosphere.accessToken');localStorage.removeItem('ecosphere.refreshToken');set({token:undefined,user:undefined,initialized:true})},
}))
