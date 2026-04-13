'use client'

import Link from 'next/link'
import { getGitHubAuthUrl } from '@/lib/auth'

export default function Auth() {
  const handleGitHubLogin = () => {
    const authUrl = getGitHubAuthUrl()
    window.location.href = authUrl
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-950 to-blue-900 flex items-center justify-center px-4">
      <div className="w-full max-w-md">
        <div className="bg-slate-800 border border-slate-700 rounded-lg p-8 space-y-6">
          {/* Header */}
          <div className="text-center">
            <Link href="/" className="text-2xl font-bold bg-gradient-to-r from-blue-400 to-cyan-400 bg-clip-text text-transparent">
              DeployDoc
            </Link>
            <p className="text-slate-400 mt-2">Sign in to your account</p>
          </div>

          {/* GitHub OAuth Button */}
          <button
            onClick={handleGitHubLogin}
            className="w-full bg-slate-900 hover:bg-slate-700 border border-slate-700 rounded-lg py-3 font-semibold flex items-center justify-center gap-2 transition"
          >
            <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
              <path fillRule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.49.5.092.682-.217.682-.482 0-.237-.008-.868-.013-1.703-2.782.603-3.369-1.343-3.369-1.343-.454-1.156-1.11-1.463-1.11-1.463-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.546 2.91 1.185.092-.923.35-1.546.636-1.903-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.025A9.578 9.578 0 0112 6.836c.85.004 1.705.114 2.504.336 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.942.359.31.678.921.678 1.856 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" clipRule="evenodd" />
            </svg>
            Continue with GitHub
          </button>

          {/* Divider */}
          <div className="relative">
            <div className="absolute inset-0 flex items-center">
              <div className="w-full border-t border-slate-700"></div>
            </div>
            <div className="relative flex justify-center text-sm">
              <span className="px-2 bg-slate-800 text-slate-400">Or continue with</span>
            </div>
          </div>

          {/* Email Form */}
          <form className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-2">Email</label>
              <input
                type="email"
                placeholder="you@example.com"
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white placeholder-slate-500 focus:outline-none focus:border-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-2">Password</label>
              <input
                type="password"
                placeholder="••••••••"
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white placeholder-slate-500 focus:outline-none focus:border-blue-500"
              />
            </div>
            <button
              type="submit"
              className="w-full bg-blue-600 hover:bg-blue-700 rounded-lg py-2 font-semibold transition"
            >
              Sign in
            </button>
          </form>

          {/* Footer */}
          <p className="text-center text-slate-400 text-sm">
            Don't have an account?{' '}
            <button className="text-blue-400 hover:text-blue-300">
              Sign up
            </button>
          </p>
        </div>

        {/* Back to home */}
        <div className="text-center mt-6">
          <Link href="/" className="text-slate-400 hover:text-blue-400 transition">
            ← Back to home
          </Link>
        </div>
      </div>
    </div>
  )
}
