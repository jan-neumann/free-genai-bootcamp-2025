import os
import sys
from pathlib import Path
from typing import List, Dict

# Add parent directory to path to allow imports
sys.path.append(str(Path(__file__).parent.parent))

from backend.vector_store import QuestionVectorStore
from backend.structured_data import process_transcript

def test_basic_operations():
    print("=== Testing Basic Vector Store Operations ===")
    
    # Initialize test vector store in a temporary directory
    test_db = "test_chroma_db"
    if os.path.exists(test_db):
        import shutil
        shutil.rmtree(test_db)
    
    vector_store = QuestionVectorStore(persist_directory=test_db)
    
    # Test adding a single question
    test_question = "How do I ask for directions to the train station?"
    qid = vector_store.add_question(
        test_question,
        {"difficulty": "easy", "topic": "directions"}
    )
    print(f"Added question with ID: {qid}")
    
    # Test retrieving the question
    retrieved = vector_store.get_question(qid)
    print(f"Retrieved question: {retrieved['text'] if retrieved else 'Not found'}")
    
    # Test similarity search
    similar = vector_store.find_similar_questions("Where is the train station?")
    print(f"\nFound {len(similar)} similar questions:")
    for i, q in enumerate(similar, 1):
        print(f"{i}. Similarity: {1 - q['distance']:.2f}")
        print(f"   Question: {q['text']}")
    
    # Clean up
    if os.path.exists(test_db):
        import shutil
        shutil.rmtree(test_db)
    print("\nBasic operations test completed!")

def test_transcript_processing():
    print("\n=== Testing Transcript Processing ===")
    
    # Path to a test transcript
    test_transcript = "transcripts/2zr8KZb1DUs.txt"
    if not os.path.exists(test_transcript):
        print(f"Test transcript not found at {test_transcript}")
        return
    
    # Process the transcript
    print(f"Processing test transcript: {test_transcript}")
    process_transcript(test_transcript, "test_questions")
    
    # Initialize the vector store to verify
    vector_store = QuestionVectorStore()
    all_questions = vector_store.get_all_questions()
    
    print(f"\nTotal questions in vector store: {len(all_questions)}")
    if all_questions:
        print("\nSample questions:")
        for i, q in enumerate(all_questions[:3], 1):
            print(f"{i}. {q['text']}")
    
    # Test similarity search with a sample query
    query = "What should I say when I'm late for class?"
    print(f"\nSearching for similar to: {query}")
    similar = vector_store.find_similar_questions(query)
    
    if similar:
        print("\nMost similar questions found:")
        for i, q in enumerate(similar[:3], 1):
            print(f"{i}. Similarity: {1 - q['distance']:.2f}")
            print(f"   Question: {q['text']}")
    else:
        print("No similar questions found.")

if __name__ == "__main__":
    # Run basic tests
    test_basic_operations()
    
    # Run transcript processing test
    test_transcript_processing()
