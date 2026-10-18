import { cn } from "@/lib/utils";
import { ReactNode } from "react";

type FormatFileProps =
  | "doc"
  | "pdf"
  | "xls"
  | "xlsx"
  | "txt"
  | "zip"
  | "rar"
  | "img"
  | "png"
  | "jpg"
  | "jpeg"
  | "video";

type FileCardProps = {
  formatFile: FormatFileProps;
  className?: string;
};

const DefaultPlaceholder = () => {
  return (
    <div className="space-y-1.5">
      <div className="flex gap-2">
        <div className="bg-foreground/20 h-0.5 w-1/2 rounded-full" />
      </div>
      <div className="flex gap-1">
        <div className="bg-foreground/10 h-0.5 w-1/3 rounded-full" />
        <div className="bg-foreground/10 h-0.5 w-1/3 rounded-full" />
      </div>
      <div className="flex gap-1">
        <div className="bg-foreground/10 h-0.5 w-1/2 rounded-full" />
        <div className="bg-foreground/10 h-0.5 w-1/3 rounded-full" />
      </div>
      <div className="flex gap-1">
        <div className="bg-foreground/10 h-0.5 w-1/3 rounded-full" />
        <div className="bg-foreground/10 h-0.5 w-1/3 rounded-full" />
      </div>
      <div className="flex gap-1">
        <div className="bg-foreground/10 h-0.5 w-1/3 rounded-full" />
        <div className="bg-foreground/10 h-0.5 w-1/2 rounded-full" />
      </div>
      <div className="flex gap-1">
        <div className="bg-foreground/10 h-0.5 w-1/3 rounded-full" />
      </div>
    </div>
  );
};

const colorBannerMap: Record<FormatFileProps, string> = {
  doc: "bg-blue-500 text-white",
  pdf: "bg-red-500 text-white",
  txt: "bg-gray-500 text-white",
  xls: "bg-emerald-600 text-white",
  xlsx: "bg-emerald-600 text-white",
  zip: "bg-purple-500 text-white",
  rar: "bg-purple-600 text-white",
  img: "bg-emerald-600 text-white",
  png: "bg-teal-600 text-white",
  jpg: "bg-green-700 text-white",
  jpeg: "bg-green-700 text-white",
  video: "bg-amber-600 text-white",
};

export const FileCard = ({ formatFile, className }: FileCardProps) => {
  const colorBannerClass = colorBannerMap[formatFile] || colorBannerMap.doc;
  let filePlaceholder: ReactNode = <DefaultPlaceholder />;

  if (formatFile === "xls" || formatFile === "xlsx") {
    filePlaceholder = (
      <div className="space-y-0.5">
        <div className="grid grid-cols-3 gap-0.5">
          <div className="bg-foreground/20 h-2" />
          <div className="bg-foreground/20 h-2" />
          <div className="bg-foreground/20 h-2" />
        </div>
        <div className="grid grid-cols-3 gap-0.5">
          <div className="bg-foreground/5 h-2" />
          <div className="bg-foreground/5 h-2" />
          <div className="bg-foreground/5 h-2" />
          <div className="bg-foreground/5 h-2" />
          <div className="bg-foreground/5 h-2" />
          <div className="bg-foreground/5 h-2" />
        </div>
        <div className="grid grid-cols-3 gap-0.5">
          <div className="bg-foreground/5 h-2" />
          <div className="bg-foreground/5 h-2" />
        </div>
        <div className="grid grid-cols-3 gap-0.5">
          <div className="bg-foreground/5 h-2" />
        </div>
      </div>
    );
  }

  if (formatFile === "zip" || formatFile === "rar") {
    filePlaceholder = (
      <div className="relative flex h-full flex-col items-center justify-center">
        <div className="space-y-0">
          <div className="flex overflow-hidden rounded-full">
            <div className="bg-foreground/20 size-1.5" />
            <div className="bg-foreground/5 size-1.5" />
          </div>
          <div className="flex overflow-hidden rounded-full">
            <div className="bg-foreground/5 size-1.5" />
            <div className="bg-foreground/20 size-1.5" />
          </div>
          <div className="flex overflow-hidden rounded-full">
            <div className="bg-foreground/20 size-1.5" />
            <div className="bg-foreground/5 size-1.5" />
          </div>
          <div className="flex overflow-hidden rounded-full">
            <div className="bg-foreground/5 size-1.5" />
            <div className="bg-foreground/20 size-1.5" />
          </div>
          <div className="flex overflow-hidden rounded-full">
            <div className="bg-foreground/20 size-1.5" />
            <div className="bg-foreground/5 size-1.5" />
          </div>
          <div className="flex overflow-hidden rounded-full">
            <div className="bg-foreground/5 size-1.5" />
            <div className="bg-foreground/20 size-1.5" />
          </div>
          <div className="flex overflow-hidden rounded-full">
            <div className="bg-foreground/20 size-1.5" />
            <div className="bg-foreground/5 size-1.5" />
          </div>
          <div className="flex overflow-hidden rounded-full">
            <div className="bg-foreground/5 size-1.5" />
            <div className="bg-foreground/20 size-1.5" />
          </div>
          <div className="flex overflow-hidden rounded-full">
            <div className="bg-foreground/20 size-1.5" />
            <div className="bg-foreground/5 size-1.5" />
          </div>
        </div>
      </div>
    );
  }

  if (
    formatFile === "img" ||
    formatFile === "png" ||
    formatFile === "jpg" ||
    formatFile === "jpeg"
  ) {
    filePlaceholder = (
      <div className="bg-foreground/5 mb-1.5 space-y-1 rounded border p-1">
        <div className="flex justify-center gap-1">
          <div className="size-3 rounded-sm bg-yellow-400/40" />
        </div>
        <div className="bg-foreground/15 mx-auto mt-1 h-0.75 w-4 rounded-full" />
        <div className="bg-foreground/15 mx-auto h-0.75 w-8 rounded-full" />
      </div>
    );
  }

  if (formatFile === "video") {
    filePlaceholder = (
      <div className="bg-foreground/5 mb-1.5 space-y-1 rounded border p-1">
        <div className="flex justify-center gap-1">
          <div className="size-0 border-y-[5px] border-l-8 border-y-transparent border-l-green-400/60" />
        </div>
        <div className="bg-foreground/15 mx-auto mt-1 h-0.75 w-4 rounded-full" />
        <div className="bg-foreground/15 mx-auto h-0.75 w-8 rounded-full" />
      </div>
    );
  }

  const sizeClass = "w-14 h-18";

  return (
    <div aria-hidden className={cn("relative size-fit", className)}>
      <div
        className={cn(
          "absolute -right-2 bottom-1.5 z-2 rounded px-1.5 py-0.5 text-[8px] font-medium uppercase shadow-xs",
          colorBannerClass
        )}
      >
        {formatFile}
      </div>
      <div
        className={cn(
          "dark:bg-secondary ring-border relative z-1 space-y-3 rounded-md bg-white p-2 ring-1",
          sizeClass
        )}
      >
        {filePlaceholder}
      </div>
    </div>
  );
};

export default FileCard;
