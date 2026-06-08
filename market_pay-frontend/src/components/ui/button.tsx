import { cn } from "@/lib/utils";
import { forwardRef, type ButtonHTMLAttributes } from "react";

const variants = {
  default: "bg-primary text-white hover:bg-primary-dark shadow-sm hover:shadow-md",
  secondary: "bg-accent text-white hover:bg-accent-dark shadow-sm hover:shadow-md",
  outline: "border border-primary text-primary hover:bg-primary/5 hover:border-primary-dark",
  ghost: "text-primary hover:bg-primary/5",
  danger: "bg-red-600 text-white hover:bg-red-700 shadow-sm hover:shadow-md",
};

const sizes = {
  sm: "h-8 px-3 text-xs",
  default: "h-10 px-5 text-sm",
  lg: "h-12 px-7 text-base",
};

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: keyof typeof variants;
  size?: keyof typeof sizes;
}

const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = "default", size = "default", ...props }, ref) => {
    return (
      <button
        className={cn(
          "inline-flex items-center justify-center rounded-lg font-medium transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2 active:scale-[0.97] disabled:cursor-not-allowed disabled:opacity-50",
          variants[variant],
          sizes[size],
          className
        )}
        ref={ref}
        {...props}
      />
    );
  }
);
Button.displayName = "Button";

export { Button };
