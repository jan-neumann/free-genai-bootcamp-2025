import VocabularyImporter from "@/components/vocabulary-importer"

export default function Home() {
  return (
    <main className="container mx-auto py-10 px-4">
      <h1 className="text-3xl font-bold mb-6 text-center">Japanese Vocabulary Importer</h1>
      <p className="text-center mb-8 text-gray-600 max-w-2xl mx-auto">
        Enter a thematic category (e.g., food, travel, business) to generate Japanese vocabulary related to that theme.
      </p>
      <div className="max-w-3xl mx-auto">
        <VocabularyImporter />
      </div>
    </main>
  )
}
