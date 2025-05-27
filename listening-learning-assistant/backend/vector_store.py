import os
import chromadb
from typing import List, Dict, Optional, Any
from chromadb.config import Settings
from chromadb.utils import embedding_functions
from dotenv import load_dotenv
import re

# Load environment variables
load_dotenv()

# Suppress tokenizers parallelism warning
os.environ["TOKENIZERS_PARALLELISM"] = "false"

class QuestionVectorStore:
    """A vector store for managing JLPT listening comprehension questions."""
    
    def __init__(self, persist_directory: str = "chroma_db"):
        """Initialize the vector store with ChromaDB.
        
        Args:
            persist_directory: Directory to store the ChromaDB data
        """
        self.persist_directory = persist_directory
        os.makedirs(persist_directory, exist_ok=True)
        
        # Initialize Chroma client with persistent storage
        self.client = chromadb.PersistentClient(
            path=persist_directory,
            settings=Settings(anonymized_telemetry=False)
        )
        
        # Get or create the collection
        self.collection = self.client.get_or_create_collection(
            name="jplt_questions",
            embedding_function=embedding_functions.DefaultEmbeddingFunction()
        )
    
    def add_question(self, question_text: str, metadata: Optional[Dict] = None) -> str:
        """Add a single question to the vector store.
        
        Args:
            question_text: The text of the question
            metadata: Additional metadata (e.g., difficulty, section, source)
            
        Returns:
            str: The ID of the added question
        """
        if metadata is None:
            metadata = {}
            
        # Generate an ID based on the question text
        question_id = f"q{abs(hash(question_text))}"
        
        # Add to collection
        self.collection.add(
            documents=[question_text],
            metadatas=[metadata],
            ids=[question_id]
        )
        
        return question_id
    
    def add_questions_from_file(self, file_path: str, metadata: Optional[Dict] = None) -> List[str]:
        """Add questions from a text file.
        
        Args:
            file_path: Path to the file containing questions
            metadata: Optional metadata to add to all questions
            
        Returns:
            List[str]: List of question IDs that were added
        """
        if metadata is None:
            metadata = {}
            
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()
            
            # Default parsing for our standard format
            if '<question>' in content:
                return self._parse_standard_questions(content, file_path, metadata)
            # One question per line
            else:
                return self._parse_simple_questions(content, file_path, metadata)
                
        except Exception as e:
            print(f"Error adding questions from {file_path}: {str(e)}")
            return []
    
    def _parse_standard_questions(self, content: str, source: str, base_metadata: Dict) -> List[str]:
        """Parse questions in the standard format with <question> tags"""
        question_pattern = r'<question>(.*?)</question>'
        matches = re.findall(question_pattern, content, re.DOTALL)
        
        question_ids = []
        for i, question_text in enumerate(matches, 1):
            question_text = question_text.strip()
            if not question_text:
                continue
                
            metadata = base_metadata.copy()
            metadata.update({
                'question_number': i,
                'source': source,
                'format': 'standard'
            })
            
            qid = self.add_question(question_text, metadata)
            question_ids.append(qid)
            
        return question_ids
    
    def _parse_simple_questions(self, content: str, source: str, base_metadata: Dict) -> List[str]:
        """Parse questions where each line is a separate question"""
        question_ids = []
        metadata = base_metadata.copy()
        metadata.update({
            'source': source,
            'format': 'simple'
        })
        
        for i, line in enumerate(content.split('\n'), 1):
            line = line.strip()
            if not line or line.startswith('#'):
                continue
                
            question_metadata = metadata.copy()
            question_metadata['line_number'] = i
            
            qid = self.add_question(line, question_metadata)
            question_ids.append(qid)
            
        return question_ids
    
    def find_similar_questions(self, query: str, n_results: int = 5) -> List[Dict]:
        """Find questions similar to the query.
        
        Args:
            query: The search query
            n_results: Number of similar questions to return
            
        Returns:
            List[Dict]: List of similar questions with metadata
        """
        results = self.collection.query(
            query_texts=[query],
            n_results=min(n_results, 10)  # Limit to 10 results max
        )
        
        # Format the results
        similar_questions = []
        for i in range(len(results['ids'][0])):
            similar_questions.append({
                'id': results['ids'][0][i],
                'text': results['documents'][0][i],
                'metadata': results['metadatas'][0][i],
                'distance': results['distances'][0][i]
            })
            
        return similar_questions
    
    def get_question(self, question_id: str) -> Optional[Dict]:
        """Retrieve a question by its ID.
        
        Args:
            question_id: The ID of the question to retrieve
            
        Returns:
            Optional[Dict]: The question data or None if not found
        """
        try:
            result = self.collection.get(ids=[question_id])
            if result['ids']:
                return {
                    'id': result['ids'][0],
                    'text': result['documents'][0],
                    'metadata': result['metadatas'][0]
                }
        except Exception as e:
            print(f"Error retrieving question {question_id}: {str(e)}")
        return None
    
    def get_all_questions(self) -> List[Dict]:
        """Retrieve all questions in the vector store.
        
        Returns:
            List[Dict]: List of all questions with metadata
        """
        try:
            result = self.collection.get()
            questions = []
            for i in range(len(result['ids'])):
                questions.append({
                    'id': result['ids'][i],
                    'text': result['documents'][i],
                    'metadata': result['metadatas'][i]
                })
            return questions
        except Exception as e:
            print(f"Error retrieving all questions: {str(e)}")
            return []


def initialize_vector_store(persist_directory: str = None) -> 'QuestionVectorStore':
    """Initialize and return a QuestionVectorStore instance.
    
    Args:
        persist_directory: Directory to store the vector store. If None, uses 'chroma_db' in the current directory.
    """
    if persist_directory is None:
        persist_directory = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'chroma_db')
    return QuestionVectorStore(persist_directory=persist_directory)


def main():
    import argparse
    
    parser = argparse.ArgumentParser(description='Manage question vector store')
    subparsers = parser.add_subparsers(dest='command')
    
    # Add questions from file
    add_parser = subparsers.add_parser('add', help='Add questions from a file')
    add_parser.add_argument('file', help='Path to the file containing questions')
    add_parser.add_argument('--section', type=int, help='Section number')
    add_parser.add_argument('--source', help='Source identifier')
    
    # Search for similar questions
    search_parser = subparsers.add_parser('search', help='Search for similar questions')
    search_parser.add_argument('query', help='Search query')
    search_parser.add_argument('-n', '--num-results', type=int, default=5, help='Number of results to return')
    
    # List all questions
    list_parser = subparsers.add_parser('list', help='List all questions')
    
    args = parser.parse_args()
    
    vector_store = initialize_vector_store()
    
    if args.command == 'add':
        metadata = {}
        if args.section:
            metadata['section'] = args.section
        if args.source:
            metadata['source'] = args.source
            
        question_ids = vector_store.add_questions_from_file(args.file, metadata)
        print(f"Added {len(question_ids)} questions to the vector store")
        
    elif args.command == 'search':
        results = vector_store.find_similar_questions(args.query, n_results=args.num_results)
        print(f"Found {len(results)} similar questions:")
        for i, q in enumerate(results, 1):
            print(f"\n{i}. Similarity: {1 - q['distance']:.2f}")
            print(f"Question: {q['text']}")
            print(f"Metadata: {q['metadata']}")
            
    elif args.command == 'list':
        questions = vector_store.get_all_questions()
        print(f"Found {len(questions)} questions in the vector store:")
        for i, q in enumerate(questions, 1):
            print(f"\n{i}. {q['text']}")
            print(f"   ID: {q['id']}")
            print(f"   Metadata: {q['metadata']}")
    else:
        parser.print_help()


if __name__ == "__main__":
    main()
