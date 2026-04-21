'use client'

import { useEffect, useState } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { apiClient } from '@/lib/api'
import { saveToken, saveGitHubToken } from '@/lib/auth'

export default function CallbackContent() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const handleCallback = async () => {
      try {
        const code = searchParams.get('code')
        const state = searchParams.get('state')

        if (!code) {
          setError('No authorization code received')
          setLoading(false)
          return
        }

        // Exchange code for token with backend
        const response = await apiClient.githubAuth(code)

        if (response) {
          // Save authentication token (JWT or session token from API)
          if (response.token) {
            saveToken(response.token)
          }
          // Save GitHub access token if provided
          if (response.github_token) {
            saveGitHubToken(response.github_token)
          }
          // Also save user info for quick access
          if (response.user_id) {
            localStorage.setItem('user_id', String(response.user_id))
            localStorage.setItem('username', response.username || '')
          }
          // Redirect to dashboard
          router.push('/dashboard')
        } else {
          setError('Failed to authenticate')
        }
      } catch (err: any) {
        setError(err.response?.data?.message || 'Authentication failed')
        console.error('Auth error:', err)
      } finally {
        setLoading(false)
      }
    }

    handleCallback()
  }, [searchParams, router])

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-950 to-blue-900 flex items-center justify-center">
        <div className="text-center space-y-4">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-blue-400 mx-auto"></div>
          <p className="text-slate-300">Authenticating with GitHub...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-950 to-blue-900 flex items-center justify-center px-4">
        <div className="bg-slate-800 border border-slate-700 rounded-lg p-8 max-w-md text-center space-y-4">
          <div className="text-red-400 text-lg font-semibold">Authentication Error</div>
          <p className="text-slate-300">{error}</p>
          <button
            onClick={() => window.location.href = '/auth'}
            className="bg-blue-600 hover:bg-blue-700 px-6 py-2 rounded font-semibold transition"
          >
            Try Again
          </button>
        </div>
      </div>
    )
  }

  return null
}
