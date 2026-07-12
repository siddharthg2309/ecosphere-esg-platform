import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { FoundationApp } from './FoundationApp'

describe('FoundationApp', () => {
  it('renders the EcoSphere product name', () => {
    render(<FoundationApp />)
    expect(screen.getByRole('heading', { name: 'EcoSphere' })).toBeInTheDocument()
  })
})
