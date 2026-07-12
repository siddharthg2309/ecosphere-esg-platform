import { useEffect } from 'react'
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { DepartmentsPage } from '../modules/settings/DepartmentsPage'
import { AppShell } from './AppShell'
import { LoginPage } from './LoginPage'
import { Providers } from './Providers'
import { useAuthStore } from './authStore'

function ProtectedApp(){const user=useAuthStore(s=>s.user);const initialized=useAuthStore(s=>s.initialized);const restore=useAuthStore(s=>s.restore);useEffect(()=>{void restore()},[restore]);if(!initialized)return <div className="center-state">Loading EcoSphere…</div>;if(!user)return <Navigate to="/login" replace/>;return <Routes><Route element={<AppShell/>}><Route index element={<Navigate to="/settings" replace/>}/><Route path="settings" element={<DepartmentsPage/>}/><Route path="*" element={<div className="content-card"><h1>Coming in the next phase</h1></div>}/></Route></Routes>}
export function App(){return <Providers><BrowserRouter><Routes><Route path="/login" element={<LoginPage/>}/><Route path="/*" element={<ProtectedApp/>}/></Routes></BrowserRouter></Providers>}
