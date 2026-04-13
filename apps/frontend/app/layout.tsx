import type { Metadata } from 'next'
import './globals.css'

export const metadata: Metadata = {
  title: 'DeployDoc - Deploy with Ease',
  description: 'Deploy your projects effortlessly',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body className="bg-slate-950 text-white">
        {children}
      </body>
    </html>
  )
}
