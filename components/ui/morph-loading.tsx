"use client"

import React from "react"
import { cn } from "@/lib/utils"

interface UniqueLoadingProps {
  variant?: "morph"
  size?: "sm" | "md" | "lg"
  className?: string
}

export default function UniqueLoading({
  variant = "morph",
  size = "md",
  className,
}: UniqueLoadingProps) {
  const containerSizes = {
    sm: "w-16 h-16",
    md: "w-24 h-24",
    lg: "w-32 h-32",
  }

  if (variant === "morph") {
    return (
      <div className={cn("relative", containerSizes[size], className)}>
        <div className="absolute inset-0 flex items-center justify-center">
          {[0, 1, 2, 3].map((i) => (
            <div
              key={i}
              className="absolute w-4 h-4 bg-sky-400 dark:bg-white shadow-[0_0_12px_rgba(56,189,248,0.7)]"
              style={{
                animation: `morph-${i} 2s infinite ease-in-out`,
                animationDelay: `${i * 0.2}s`,
              }}
            />
          ))}
        </div>
      </div>
    )
  }

  return null
}
