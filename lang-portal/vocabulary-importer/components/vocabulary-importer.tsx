"use client"

import type React from "react"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { Card, CardContent } from "@/components/ui/card"
import { Loader2, Copy, Check, AlertCircle } from "lucide-react"
import { generateVocabulary } from "@/app/actions/vocabulary-actions"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"

export default function VocabularyImporter() {
  const [theme, setTheme] = useState("")
  const [result, setResult] = useState("")
  const [isLoading, setIsLoading] = useState(false)
  const [copied, setCopied] = useState(false)
  const [error, setError] = useState("")

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!theme.trim()) return

    setIsLoading(true)
    setError("")
    setResult("")

    try {
      const data = await generateVocabulary(theme)

      // Validate the data structure
      if (!Array.isArray(data)) {
        throw new Error("Invalid response format: not an array")
      }

      if (data.length === 0) {
        throw new Error("No vocabulary items were generated")
      }

      // Check if any items are missing required fields
      const invalidItems = data.filter(
        (item) => !item.japanese || !item.romaji || !item.english || !Array.isArray(item.parts),
      )

      if (invalidItems.length > 0) {
        console.warn("Some items are missing required fields:", invalidItems)
      }

      // Check if any items are missing Japanese characters
      const missingJapanese = data.some(
        (item) => !/[\u3000-\u303f\u3040-\u309f\u30a0-\u30ff\uff00-\uff9f\u4e00-\u9faf]/.test(item.japanese),
      )

      if (missingJapanese) {
        setError("Note: Some items may be missing proper Japanese characters. You might want to try again.")
      }

      setResult(JSON.stringify(data, null, 2))
    } catch (err) {
      console.error("Error in component:", err)
      setError(err instanceof Error ? err.message : "Failed to generate vocabulary. Please try again.")
    } finally {
      setIsLoading(false)
    }
  }

  const copyToClipboard = async () => {
    try {
      await navigator.clipboard.writeText(result)
      setCopied(true)
      alert("Vocabulary copied to clipboard!")
      setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      console.error("Failed to copy:", err)
    }
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardContent className="pt-6">
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <label htmlFor="theme" className="text-sm font-medium">
                Thematic Category
              </label>
              <Input
                id="theme"
                placeholder="Enter a theme (e.g., food, travel, business)"
                value={theme}
                onChange={(e) => setTheme(e.target.value)}
                disabled={isLoading}
              />
            </div>
            <Button type="submit" className="w-full" disabled={isLoading}>
              {isLoading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Generating...
                </>
              ) : (
                "Generate Vocabulary"
              )}
            </Button>
          </form>
        </CardContent>
      </Card>

      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {result && (
        <Card>
          <CardContent className="pt-6 space-y-4">
            <div className="flex justify-between items-center">
              <h3 className="text-lg font-medium">Generated Vocabulary</h3>
              <Button variant="outline" size="sm" onClick={copyToClipboard} className="flex items-center gap-1">
                {copied ? (
                  <>
                    <Check className="h-4 w-4" />
                    Copied
                  </>
                ) : (
                  <>
                    <Copy className="h-4 w-4" />
                    Copy
                  </>
                )}
              </Button>
            </div>
            <Textarea value={result} readOnly className="font-mono text-sm h-96" />
          </CardContent>
        </Card>
      )}
    </div>
  )
}
