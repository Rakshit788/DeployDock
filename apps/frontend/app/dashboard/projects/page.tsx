'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { apiClient } from '@/lib/api'
import { getStoredToken } from '@/lib/auth'
import { useRouter } from 'next/navigation'

interface Project {
  id: number
  name: string
  repo_url: string
  created_at: string
}

export default function Projects() {
  const router = useRouter()
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)
  const [repoUrl, setRepoUrl] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [user, setUser] = useState<any>(null)

  useEffect(() => {
    if (!getStoredToken()) {
      router.push('/auth')
      return
    }

    fetchData()
  }, [router])

  const fetchData = async () => {
    try {
      const userId = Number(localStorage.getItem('user_id') || '0')
      const [projectsData, userData] = await Promise.all([
        apiClient.listProjects(userId || undefined),
        apiClient.getUser().catch(() => null)
      ])
      setProjects(projectsData || [])
      setUser(userData)
    } catch (error) {
      console.error('Error fetching data:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleCreateProject = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!repoUrl) return

    setSubmitting(true)
    try {
      const userId = Number(localStorage.getItem('user_id') || '0')
      const result = await apiClient.createProject(repoUrl, undefined, userId || undefined)
      if (!result?.project_id) {
        throw new Error('Project creation response missing project_id')
      }
      await fetchData()
      setRepoUrl('')
      alert('Project created successfully!')
    } catch (error: any) {
      console.error('Error creating project:', error)
      alert(error.response?.data?.details || error.response?.data?.error || error.message || 'Failed to create project')
    } finally {
      setSubmitting(false)
    }
  }

  const handleDeploy = async (projectId: number) => {
    try {
      const deployment = await apiClient.createDeployment(projectId)
      alert(`Deployment started! ID: ${deployment.DEPLOYMENT_ID || deployment.deployment_id || 'pending'}`)
      // Refresh deployments
      router.push('/dashboard')
    } catch (error: any) {
      console.error('Error deploying:', error)
      alert(error.response?.data?.message || 'Failed to start deployment')
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
            <Link href="/dashboard/projects" className="text-blue-400">Projects</Link>
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
            <h2 className="text-3xl font-bold mb-2">Projects</h2>
            <p className="text-slate-400">Manage and deploy your projects</p>
          </div>

          {/* Create Project Form */}
          <div className="bg-slate-800 p-6 rounded-lg border border-slate-700">
            <h3 className="text-lg font-semibold mb-4">Create New Project</h3>
            <form onSubmit={handleCreateProject} className="flex flex-col sm:flex-row gap-2">
              <input
                type="url"
                placeholder="GitHub repo URL (e.g., https://github.com/user/repo)"
                value={repoUrl}
                onChange={(e) => setRepoUrl(e.target.value)}
                className="flex-1 bg-slate-900 border border-slate-700 rounded px-4 py-2 text-white placeholder-slate-500 focus:outline-none focus:border-blue-500"
                required
                disabled={submitting}
              />
              <button
                type="submit"
                className="bg-blue-600 hover:bg-blue-700 disabled:bg-slate-600 px-6 py-2 rounded font-medium transition"
                disabled={submitting}
              >
                {submitting ? 'Creating...' : 'Create'}
              </button>
            </form>
          </div>

          {/* Projects List */}
          <div>
            <h3 className="text-lg font-semibold mb-4">Your Projects</h3>
            {loading ? (
              <div className="text-slate-400">Loading...</div>
            ) : projects.length === 0 ? (
              <div className="bg-slate-800 p-8 rounded-lg border border-slate-700 text-center">
                <div className="text-slate-400 mb-4">No projects yet. Create one above.</div>
              </div>
            ) : (
              <div className="grid gap-4">
                {projects.map(project => (
                  <div key={project.id} className="bg-slate-800 p-6 rounded-lg border border-slate-700 hover:border-slate-600 transition">
                    <div className="flex justify-between items-start mb-4">
                      <div className="flex-1">
                        <h4 className="text-lg font-semibold mb-2">{project.name}</h4>
                        <a
                          href={project.repo_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-sm text-blue-400 hover:text-blue-300 break-all"
                        >
                          {project.repo_url}
                        </a>
                        <div className="text-xs text-slate-500 mt-2">
                          Created: {new Date(project.created_at).toLocaleString()}
                        </div>
                      </div>
                      <button
                        onClick={() => handleDeploy(project.id)}
                        className="bg-green-600 hover:bg-green-700 px-4 py-2 rounded font-medium transition whitespace-nowrap ml-4"
                      >
                        Deploy
                      </button>
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
