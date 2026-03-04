export function Spinner({ size = 'default', className = '' }: { size?: 'default' | 'lg'; className?: string }) {
  return <span className={`spinner ${size === 'lg' ? 'spinner-lg' : ''} ${className}`} />
}

export function LoadingPage({ message = 'Loading...' }: { message?: string }) {
  return (
    <div className="page flex flex-col items-center gap-4 pt-24">
      <Spinner size="lg" />
      <p className="text-secondary text-sm">{message}</p>
    </div>
  )
}
