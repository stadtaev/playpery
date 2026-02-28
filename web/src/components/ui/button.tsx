import { forwardRef, type ButtonHTMLAttributes } from "react"
import { cva, type VariantProps } from "class-variance-authority"
import { motion, type HTMLMotionProps } from "framer-motion"
import { cn } from "@/lib/utils"

const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring disabled:pointer-events-none disabled:opacity-50 cursor-pointer",
  {
    variants: {
      variant: {
        default: "bg-accent text-accent-foreground hover:bg-accent-hover",
        secondary: "bg-card text-text-primary border border-border hover:bg-popover",
        outline: "border border-border text-text-primary hover:bg-card hover:border-border-hover",
        ghost: "text-text-secondary hover:bg-card hover:text-text-primary",
        destructive: "bg-destructive text-white hover:bg-destructive-hover",
      },
      size: {
        sm: "h-8 px-3 text-xs",
        default: "h-9 px-4",
        lg: "h-10 px-6 text-base",
        icon: "h-9 w-9",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
)

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> &
  VariantProps<typeof buttonVariants>

const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, ...props }, ref) => (
    <button
      className={cn(buttonVariants({ variant, size, className }))}
      ref={ref}
      {...props}
    />
  )
)
Button.displayName = "Button"

type MotionButtonProps = HTMLMotionProps<"button"> &
  VariantProps<typeof buttonVariants>

const MotionButton = forwardRef<HTMLButtonElement, MotionButtonProps>(
  ({ className, variant, size, ...props }, ref) => (
    <motion.button
      className={cn(buttonVariants({ variant, size, className }))}
      whileTap={{ scale: 0.97 }}
      ref={ref}
      {...props}
    />
  )
)
MotionButton.displayName = "MotionButton"

export { Button, MotionButton, buttonVariants }
