# Implementation Guide

## User Authentication & Credentials Flow

### 1. GitHub Login → Callback
**File:** `app/auth/callback/callback-content.tsx`
- User clicks GitHub login
- Gets redirected to GitHub OAuth
- GitHub redirects back with authorization code
- Frontend exchanges code with API via `apiClient.githubAuth(code)`
- **Stores:**
  - JWT/Auth token (from API) → used for authenticated API calls
  - GitHub access token (from API) → used to fetch user's repos
  - User ID and username → for UI display

### 2. Dashboard - Three Sections

#### Overview Tab
- Shows stats (repos, projects, deployments count)
- Lists recent deployments
- Quick action buttons

#### Repos Tab (GitHub Integration)
```
User's GitHub repos fetched via:
- Uses stored GitHub access token
- Makes direct call to GitHub API
- Shows all public repositories
- "Create Project" button for each repo
```

#### Projects Tab (Deployments)
```
Shows created projects with:
- Project name & subdomain
- Deploy button (creates new deployment)
- Visit link (accesses deployed app)
- Recent deployment history
```

## API Endpoints Used

From backend that we're calling:

```
POST /auth/github/callback      - Exchange code for token
GET  /create-project            - Create project from repo
POST /create-deployment         - Deploy a project
GET  /deployments/{id}/status   - Check deployment status
GET  /project/list              - List user's projects
GET  /deployments               - List all deployments
```

## Data Flow

```
GitHub Login
    ↓
[Callback Handler]
    ↓
Exchange Code → Get Token
    ↓
Store: JWT Token + GitHub Token
    ↓
[Dashboard]
    ├─ Repos Tab
    │   ├─ Use GitHub token
    │   └─ Fetch repos from GitHub API
    │
    ├─ Projects Tab
    │   ├─ Fetch projects from backend
    │   └─ Fetch deployments from backend
    │
    └─ Overview Tab
        └─ Show stats & recent deployments
```

## How to Use

### For Development

1. Make sure backend is running on `:8080`
2. Frontend should be on `localhost:3000`
3. GitHub OAuth app configured with redirect URI: `{your-url}/auth/callback`

### User Actions

1. **View Repos:**
   - Go to "Repos" tab
   - All GitHub repos automatically loaded
   - Click "Create Project" to create from a repo

2. **Deploy Project:**
   - Go to "Projects" tab
   - Click "Deploy" on any project
   - Watch deployment status update
   - Click "Visit" once deployed

3. **Monitor Deployments:**
   - See all recent deployments in "Overview" tab
   - Check individual project deployment history in "Projects" tab
   - Status badge shows: pending/deployed/failed

## Key Features

✅ **No Database Changes** - Uses existing schema
✅ **GitHub Integration** - Direct API calls for repos
✅ **Credential Management** - Tokens stored in localStorage
✅ **Project Management** - Create and deploy from one place
✅ **Status Tracking** - Real-time deployment status
✅ **User-Friendly** - Tab-based interface with clear actions

## Notes

- Tokens cleared on logout
- Protected routes redirect to auth if not authenticated  
- Error handling for failed API calls
- Loading states for all async operations
- Responsive design for mobile/desktop
