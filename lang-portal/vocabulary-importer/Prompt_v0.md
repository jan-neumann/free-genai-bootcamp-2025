## Prompt (given to v0.dev)

Please give me a vocabulary language importer where we have a text field that allows us to import a thematic category for the generation of language vocabulary.

When submitting that text field, it should hit an api endpoint (api route in app router) to invoke an LLM chat completion in Groq (LLM) on the server-side and then pass that information back to the front-end.

It has to create a structured json output:

```
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
  },
]
```

The json that is outputted back to the front-end should be copy-able... so it should be sent to an input field and there should be a copy button so that it can be copied to the clipboard and that should give an alert that it was copied to the user's clipboard. 

The app should use app router and the latest version of next.js... and the llm calls should run on an api route on the server-side.