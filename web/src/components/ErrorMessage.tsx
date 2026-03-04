export function ErrorMessage({ message }: { message: string }) {
  return <p className="text-feedback-error" role="alert">{message}</p>
}
