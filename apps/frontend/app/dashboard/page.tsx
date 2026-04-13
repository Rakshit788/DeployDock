'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { apiClient } from '@/lib/api'
import { getStoredToken } from '@/lib/auth'
import { useRouter } from 'next/navigation'

interface Deployment {
  id: string
  project_id: string
  status: string
  url?: string
  created_at: string
}

export default function Dashboard() {
  const router = useRouter()
  const [deployments, setDeployments] = useState<Deployment[]>([])
  const [loading, setLoading] = useState(true)
  const [user, setUser] = useState<any>(null)

  useEffect(() => {
    // Check if authenticated
    if (!getStoredToken()) {
      router.push('/auth')
      return
    }

    fetchData()
  }, [router])

  const fetchData = async () => {
    try {
      const [deploymentsData, userData] = await Promise.all([
        apiClient.listDeployments(),
        apiClient.getUser().catch(() => null)
      ])
      setDeployments(deploymentsData || [])
      setUser(userData)
    } catch (error) {
      console.error('Error fetching data:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleLogout = () => {
    localStorage.removeItem('auth_token')
    router.push('/auth')
  }

  return (
    <div className="min-h-screen bg-slate-950">
      {/* Header */}
      <header className="bg-slate-900 border-b border-slate-800 sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-4 py-4 flex justify-between items-center">
          <Link href="/" className="text-2xl font-bold text-blue-400">DeployDoc</Link>
          <nav className="flex gap-6 items-center">
            <Link href="/dashboard" className="hover:text-blue-400 transition">Dashboard</Link>
            <Link href="/dashboard/projects" className="hover:text-blue-400 transition">Projects</Link>
            <div className="flex items-center gap-4 pl-6 border-l border-slate-700">
              {user && <span className="text-sm text-slate-400">{user.login}</span>}
              <button
                onClick={handleLogout}
                className="text-red-400 hover:text-red-300 transition text-sm"
              >
                Logout
              </button>
            </div>
          </nav>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 py-8">
        <div className="space-y-8">
          <div>
            <h2 className="text-3xl font-bold mb-2">Dashboard</h2>
            <p className="text-slate-400">Welcome back! Here's your deployment overview</p>
          </div>

          {/* Stats */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="bg-slate-800 p-6 rounded-lg border border-slate-700">
              <div className="text-slate-400 text-sm mb-2">Total Deployments</div>
              <div className="text-3xl font-bold">{deployments.length}</div>
            </div>
            <div className="bg-slate-800 p-6 rounded-lg border border-slate-700">
              <div className="text-slate-400 text-sm mb-2">Active</div>
              <div className="text-3xl font-bold text-green-400">
                {deployments.filter(d => d.status === 'deployed').length}
              </div>
            </div>
            <div className="bg-slate-800 p-6 rounded-lg border border-slate-700">
              <div className="text-slate-400 text-sm mb-2">Failed</div>
              <div className="text-3xl font-bold text-red-400">
                {deployments.filter(d => d.status === 'failed').length}
              </div>
            </div>
          </div>

          {/* Recent Deployments */}
          <div>
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-xl font-bold">Recent Deployments</h3>
              <Link href="/dashboard/projects" className="text-blue-400 hover:text-blue-300 transition">
                New Deployment →
              </Link>
            </div>
            {loading ? (
              <div className="text-slate-400">Loading...</div>
            ) : deployments.length === 0 ? (
              <div className="bg-slate-800 p-8 rounded-lg border border-slate-700 text-center">
                <div className="text-slate-400 mb-4">No deployments yet</div>
                <Link
                  href="/dashboard/projects"
                  className="bg-blue-600 hover:bg-blue-700 px-4 py-2 rounded inline-block transition"
                >
                  Create Your First Deployment
                </Link>
              </div>
            ) : (
              <div className="space-y-2">
                {deployments.slice(0, 10).map(deployment => (
                  <div key={deployment.id} className="bg-slate-800 p-4 rounded-lg border border-slate-700 flex justify-between items-center hover:border-slate-600 transition">
                    <div>
                      <div className="font-semibold">{deployment.project_id}</div>
                      <div className="text-sm text-slate-400">
                        {new Date(deployment.created_at).toLocaleString()}
                      </div>
                    </div>
                    <div className="flex items-center gap-4">
                      {deployment.url && (
                        <a
                          href={deployment.url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-blue-400 hover:text-blue-300 text-sm"
                        >
                          Visit →
                        </a>
                      )}
                      <span className={`px-3 py-1 rounded text-sm font-medium ${
                        deployment.status === 'deployed'
                          ? 'bg-green-500/20 text-green-400'
                          : deployment.status === 'failed'
                          ? 'bg-red-500/20 text-red-400'
                          : 'bg-yellow-500/20 text-yellow-400'
                      }`}>
                        {deployment.status}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </main>
    </div>
  )
}
