import axios, { AxiosInstance } from 'axios'
import { API_URL, getStoredToken, saveToken } from './auth'

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

  async listProjects() {
    const response = await this.client.get('/api/project/list')
    return response.data
  }

  async createProject(repoUrl: string) {
    const response = await this.client.post('/api/project/create', {
      repo_url: repoUrl,
    })
    return response.data
  }

  async listDeployments() {
    const response = await this.client.get('/api/deployment/list')
    return response.data
  }

  async createDeployment(projectId: string) {
    const response = await this.client.post('/api/deployment/create', {
      project_id: projectId,
    })
    return response.data
  }
}

export const apiClient = new ApiClient()

export default apiClient
