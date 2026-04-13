export const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
export const GITHUB_CLIENT_ID = process.env.NEXT_PUBLIC_GITHUB_CLIENT_ID || ''
export const GITHUB_REDIRECT_URI = `${typeof window !== 'undefined' ? window.location.origin : ''}/auth/callback`

export const getGitHubAuthUrl = () => {
  const params = new URLSearchParams({
    client_id: GITHUB_CLIENT_ID,
    redirect_uri: GITHUB_REDIRECT_URI,
    scope: 'user:email repository',
    state: Math.random().toString(36).substring(7),
  })
  return `https://github.com/login/oauth/authorize?${params.toString()}`
}

export const getStoredToken = () => {
  if (typeof window === 'undefined') return null
  return localStorage.getItem('auth_token')
}

export const saveToken = (token: string) => {
  if (typeof window === 'undefined') return
  localStorage.setItem('auth_token', token)
}

export const clearToken = () => {
  if (typeof window === 'undefined') return
  localStorage.removeItem('auth_token')
}
