"use server"

import { generateText } from "ai"
import { groq } from "@ai-sdk/groq"

type VocabularyItem = {
  japanese: string
  romaji: string
  english: string
  parts: string[]
}

export async function generateVocabulary(theme: string): Promise<VocabularyItem[]> {
  try {
    const prompt = `
You are a Japanese language expert. Generate a list of 10 Japanese vocabulary words related to the theme: "${theme}".

Your response must be a valid JSON array with objects having this exact structure:
{
  "japanese": "Japanese word in kanji/hiragana",
  "romaji": "romanized pronunciation",
  "english": "English translation",
  "parts": ["part of speech", "additional info like verb type if applicable"]
}

IMPORTANT FORMATTING RULES:
1. Use ONLY double quotes for property names and string values
2. Do not include any text before or after the JSON array
3. Do not include trailing commas
4. Ensure all property names and values are properly quoted
5. The response should start with '[' and end with ']'
6. Each item must have all four properties: japanese, romaji, english, and parts

Example of a correctly formatted response for the theme "food":
[
  {
    "japanese": "食べる",
    "romaji": "taberu",
    "english": "to eat",
    "parts": ["verb", "ichidan"]
  },
  {
    "japanese": "飲む",
    "romaji": "nomu",
    "english": "to drink",
    "parts": ["verb", "godan"]
  }
]

Ensure you include actual Japanese characters (kanji/hiragana/katakana) in the "japanese" field.
`

    const { text } = await generateText({
      model: groq("llama3-70b-8192"),
      prompt,
      temperature: 0.5, // Lower temperature for more consistent formatting
      maxTokens: 2048,
    })

    // Clean and prepare the text for JSON parsing
    let cleanedText = text.trim()

    // Find the first '[' and last ']' to extract just the JSON array
    const startIndex = cleanedText.indexOf("[")
    const endIndex = cleanedText.lastIndexOf("]")

    if (startIndex === -1 || endIndex === -1 || startIndex > endIndex) {
      throw new Error("Invalid JSON format: Could not find proper array brackets")
    }

    cleanedText = cleanedText.substring(startIndex, endIndex + 1)

    // Try to parse the JSON
    try {
      const jsonData = JSON.parse(cleanedText) as VocabularyItem[]
      return jsonData
    } catch (parseError) {
      console.error("JSON parse error:", parseError)

      // Attempt to fix common JSON formatting issues
      // Replace single quotes with double quotes
      const fixedText = cleanedText
        .replace(/'/g, '"')
        // Remove trailing commas before closing brackets
        .replace(/,\s*}/g, "}")
        .replace(/,\s*]/g, "]")

      try {
        const jsonData = JSON.parse(fixedText) as VocabularyItem[]
        return jsonData
      } catch (secondParseError) {
        console.error("Second JSON parse error:", secondParseError)
        throw new Error("Could not parse the generated vocabulary")
      }
    }
  } catch (error) {
    console.error("Error generating vocabulary:", error)
    throw new Error(`Failed to generate vocabulary: ${error instanceof Error ? error.message : "Unknown error"}`)
  }
}
