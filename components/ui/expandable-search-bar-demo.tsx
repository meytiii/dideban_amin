import ExpandableSearchBar from "@/components/ui/expandable-search-bar";

export default function ExpandableSearchBarDemo() {
  return (
    <main className="min-h-screen flex flex-col items-center justify-center gap-16">
      <div className="w-full max-w-xs sm:max-w-md flex flex-col items-center">
        <ExpandableSearchBar
          expandDirection="right"
          width={250}
          placeholder="Search right..."
          onSearch={q => console.log("search right", q)}
          className="w-full"
        />
      </div>

      <div className="w-full max-w-xs sm:max-w-md flex flex-col items-center">
        <div className="w-full flex justify-end">
          <ExpandableSearchBar
            expandDirection="left"
            width={250}
            placeholder="Search left..."
            onSearch={q => console.log("search left", q)}
            className="w-full"
          />
        </div>
      </div>
    </main>
  );
}
