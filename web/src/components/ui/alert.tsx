import { type HTMLAttributes } from "react"
import { cva, type VariantProps } from "class-variance-authority"
import { AlertCircle, CheckCircle2, Info, AlertTriangle } from "lucide-react"
import { cn } from "@/lib/utils"

const alertVariants = cva(
  "flex items-start gap-3 rounded-lg border p-3 text-sm",
  {
    variants: {
      variant: {
        error: "border-error/30 bg-error/10 text-error",
        success: "border-success/30 bg-success/10 text-success",
        warning: "border-warning/30 bg-warning/10 text-warning",
        info: "border-info/30 bg-info/10 text-info",
      },
    },
    defaultVariants: {
      variant: "info",
    },
  }
)

const iconMap = {
  error: AlertCircle,
  success: CheckCircle2,
  warning: AlertTriangle,
  info: Info,
}

type AlertProps = HTMLAttributes<HTMLDivElement> &
  VariantProps<typeof alertVariants>

function Alert({ className, variant, children, ...props }: AlertProps) {
  const Icon = iconMap[variant ?? "info"]
  return (
    <div className={cn(alertVariants({ variant, className }))} {...props}>
      <Icon size={16} className="mt-0.5 shrink-0" />
      <div>{children}</div>
    </div>
  )
}

export { Alert, alertVariants }
