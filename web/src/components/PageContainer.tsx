const sizeClasses = {
  sm: 'page',
  md: 'page-md',
  wide: 'page-wide',
} as const

export function PageContainer({ size = 'sm', className = '', children }: {
  size?: keyof typeof sizeClasses
  className?: string
  children: React.ReactNode
}) {
  return (
    <main className={`${sizeClasses[size]} ${className}`}>
      {children}
    </main>
  )
}
