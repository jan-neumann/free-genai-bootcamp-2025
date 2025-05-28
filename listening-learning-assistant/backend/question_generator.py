import os
import sys
from pathlib import Path
import re
from typing import List, Dict, Optional
from groq import Groq
from dotenv import load_dotenv

# Use absolute import
from backend.vector_store import initialize_vector_store

# Load environment variables
load_dotenv()

class QuestionGenerator:
    def __init__(self, model_name: str = "qwen-qwq-32b"):
        """Initialize the question generator with Groq client and vector store"""
        api_key = os.getenv("GROQ_API_KEY")
        if not api_key:
            raise ValueError("GROQ_API_KEY environment variable not set")
            
        self.client = Groq(api_key=api_key)
        self.model_name = model_name
        # Initialize vector store
        self.vector_store = None
        try:
            self.vector_store = initialize_vector_store()
        except Exception as e:
            print(f"Warning: Could not initialize vector store: {str(e)}")
    
    def generate_question(self, question_type: str, topic: str = None) -> Dict:
        """
        Generate a question using RAG with vector store context
        
        Args:
            question_type: Type of question to generate (e.g., 'Dialogue', 'Vocabulary', 'Listening')
            topic: Optional topic to guide question generation
            
        Returns:
            Dict containing question details or None if generation fails
        """
        try:
            # First, find similar questions from the vector store
            try:
                query = f"{question_type} question about {topic}" if topic else question_type
                similar_questions = self.vector_store.find_similar_questions(query, n_results=3)
                
                # Prepare context from similar questions
                context = "\n".join([q.get('text', '') for q in similar_questions])
            except Exception as e:
                print(f"Error getting similar questions: {str(e)}")
                context = "No similar questions found."
            
            # Generate new question using RAG context
            prompt = f"""
            Create a JLPT N5 level question in Japanese with this exact format:
            
            [Situation]
            [1-2 lines of context in Japanese]
            
            [Conversation]
            [2-4 lines of natural Japanese dialogue]
            
            [Question]
            [1 question in Japanese]
            
            Example:
            [Situation]
            男の人と女の人が話しています。
            
            [Conversation]
            女：今朝のニュース見た？
            男：ううん、まだ見てないよ。何かあったの？
            女：駅の近くに新しいレストランができたって。
            
            [Question]
            二人は何について話していますか？
            
            Create a new question about: {topic if topic else 'everyday situations'}"""
            
            if topic:
                prompt += f"\n\nMake the question related to: {topic}"
                
            try:
                response = self.client.chat.completions.create(
                    model=self.model_name,
                    messages=[
                        {"role": "system", "content": "You are a helpful Japanese language teaching assistant for JLPT N5 level."},
                        {"role": "user", "content": prompt}
                    ],
                    temperature=0.7,
                    max_tokens=500
                )
            except Exception as e:
                print(f"Error calling Groq API: {str(e)}")
                return None
                
        except Exception as e:
            print(f"Error in question generation setup: {str(e)}")
            return None
        
        try:
            response = self.client.chat.completions.create(
                model=self.model_name,
                messages=[
                    {"role": "system", "content": "You are a helpful Japanese language teaching assistant for JLPT N5 level."},
                    {"role": "user", "content": prompt}
                ],
                temperature=0.7,
                max_tokens=500
            )
            
            # Parse the response
            response_text = response.choices[0].message.content
            
            # Try to parse the response in the expected format
            parts = {}
            current_section = None
            
            # Map of possible section headers (case-insensitive)
            section_mapping = {
                'situation': 'Situation',
                'conversation': 'Conversation',
                'question': 'Question',
                '状況': 'Situation',
                '会話': 'Conversation',
                '質問': 'Question'
            }
            
            for line in response_text.split('\n'):
                line = line.strip()
                if not line:
                    continue
                    
                # Check for section headers (case-insensitive)
                if line.startswith('[') and line.endswith(']'):
                    section_name = line[1:-1].strip().lower()
                    current_section = section_mapping.get(section_name, section_name)
                    parts[current_section] = []
                elif current_section:
                    parts[current_section].append(line)
            
            # Rebuild the response in the correct format with English headers
            response_text = ''
            for section in ['Situation', 'Conversation', 'Question']:
                if section in parts and parts[section]:
                    response_text += f"{section}:\n"
                    response_text += '\n'.join(parts[section]) + '\n\n'
            
            # If parsing failed, return the original text but clean it up
            if not response_text:
                response_text = re.sub(r'^[\s\[\]]+', '', response_text)
                response_text = '\n'.join(line.strip() for line in response_text.split('\n') if line.strip())
            
            # Debug: Print the processed response for troubleshooting
            print("\n--- Processed Response from API ---")
            print(response_text)
            print("--- End of Processed Response ---\n")
            
            # Try to parse the response
            parsed_data = None
            
            # Strategy 1: Look for explicit section headers
            intro_match = re.search(r'(?i)introduction[：:](.*?)(?=\n\n|$)', response_text, re.DOTALL)
            conversation_match = re.search(r'(?i)conversation[：:](.*?)(?=\n\nquestion|$)', response_text, re.DOTALL)
            question_match = re.search(r'(?i)question[：:](.*?)(?=\n|$)', response_text, re.DOTALL)
            
            if all([intro_match, conversation_match, question_match]):
                intro = intro_match.group(1).strip()
                conversation = conversation_match.group(1).strip()
                question_text = question_match.group(1).strip()
                parsed_data = (intro, conversation, question_text)
            
            # Strategy 2: Look for line break patterns
            if not parsed_data:
                parts = [p.strip() for p in re.split(r'\n{2,}', response_text) if p.strip()]
                if len(parts) >= 3:
                    # First non-empty part is intro
                    intro = parts[0].replace('Introduction:', '').replace('Introduction：', '').strip()
                    # Last part is question
                    question_text = parts[-1].replace('Question:', '').replace('Question：', '').strip()
                    # Everything in between is conversation
                    conversation = '\n'.join(parts[1:-1])
                    conversation = conversation.replace('Conversation:', '').replace('Conversation：', '').strip()
                    parsed_data = (intro, conversation, question_text)
            
            # Strategy 3: Fallback to line-based parsing
            if not parsed_data and '\n' in response_text:
                lines = [line.strip() for line in response_text.split('\n') if line.strip()]
                if len(lines) >= 4:  # At least 1 intro + 2 convo + 1 question
                    intro = lines[0].replace('Introduction:', '').replace('Introduction：', '').strip()
                    conversation = '\n'.join(lines[1:-1])
                    question_text = lines[-1].replace('Question:', '').replace('Question：', '').strip()
                    parsed_data = (intro, conversation, question_text)
            
            if not parsed_data:
                print("Error: Could not parse question format from response")
                print("Response content:", response_text)
                return None
                
            intro, conversation, question_text = parsed_data
            
            # Remove content within <think> tags and clean up text
            def clean_text(text):
                # Remove think tags first
                text = re.sub(r'<think>.*?</think>', '', text, flags=re.DOTALL)
                # Remove any remaining HTML-like tags
                text = re.sub(r'<[^>]+>', '', text)
                # Remove leading colons/spaces and extra whitespace
                text = re.sub(r'^[：:\s]+', '', text)
                text = ' '.join(text.split())  # Normalize whitespace
                return text.strip()
            
            # Clean up all text parts
            intro = clean_text(intro)
            conversation = clean_text(conversation)
            question_text = clean_text(question_text)
            
            # Ensure conversation has proper line breaks
            conversation = '\n'.join(line.strip() for line in conversation.split('\n') if line.strip())
            
            # Format the full question text for display
            formatted_question = f"{intro}\n\n{conversation}\n\n{question_text}"
            
            # For now, we'll return the formatted question directly
            # In a real implementation, you might want to generate multiple choice options
            # based on the question content
            return {
                'question': formatted_question,
                'options': ["A) はい", "B) いいえ", "C) わかりません", "D) もう一度お願いします"],
                'correct_answer': 0,  # Default to first option
                'explanation': "",
                'raw_response': response_text
            }
            
        except Exception as e:
            print(f"Error generating question: {str(e)}")
            import traceback
            traceback.print_exc()
            return None

# Singleton instance
question_generator = QuestionGenerator()