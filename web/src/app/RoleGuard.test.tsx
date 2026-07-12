import { render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it } from 'vitest'
import { RoleGuard } from './RoleGuard'
import { useAuthStore } from './authStore'

describe('RoleGuard',()=>{beforeEach(()=>useAuthStore.setState({user:undefined,initialized:true}));it('renders for an allowed role',()=>{useAuthStore.setState({user:{id:'1',name:'Admin',email:'a@b.com',role:'admin'}});render(<RoleGuard roles={['admin']}><button>Create</button></RoleGuard>);expect(screen.getByRole('button',{name:'Create'})).toBeInTheDocument()});it('blocks a disallowed role',()=>{useAuthStore.setState({user:{id:'2',name:'Employee',email:'e@b.com',role:'employee'}});render(<RoleGuard roles={['admin']}><button>Create</button></RoleGuard>);expect(screen.queryByRole('button')).not.toBeInTheDocument()})})
