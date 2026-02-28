import { forwardRef, type LabelHTMLAttributes } from "react"
import { cn } from "@/lib/utils"

const Label = forwardRef<HTMLLabelElement, LabelHTMLAttributes<HTMLLabelElement>>(
  ({ className, ...props }, ref) => (
    <label
      className={cn("text-sm font-medium text-text-secondary", className)}
      ref={ref}
      {...props}
    />
  )
)
Label.displayName = "Label"

export { Label }
