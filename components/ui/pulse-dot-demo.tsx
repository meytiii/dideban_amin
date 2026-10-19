import { PulseDot } from "@/components/ui/pulse-dot";
import React from "react";

export default function PulseDotDemo() {
  return (
    <div className="flex min-h-80 w-full items-center justify-center gap-8 bg-background text-foreground">
      <PulseDot className="size-3" />
      <PulseDot className="size-4" />
      <PulseDot className="size-6" />
    </div>
  );
}
