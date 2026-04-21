import axios, { AxiosInstance } from 'axios'
import { API_URL, getStoredToken, saveToken, getGitHubToken } from './auth'

class ApiClient {
  client: AxiosInstance

  constructor() {
    this.client = axios.create({
      baseURL: API_URL,
    })

    // Add token to requests
    this.client.interceptors.request.use((config) => {
      const token = getStoredToken()
      if (token) {
        config.headers.Authorization = `Bearer ${token}`
      }
      return config
    })

    // Handle responses
    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.status === 401) {
          // Token expired
          if (typeof window !== 'undefined') {
            window.location.href = '/auth'
          }
        }
        return Promise.reject(error)
      }
    )
  }

  async githubAuth(code: string) {
    const response = await this.client.post('/api/auth/github', { code })
    if (response.data.token) {
      saveToken(response.data.token)
    }
    return response.data
  }

  async getUser() {
    const response = await this.client.get('/api/auth/user')
    return response.data
  }

  async getGitHubRepos() {
    const githubToken = getGitHubToken()
    if (!githubToken) {
      throw new Error('GitHub token not available')
    }
    const response = await axios.get('https://api.github.com/user/repos', {
      headers: {
        Authorization: `Bearer ${githubToken}`,
        Accept: 'application/vnd.github.v3+json',
      },
      params: {
        sort: 'updated',
        per_page: 100,
      },
    })
    return response.data
  }

  async listProjects(userId?: number) {
    const response = await this.client.get('/api/project/list', {
      params: userId ? { user_id: userId } : undefined,
    })
    return response.data
  }

  async createProject(repoUrl: string, repoName?: string, userId?: number) {
    const normalizedRepoUrl = repoUrl.trim()
    const inferredName = repoName || normalizedRepoUrl.split('/').filter(Boolean).pop() || 'project'
    const localUserId = typeof window !== 'undefined' ? Number(localStorage.getItem('user_id') || '0') : 0
    const resolvedUserId = userId || localUserId || 1

    const response = await this.client.post('/create-project', {
      repo_url: normalizedRepoUrl,
      name: inferredName,
      user_id: resolvedUserId,
    })
    return response.data
  }

  async listDeployments() {
    const response = await this.client.get('/api/deployment/list')
    return response.data
  }

  async createDeployment(projectId: number | string) {
    const response = await this.client.post('/create-deployment', {
      project_id: Number(projectId),
    })
    return response.data
  }

  async getDeploymentStatus(deploymentId: string) {
    const response = await this.client.get(`/deployments/${deploymentId}/status`)
    return response.data
  }
}

export const apiClient = new ApiClient()

export default apiClient
