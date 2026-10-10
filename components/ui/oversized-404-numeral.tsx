import { Home, ArrowRight } from "lucide-react";
import * as React from "react";

import { Button } from "@/components/ui/button";

export default function ErrorBlock() {
  return (
    <section className="flex min-h-svh w-full items-center justify-center bg-background px-6 py-16 text-foreground" dir="rtl">
      <div className="mx-auto flex max-w-lg flex-col items-center text-center">
        <span
          className="text-[8rem] leading-none font-bold tracking-tighter text-foreground tabular-nums select-none sm:text-[11rem]"
          aria-hidden="true"
        >
          404
        </span>
        <h1 className="-mt-2 font-heading text-2xl font-bold tracking-tight sm:text-3xl">
          صفحه مورد نظر شما پیدا نشد.
        </h1>
        <p className="mt-3 max-w-sm text-sm text-pretty text-muted-foreground">
          صفحه‌ای که دنبالش هستید وجود ندارد یا ممکن است منتقل شده باشد. لطفاً به صفحه اصلی بازگردید.
        </p>
        <div className="mt-8 flex flex-col gap-3 sm:flex-row">
          <Button asChild>
            <a href="/">
              <Home className="ml-2 h-4 w-4" aria-hidden="true" />
              صفحه اصلی
            </a>
          </Button>
          <Button asChild variant="outline">
            <a
              href="#"
              onClick={(e) => {
                e.preventDefault();
                if (typeof window !== "undefined") {
                  window.history.back();
                }
              }}
            >
              <ArrowRight className="ml-2 h-4 w-4" aria-hidden="true" />
              بازگشت به صفحه قبلی
            </a>
          </Button>
        </div>
      </div>
    </section>
  );
}
