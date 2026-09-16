import React from 'react';
import UniqueLoading from "@/components/ui/morph-loading";

export default function DemoOne() {
  return (
    <div className="flex w-full min-h-screen items-center justify-center bg-slate-950">
      <UniqueLoading variant="morph" size="lg" className="w-32 h-32" />
    </div>
  );
}
