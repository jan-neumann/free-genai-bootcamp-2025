#!/usr/bin/env python3
"""
Test script to verify imports are working correctly.
"""
import sys
import os
from pathlib import Path

# Add the project root to the Python path
project_root = str(Path(__file__).parent)
if project_root not in sys.path:
    sys.path.insert(0, project_root)

try:
    print("Testing imports...")
    from backend.vector_store import initialize_vector_store
    from backend.question_generator import QuestionGenerator
    
    print("All imports successful!")
    print("Vector store:", initialize_vector_store)
    print("QuestionGenerator:", QuestionGenerator)
    
except ImportError as e:
    print(f"Import error: {e}")
    print("Current Python path:")
    for p in sys.path:
        print(f"  - {p}")
    print("\nCurrent directory:", os.getcwd())
    print("Files in current directory:", os.listdir('.'))
