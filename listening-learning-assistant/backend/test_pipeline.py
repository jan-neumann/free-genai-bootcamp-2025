#!/usr/bin/env python3
"""
Test pipeline that demonstrates the integration between structured_data.py and vector_store.py
"""
import os
import sys
from pathlib import Path
from typing import Dict, List

# Add parent directory to path so we can import our modules
sys.path.append(str(Path(__file__).parent.parent))

from backend.structured_data import TranscriptStructurer, process_transcript
from backend.vector_store import initialize_vector_store

def main():
    # Path to the transcript file
    transcript_path = "backend/transcripts/2zr8KZb1DUs.txt"
    
    print("=== Starting Pipeline Test ===")
    print(f"Processing transcript: {transcript_path}")
    
    # Step 1: Process the transcript and save to files
    print("\n--- Step 1: Processing transcript ---")
    structured_sections = process_transcript(transcript_path, output_dir="questions")
    
    if not structured_sections:
        print("Error: Failed to process transcript")
        return
        
    print(f"Successfully processed {len(structured_sections)} sections")
    
    # Step 2: Initialize vector store
    print("\n--- Step 2: Initializing vector store ---")
    vector_store = initialize_vector_store()
    
    # Step 3: Add questions to vector store
    print("\n--- Step 3: Adding questions to vector store ---")
    for section_num, content in structured_sections.items():
        print(f"\nProcessing section {section_num}...")
        
        # Extract questions from the structured content
        structurer = TranscriptStructurer()
        questions = structurer.parse_questions_from_content(content, section_num)
        
        # Add each question to the vector store
        for q in questions:
            vector_store.add_question(q['text'], q['metadata'])
        
        print(f"Added {len(questions)} questions from section {section_num}")
    
    # Step 4: Verify the questions were added
    print("\n--- Step 4: Verifying questions in vector store ---")
    all_questions = vector_store.get_all_questions()
    print(f"Total questions in vector store: {len(all_questions)}")
    
    if all_questions:
        print("\nSample questions:")
        for i, q in enumerate(all_questions[:3], 1):
            print(f"{i}. {q['text']}")
    
    # Step 5: Test similarity search
    print("\n--- Step 5: Testing similarity search ---")
    test_queries = [
        "What should I say when I'm late for class?",
        "How do I ask for directions?",
        "What's the weather like today?"
    ]
    
    for query in test_queries:
        print(f"\nSearching for similar to: {query}")
        similar = vector_store.find_similar_questions(query, n_results=2)
        
        if similar:
            for i, q in enumerate(similar, 1):
                print(f"  {i}. Similarity: {1 - q['distance']:.2f}")
                print(f"     {q['text']}")
        else:
            print("  No similar questions found")
    
    print("\n=== Pipeline test completed successfully ===")

if __name__ == "__main__":
    main()
