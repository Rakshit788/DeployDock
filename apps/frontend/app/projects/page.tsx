'use client'

import { useState, useEffect } from 'react'
import axios from 'axios'
import Link from 'next/link'

interface Project {
  id: string
  name: string
  repo_url: string
  created_at: string
}

export default function Projects() {
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)
  const [repoUrl, setRepoUrl] = useState('')
  const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

  useEffect(() => {
    fetchProjects()
  }, [])

  const fetchProjects = async () => {
    try {
      const response = await axios.get(`${apiUrl}/api/project/list`)
      setProjects(response.data || [])
    } catch (error) {
      console.error('Error fetching projects:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleCreateProject = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!repoUrl) return

    try {
      const response = await axios.post(`${apiUrl}/api/project/create`, {
        repo_url: repoUrl,
      })
      setProjects([response.data, ...projects])
      setRepoUrl('')
    } catch (error) {
      console.error('Error creating project:', error)
      alert('Failed to create project')
    }
  }

  const handleDeploy = async (projectId: string) => {
    try {
      await axios.post(`${apiUrl}/api/deployment/create`, {
        project_id: projectId,
      })
      alert('Deployment started!')
    } catch (error) {
      console.error('Error deploying:', error)
      alert('Failed to start deployment')
    }
  }

  return (
    <div className="space-y-8">
      <div>
        <h2 className="text-3xl font-bold mb-2">Projects</h2>
        <p className="text-slate-400">Manage and deploy your projects</p>
      </div>

      <div className="bg-slate-800 p-6 rounded-lg border border-slate-700">
        <h3 className="text-lg font-semibold mb-4">Create New Project</h3>
        <form onSubmit={handleCreateProject} className="flex gap-2">
          <input
            type="url"
            placeholder="GitHub repo URL (e.g., https://github.com/user/repo)"
            value={repoUrl}
            onChange={(e) => setRepoUrl(e.target.value)}
            className="flex-1 bg-slate-900 border border-slate-700 rounded px-4 py-2 text-white placeholder-slate-500 focus:outline-none focus:border-blue-500"
            required
          />
          <button
            type="submit"
            className="bg-blue-600 hover:bg-blue-700 px-6 py-2 rounded font-medium transition"
          >
            Create
          </button>
        </form>
      </div>

      <div>
        <h3 className="text-lg font-semibold mb-4">Your Projects</h3>
        {loading ? (
          <div className="text-slate-400">Loading...</div>
        ) : projects.length === 0 ? (
          <div className="text-slate-400">No projects yet. Create one above.</div>
        ) : (
          <div className="grid gap-4">
            {projects.map(project => (
              <div key={project.id} className="bg-slate-800 p-6 rounded-lg border border-slate-700">
                <div className="flex justify-between items-start mb-4">
                  <div>
                    <h4 className="text-lg font-semibold">{project.name}</h4>
                    <a
                      href={project.repo_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-sm text-blue-400 hover:text-blue-300"
                    >
                      {project.repo_url}
                    </a>
                    <div className="text-xs text-slate-500 mt-1">
                      Created: {new Date(project.created_at).toLocaleString()}
                    </div>
                  </div>
                  <button
                    onClick={() => handleDeploy(project.id)}
                    className="bg-green-600 hover:bg-green-700 px-4 py-2 rounded font-medium transition"
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
  )
}
