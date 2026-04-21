'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { apiClient } from '@/lib/api'
import { getStoredToken, clearToken } from '@/lib/auth'
import { useRouter } from 'next/navigation'

interface GitHubRepo {
  id: number
  name: string
  full_name: string
  html_url: string
  description: string
  language: string
}

interface Project {
  id: number
  name: string
  repo_url: string
  subdomain: string
  created_at: string
}

interface Deployment {
  id: number
  project_id: number
  status: string
  url?: string
  created_at: string
}

export default function Dashboard() {
  const router = useRouter()
  const [repos, setRepos] = useState<GitHubRepo[]>([])
  const [projects, setProjects] = useState<Project[]>([])
  const [deployments, setDeployments] = useState<Deployment[]>([])
  const [loadingRepos, setLoadingRepos] = useState(true)
  const [loadingProjects, setLoadingProjects] = useState(true)
  const [creatingProject, setCreatingProject] = useState<string | null>(null)
  const [deployingProject, setDeployingProject] = useState<number | null>(null)
  const [username, setUsername] = useState('')
  const [tab, setTab] = useState<'overview' | 'repos' | 'projects'>('overview')

  useEffect(() => {
    if (!getStoredToken()) {
      router.push('/auth')
      return
    }

    const storedUsername = localStorage.getItem('username')
    if (storedUsername) setUsername(storedUsername)

    fetchData()
  }, [router])

  const fetchData = async () => {
    try {
      // Fetch repos from GitHub
      fetchRepos()
      // Fetch projects from API
      fetchProjects()
    } catch (error) {
      console.error('Error fetching data:', error)
    }
  }

  const fetchRepos = async () => {
    setLoadingRepos(true)
    try {
      const reposData = await apiClient.getGitHubRepos()
      setRepos(reposData)
    } catch (error) {
      console.error('Error fetching repos:', error)
      setRepos([])
    } finally {
      setLoadingRepos(false)
    }
  }

  const fetchProjects = async () => {
    setLoadingProjects(true)
    try {
      const userId = Number(localStorage.getItem('user_id') || '0')
      const projectsData = await apiClient.listProjects(userId || undefined)
      setProjects(projectsData || [])

      // Also fetch deployments
      const deploymentsData = await apiClient.listDeployments()
      setDeployments(deploymentsData || [])
    } catch (error) {
      console.error('Error fetching projects:', error)
    } finally {
      setLoadingProjects(false)
    }
  }

  const handleCreateProject = async (repo: GitHubRepo) => {
    setCreatingProject(repo.name)
    try {
      const userId = Number(localStorage.getItem('user_id') || '0')
      const result = await apiClient.createProject(repo.html_url, repo.name, userId || undefined)
      if (result?.project_id) {
        alert('Project created successfully!')
        await fetchProjects()
      } else {
        throw new Error('Project creation response missing project_id')
      }
    } catch (error: any) {
      console.error('Error creating project:', error)
      alert(error.response?.data?.details || error.response?.data?.error || error.message || 'Failed to create project')
    } finally {
      setCreatingProject(null)
    }
  }

  const handleCreateDeployment = async (projectId: number) => {
    setDeployingProject(projectId)
    try {
      const result = await apiClient.createDeployment(projectId)
      if (result) {
        alert('Deployment initiated!')
        fetchProjects()
      }
    } catch (error: any) {
      console.error('Error creating deployment:', error)
      alert(error.response?.data?.message || 'Failed to create deployment')
    } finally {
      setDeployingProject(null)
    }
  }

  const handleLogout = () => {
    clearToken()
    router.push('/auth')
  }

  const getProjectDeployments = (projectId: number) => {
    return deployments.filter(d => d.project_id === projectId)
  }

  return (
    <div className="min-h-screen bg-slate-950">
      {/* Header */}
      <header className="bg-slate-900 border-b border-slate-800 sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-4 py-4 flex justify-between items-center">
          <Link href="/" className="text-2xl font-bold text-blue-400">
            DeployHub
          </Link>
          <div className="flex items-center gap-4">
            <span className="text-sm text-slate-400">{username}</span>
            <button
              onClick={handleLogout}
              className="text-red-400 hover:text-red-300 transition text-sm"
            >
              Logout
            </button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 py-8">
        <div className="space-y-8">
          <div>
            <h2 className="text-3xl font-bold mb-2">Dashboard</h2>
            <p className="text-slate-400">Manage your GitHub repos and deployments</p>
          </div>

          {/* Tabs */}
          <div className="flex gap-4 border-b border-slate-700">
            {(['overview', 'repos', 'projects'] as const).map(t => (
              <button
                key={t}
                onClick={() => setTab(t)}
                className={`px-4 py-2 font-medium transition capitalize ${
                  tab === t
                    ? 'text-blue-400 border-b-2 border-blue-400 -mb-[2px]'
                    : 'text-slate-400 hover:text-slate-300'
                }`}
              >
                {t}
              </button>
            ))}
          </div>

          {/* Overview Tab */}
          {tab === 'overview' && (
            <div className="space-y-8">
              {/* Stats */}
              <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                <div className="bg-slate-800 p-6 rounded-lg border border-slate-700">
                  <div className="text-slate-400 text-sm mb-2">GitHub Repos</div>
                  <div className="text-3xl font-bold text-blue-400">{repos.length}</div>
                </div>
                <div className="bg-slate-800 p-6 rounded-lg border border-slate-700">
                  <div className="text-slate-400 text-sm mb-2">Projects</div>
                  <div className="text-3xl font-bold text-green-400">{projects.length}</div>
                </div>
                <div className="bg-slate-800 p-6 rounded-lg border border-slate-700">
                  <div className="text-slate-400 text-sm mb-2">Deployments</div>
                  <div className="text-3xl font-bold text-purple-400">{deployments.length}</div>
                </div>
                <div className="bg-slate-800 p-6 rounded-lg border border-slate-700">
                  <div className="text-slate-400 text-sm mb-2">Active</div>
                  <div className="text-3xl font-bold text-yellow-400">
                    {deployments.filter(d => d.status === 'deployed').length}
                  </div>
                </div>
              </div>

              {/* Recent Deployments */}
              <div>
                <h3 className="text-xl font-bold mb-4">Recent Deployments</h3>
                {loadingProjects ? (
                  <div className="text-slate-400">Loading...</div>
                ) : deployments.length === 0 ? (
                  <div className="bg-slate-800 p-8 rounded-lg border border-slate-700 text-center">
                    <div className="text-slate-400 mb-4">No deployments yet</div>
                    <button
                      onClick={() => setTab('repos')}
                      className="bg-blue-600 hover:bg-blue-700 px-4 py-2 rounded inline-block transition"
                    >
                      Create First Deployment
                    </button>
                  </div>
                ) : (
                  <div className="space-y-2">
                    {deployments.slice(0, 5).map(deployment => {
                      const project = projects.find(p => p.id === deployment.project_id)
                      return (
                        <div
                          key={deployment.id}
                          className="bg-slate-800 p-4 rounded-lg border border-slate-700 flex justify-between items-center"
                        >
                          <div>
                            <div className="font-semibold">{project?.name || deployment.project_id}</div>
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
                            <span
                              className={`px-3 py-1 rounded text-sm font-medium ${
                                deployment.status === 'deployed'
                                  ? 'bg-green-500/20 text-green-400'
                                  : deployment.status === 'failed'
                                  ? 'bg-red-500/20 text-red-400'
                                  : 'bg-yellow-500/20 text-yellow-400'
                              }`}
                            >
                              {deployment.status}
                            </span>
                          </div>
                        </div>
                      )
                    })}
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Repos Tab */}
          {tab === 'repos' && (
            <div>
              <div className="flex justify-between items-center mb-4">
                <h3 className="text-xl font-bold">Your GitHub Repositories</h3>
                <button
                  onClick={fetchRepos}
                  className="text-sm text-blue-400 hover:text-blue-300"
                >
                  Refresh
                </button>
              </div>

              {loadingRepos ? (
                <div className="text-slate-400">Loading repositories...</div>
              ) : repos.length === 0 ? (
                <div className="bg-slate-800 p-8 rounded-lg border border-slate-700 text-center">
                  <div className="text-slate-400">No public repositories found</div>
                </div>
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  {repos.map(repo => {
                    const existingProject = projects.find(p => p.repo_url === repo.html_url)
                    return (
                      <div
                        key={repo.id}
                        className="bg-slate-800 p-6 rounded-lg border border-slate-700 hover:border-slate-600 transition"
                      >
                        <div className="flex justify-between items-start mb-3">
                          <div className="flex-1">
                            <a
                              href={repo.html_url}
                              target="_blank"
                              rel="noopener noreferrer"
                              className="text-lg font-semibold text-blue-400 hover:text-blue-300"
                            >
                              {repo.name}
                            </a>
                            {repo.language && (
                              <div className="text-sm text-slate-400 mt-1">{repo.language}</div>
                            )}
                          </div>
                        </div>
                        {repo.description && (
                          <p className="text-sm text-slate-400 mb-4 line-clamp-2">
                            {repo.description}
                          </p>
                        )}
                        <button
                          onClick={() => handleCreateProject(repo)}
                          disabled={creatingProject === repo.name || !!existingProject}
                          className={`w-full py-2 rounded font-medium transition ${
                            existingProject
                              ? 'bg-slate-700 text-slate-500 cursor-not-allowed'
                              : creatingProject === repo.name
                              ? 'bg-blue-600 text-white'
                              : 'bg-blue-600 hover:bg-blue-700 text-white'
                          }`}
                        >
                          {creatingProject === repo.name
                            ? 'Creating...'
                            : existingProject
                            ? 'Already Created'
                            : 'Create Project'}
                        </button>
                      </div>
                    )
                  })}
                </div>
              )}
            </div>
          )}

          {/* Projects Tab */}
          {tab === 'projects' && (
            <div>
              <h3 className="text-xl font-bold mb-4">Your Projects</h3>

              {loadingProjects ? (
                <div className="text-slate-400">Loading projects...</div>
              ) : projects.length === 0 ? (
                <div className="bg-slate-800 p-8 rounded-lg border border-slate-700 text-center">
                  <div className="text-slate-400 mb-4">No projects yet</div>
                  <button
                    onClick={() => setTab('repos')}
                    className="bg-blue-600 hover:bg-blue-700 px-4 py-2 rounded inline-block transition"
                  >
                    Create Project from Repository
                  </button>
                </div>
              ) : (
                <div className="space-y-4">
                  {projects.map(project => {
                    const projectDeployments = getProjectDeployments(project.id)
                    return (
                      <div
                        key={project.id}
                        className="bg-slate-800 p-6 rounded-lg border border-slate-700"
                      >
                        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-4">
                          <div>
                            <div className="text-sm text-slate-400">Project</div>
                            <div className="font-semibold">{project.name}</div>
                          </div>
                          <div>
                            <div className="text-sm text-slate-400">Subdomain</div>
                            <div className="font-semibold text-blue-400">{project.subdomain}</div>
                          </div>
                          <div>
                            <div className="text-sm text-slate-400">Deployments</div>
                            <div className="font-semibold">{projectDeployments.length}</div>
                          </div>
                          <div>
                            <div className="text-sm text-slate-400">Status</div>
                            <div className="font-semibold">
                              {projectDeployments.length > 0
                                ? projectDeployments[0].status
                                : 'No deployment'}
                            </div>
                          </div>
                        </div>
                        <div className="flex gap-2">
                          <button
                            onClick={() => handleCreateDeployment(project.id)}
                            disabled={deployingProject === project.id}
                            className="bg-green-600 hover:bg-green-700 disabled:bg-slate-600 px-4 py-2 rounded transition font-medium text-sm"
                          >
                            {deployingProject === project.id ? 'Deploying...' : 'Deploy'}
                          </button>
                          <a
                            href={`https://${project.subdomain}.vercelclone.local`}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="bg-blue-600 hover:bg-blue-700 px-4 py-2 rounded transition font-medium text-sm"
                          >
                            Visit
                          </a>
                        </div>
                        {projectDeployments.length > 0 && (
                          <div className="mt-4 space-y-2">
                            <div className="text-sm text-slate-400">Recent Deployments:</div>
                            {projectDeployments.slice(0, 3).map(dep => (
                              <div key={dep.id} className="text-sm bg-slate-700 p-2 rounded">
                                <span className={`inline-block px-2 py-1 rounded text-xs font-medium mr-2 ${
                                  dep.status === 'deployed'
                                    ? 'bg-green-500/20 text-green-400'
                                    : dep.status === 'failed'
                                    ? 'bg-red-500/20 text-red-400'
                                    : 'bg-yellow-500/20 text-yellow-400'
                                }`}>
                                  {dep.status}
                                </span>
                                {new Date(dep.created_at).toLocaleString()}
                              </div>
                            ))}
                          </div>
                        )}
                      </div>
                    )
                  })}
                </div>
              )}
            </div>
          )}
        </div>
      </main>
    </div>
  )
}
