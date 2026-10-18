import { FileCard } from "@/components/ui/file-card-collections";

const formats = [
  "pdf", "png", "jpg", "jpeg", "img", "zip", "rar", "doc", "xls", "xlsx", "txt", "video",
] as const;

export default function Playground() {
  return (
    <div className="bg-background text-foreground flex min-h-screen items-center justify-center">
      <div className="flex max-w-xl flex-wrap justify-center gap-6 px-6">
        {formats.map((format) => (
          <FileCard key={format} formatFile={format} />
        ))}
      </div>
    </div>
  );
}
