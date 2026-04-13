'use client'

import { Suspense } from 'react'
import CallbackContent from './callback-content'

export default function Callback() {
  return (
    <Suspense fallback={
      <div className="min-h-screen bg-gradient-to-br from-slate-950 to-blue-900 flex items-center justify-center">
        <div className="text-center space-y-4">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-blue-400 mx-auto"></div>
          <p className="text-slate-300">Authenticating with GitHub...</p>
        </div>
      </div>
    }>
      <CallbackContent />
    </Suspense>
  )
}
