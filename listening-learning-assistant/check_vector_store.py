import os
import sys
from pathlib import Path

# Add the project root to the path
project_root = str(Path(__file__).parent)
sys.path.append(project_root)

from backend.vector_store import initialize_vector_store

def list_vector_store_contents():
    try:
        # Initialize the vector store
        vector_store = initialize_vector_store()
        
        # Get the collection
        collection = vector_store.collection
        
        # Get all items in the collection
        results = collection.get(include=["documents", "metadatas"])
        
        if not results or not results.get('ids'):
            print("Vector store is empty.")
            return
        
        print(f"Found {len(results['ids'])} items in the vector store:\n")
        
        # Print each item with its metadata
        for i, (doc_id, doc, metadata) in enumerate(zip(results['ids'], results['documents'], results['metadatas'])):
            print(f"Item {i+1}:")
            print(f"  ID: {doc_id}")
            print(f"  Content: {doc[:200]}..."  # Show first 200 chars
                  if len(doc) > 200 else f"  Content: {doc}")
            if metadata:
                print("  Metadata:")
                for key, value in metadata.items():
                    print(f"    {key}: {value}")
            print("-" * 80)
            
    except Exception as e:
        print(f"Error accessing vector store: {str(e)}")

if __name__ == "__main__":
    list_vector_store_contents()
