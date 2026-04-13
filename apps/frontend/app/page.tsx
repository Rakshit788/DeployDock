'use client'

import Link from 'next/link'
import { useState } from 'react'

export default function Home() {
  const [isMenuOpen, setIsMenuOpen] = useState(false)

  return (
    <div className="min-h-screen bg-gradient-to-b from-slate-950 via-blue-900 to-slate-950 text-white">
      {/* Navigation */}
      <nav className="fixed top-0 w-full bg-slate-950/80 backdrop-blur border-b border-slate-800 z-50">
        <div className="max-w-7xl mx-auto px-4 py-4 flex justify-between items-center">
          <div className="text-2xl font-bold bg-gradient-to-r from-blue-400 to-cyan-400 bg-clip-text text-transparent">
            DeployDoc
          </div>
          <div className="hidden md:flex gap-8">
            <a href="#features" className="hover:text-blue-400 transition">Features</a>
            <a href="#how-it-works" className="hover:text-blue-400 transition">How It Works</a>
            <a href="#pricing" className="hover:text-blue-400 transition">Pricing</a>
            <Link href="/auth" className="bg-blue-600 hover:bg-blue-700 px-4 py-2 rounded transition">
              Get Started
            </Link>
          </div>
          <button
            className="md:hidden text-2xl"
            onClick={() => setIsMenuOpen(!isMenuOpen)}
          >
            ☰
          </button>
        </div>
        {isMenuOpen && (
          <div className="md:hidden bg-slate-900 border-t border-slate-800 p-4 space-y-2">
            <a href="#features" className="block hover:text-blue-400">Features</a>
            <a href="#how-it-works" className="block hover:text-blue-400">How It Works</a>
            <a href="#pricing" className="block hover:text-blue-400">Pricing</a>
            <Link href="/auth" className="block bg-blue-600 px-4 py-2 rounded text-center">
              Get Started
            </Link>
          </div>
        )}
      </nav>

      {/* Hero Section */}
      <section className="max-w-7xl mx-auto px-4 pt-32 pb-20 text-center">
        <div className="space-y-6">
          <h1 className="text-6xl md:text-7xl font-bold leading-tight">
            Deploy Your Projects
            <br />
            <span className="bg-gradient-to-r from-blue-400 via-cyan-400 to-blue-400 bg-clip-text text-transparent">
              Effortlessly
            </span>
          </h1>
          <p className="text-xl md:text-2xl text-slate-300 max-w-2xl mx-auto">
            Connect your GitHub repositories and deploy instantly. No configuration needed. Just push to deploy.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center pt-8">
            <Link
              href="/auth"
              className="bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 px-8 py-4 rounded-lg font-semibold text-lg transition transform hover:scale-105"
            >
              Start Deploying Now →
            </Link>
            <button className="border-2 border-slate-600 hover:border-blue-400 px-8 py-4 rounded-lg font-semibold text-lg transition">
              View Documentation
            </button>
          </div>
        </div>

        {/* Hero Image/Animation */}
        <div className="mt-16 relative">
          <div className="absolute inset-0 bg-gradient-to-r from-blue-500 to-cyan-500 rounded-lg blur-3xl opacity-20"></div>
          <div className="relative bg-slate-800 border border-slate-700 rounded-lg p-8 backdrop-blur">
            <div className="text-sm text-slate-400 mb-4">$ git push main</div>
            <div className="space-y-2 text-left text-slate-300 font-mono text-sm">
              <div>🚀 Deploying...</div>
              <div>✅ Build succeeded</div>
              <div>✅ Tests passed</div>
              <div>✅ Deployed to: <span className="text-green-400">https://your-app.vercel-clone.dev</span></div>
            </div>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section id="features" className="max-w-7xl mx-auto px-4 py-20 border-t border-slate-800">
        <h2 className="text-5xl font-bold text-center mb-16">
          Powerful Features
        </h2>
        <div className="grid md:grid-cols-3 gap-8">
          {[
            {
              icon: '⚡',
              title: 'Auto Deployments',
              description: 'Push to GitHub and your app deploys automatically in seconds'
            },
            {
              icon: '🔐',
              title: 'GitHub Integration',
              description: 'Seamless OAuth authentication with your GitHub account'
            },
            {
              icon: '📊',
              title: 'Live Monitoring',
              description: 'Real-time deployment status and logs'
            },
            {
              icon: '🌐',
              title: 'Multiple Environments',
              description: 'Deploy different versions across staging and production'
            },
            {
              icon: '🔄',
              title: 'Instant Rollback',
              description: 'Revert to any previous deployment with one click'
            },
            {
              icon: '💻',
              title: 'Full Control',
              description: 'Manage Docker containers directly from your dashboard'
            }
          ].map((feature, i) => (
            <div key={i} className="bg-slate-800/50 border border-slate-700 rounded-lg p-6 hover:border-blue-500 transition">
              <div className="text-4xl mb-4">{feature.icon}</div>
              <h3 className="text-xl font-semibold mb-2">{feature.title}</h3>
              <p className="text-slate-300">{feature.description}</p>
            </div>
          ))}
        </div>
      </section>

      {/* How It Works */}
      <section id="how-it-works" className="max-w-7xl mx-auto px-4 py-20 border-t border-slate-800">
        <h2 className="text-5xl font-bold text-center mb-16">
          How It Works
        </h2>
        <div className="space-y-8">
          {[
            { step: 1, title: 'Sign in with GitHub', desc: 'Connect your GitHub account in one click' },
            { step: 2, title: 'Select a Repository', desc: 'Choose which repositories to deploy' },
            { step: 3, title: 'Click Deploy', desc: 'Start your deployment with a single button click' },
            { step: 4, title: 'Watch it Live', desc: 'Monitor deployment progress in real-time' }
          ].map((item) => (
            <div key={item.step} className="flex gap-6 items-start">
              <div className="flex-shrink-0">
                <div className="flex items-center justify-center h-12 w-12 rounded-md bg-blue-600 text-white text-xl font-bold">
                  {item.step}
                </div>
              </div>
              <div>
                <h3 className="text-xl font-semibold mb-2">{item.title}</h3>
                <p className="text-slate-300">{item.desc}</p>
              </div>
            </div>
          ))}
        </div>
      </section>

      {/* Pricing Section */}
      <section id="pricing" className="max-w-7xl mx-auto px-4 py-20 border-t border-slate-800">
        <h2 className="text-5xl font-bold text-center mb-16">
          Simple Pricing
        </h2>
        <div className="grid md:grid-cols-3 gap-8">
          {[
            { name: 'Starter', price: 'Free', features: ['1 Project', 'Basic Deployments', 'Community Support'] },
            { name: 'Pro', price: '$29', features: ['Unlimited Projects', 'Advanced Deployments', 'Priority Support', 'Custom Domains'], highlighted: true },
            { name: 'Enterprise', price: 'Custom', features: ['Everything in Pro', 'Dedicated Support', 'SLA', 'Custom Features'] }
          ].map((plan, i) => (
            <div
              key={i}
              className={`border rounded-lg p-8 ${
                plan.highlighted
                  ? 'border-blue-500 bg-blue-500/10 transform md:scale-105'
                  : 'border-slate-700 bg-slate-800/50'
              }`}
            >
              <h3 className="text-2xl font-bold mb-2">{plan.name}</h3>
              <div className="text-4xl font-bold mb-6">{plan.price}<span className="text-lg text-slate-400">/mo</span></div>
              <ul className="space-y-3 mb-8">
                {plan.features.map((feature) => (
                  <li key={feature} className="flex items-center gap-2">
                    <span className="text-green-400">✓</span>
                    <span>{feature}</span>
                  </li>
                ))}
              </ul>
              <button className={`w-full py-2 rounded font-semibold transition ${
                plan.highlighted
                  ? 'bg-blue-600 hover:bg-blue-700'
                  : 'border border-slate-600 hover:border-blue-400'
              }`}>
                Get Started
              </button>
            </div>
          ))}
        </div>
      </section>

      {/* CTA Section */}
      <section className="max-w-7xl mx-auto px-4 py-20 border-t border-slate-800">
        <div className="bg-gradient-to-r from-blue-600 to-cyan-600 rounded-lg p-12 text-center">
          <h2 className="text-4xl font-bold mb-4">Ready to Deploy?</h2>
          <p className="text-lg mb-8 opacity-90">Start deploying your projects in minutes with DeployDoc</p>
          <Link
            href="/auth"
            className="bg-slate-950 hover:bg-slate-900 px-8 py-3 rounded-lg font-semibold transition inline-block"
          >
            Get Started Free →
          </Link>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-slate-800 py-12 text-center text-slate-400">
        <p>&copy; 2026 DeployDoc. All rights reserved.</p>
      </footer>
    </div>
  )
}
