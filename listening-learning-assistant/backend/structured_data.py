import re
from typing import Optional, Dict, List, Any
import os
import sys
from pathlib import Path
from groq import Groq
from dotenv import load_dotenv

# Suppress tokenizers parallelism warning
os.environ["TOKENIZERS_PARALLELISM"] = "false"

# Load environment variables
load_dotenv()

class TranscriptStructurer:
    def __init__(self, model_name: str = "qwen-qwq-32b"):
        """Initialize Groq client with Qwen model"""
        api_key = os.getenv("GROQ_API_KEY")
        if not api_key:
            raise ValueError("GROQ_API_KEY environment variable not set")
            
        self.client = Groq(api_key=api_key)
        self.model_name = model_name
        self.prompts = {
            1: """Extract questions from section 問題1 of this JLPT transcript where the answer can be determined solely from the conversation without needing visual aids.
            
            ONLY include questions that meet these criteria:
            - The answer can be determined purely from the spoken dialogue
            - No spatial/visual information is needed (like locations, layouts, or physical appearances)
            - No physical objects or visual choices need to be compared
            
            For example, INCLUDE questions about:
            - Times and dates
            - Numbers and quantities
            - Spoken choices or decisions
            - Clear verbal directions
            
            DO NOT include questions about:
            - Physical locations that need a map or diagram
            - Visual choices between objects
            - Spatial arrangements or layouts
            - Physical appearances of people or things

            Format each question exactly like this:

            <question>
            Introduction:
            [the situation setup in japanese]
            
            Conversation:
            [the dialogue in japanese]
            
            Question:
            [the question being asked in japanese]

            Options:
            1. [first option in japanese]
            2. [second option in japanese]
            3. [third option in japanese]
            4. [fourth option in japanese]
            </question>

            Rules:
            - Only extract questions from the 問題1 section
            - Only include questions where answers can be determined from dialogue alone
            - Ignore any practice examples (marked with 例)
            - Do not translate any Japanese text
            - Do not include any section descriptions or other text
            - Output questions one after another with no extra text between them
            """,
            
            2: """Extract questions from section 問題2 of this JLPT transcript where the answer can be determined solely from the conversation without needing visual aids.
            
            ONLY include questions that meet these criteria:
            - The answer can be determined purely from the spoken dialogue
            - No spatial/visual information is needed (like locations, layouts, or physical appearances)
            - No physical objects or visual choices need to be compared
            
            For example, INCLUDE questions about:
            - Times and dates
            - Numbers and quantities
            - Spoken choices or decisions
            - Clear verbal directions
            
            DO NOT include questions about:
            - Physical locations that need a map or diagram
            - Visual choices between objects
            - Spatial arrangements or layouts
            - Physical appearances of people or things

            Format each question exactly like this:

            <question>
            Introduction:
            [the situation setup in japanese]
            
            Conversation:
            [the dialogue in japanese]
            
            Question:
            [the question being asked in japanese]
            </question>

            Rules:
            - Only extract questions from the 問題2 section
            - Only include questions where answers can be determined from dialogue alone
            - Ignore any practice examples (marked with 例)
            - Do not translate any Japanese text
            - Do not include any section descriptions or other text
            - Output questions one after another with no extra text between them
            """,
            
            3: """Extract all questions from section 問題3 of this JLPT transcript.
            Format each question exactly like this:

            <question>
            Situation:
            [the situation in japanese where a phrase is needed]
            
            Question:
            何と言いますか
            </question>

            Rules:
            - Only extract questions from the 問題3 section
            - Ignore any practice examples (marked with 例)
            - Do not translate any Japanese text
            - Do not include any section descriptions or other text
            - Output questions one after another with no extra text between them
            """
        }

    def _invoke_groq(self, prompt: str, transcript: str) -> Optional[str]:
        """Make a single call to Groq with the given prompt"""
        full_prompt = f"{prompt}\n\nHere's the transcript:\n{transcript}"
        
        try:
            response = self.client.chat.completions.create(
                model=self.model_name,
                messages=[
                    {"role": "system", "content": "You are a helpful assistant that processes JLPT test transcripts."},
                    {"role": "user", "content": full_prompt}
                ],
                temperature=0.3
            )
            return response.choices[0].message.content
        except Exception as e:
            print(f"Error invoking Groq: {str(e)}")
            return None

    def _clean_content(self, content: str) -> str:
        """Clean the content by removing think tags and fixing any formatting issues"""
        if not content:
            return ""
            
        # First, remove all think tags and their contents
        import re
        cleaned = re.sub(r'<think>.*?</think>', '', content, flags=re.DOTALL)
        
        # Clean up any double newlines that might have been created
        cleaned = re.sub(r'\n{3,}', '\n\n', cleaned)
        
        # Remove any leading/trailing whitespace
        return cleaned.strip()

    def structure_transcript(self, transcript: str) -> Dict[int, str]:
        """Structure the transcript into sections, skipping section 1"""
        results = {}
        # Only process sections 2 and 3
        for section_num in [2, 3]:
            if section_num not in self.prompts:
                print(f"Warning: No prompt defined for section {section_num}")
                continue
                
            print(f"Processing section {section_num}...")
            result = self._invoke_groq(self.prompts[section_num], transcript)
            if result:
                # Clean the content by removing <think> tags
                cleaned_result = self._clean_content(result)
                results[section_num] = cleaned_result
                print(f"Section {section_num} processed successfully")
            else:
                print(f"Failed to process section {section_num}")
        return results

    def save_questions(self, structured_sections: Dict[int, str], base_filename: str) -> bool:
        """Save structured sections to files.
        
        Args:
            structured_sections: Dictionary mapping section numbers to their content
            base_filename: Base path for output files
            
        Returns:
            bool: True if successful, False otherwise
        """
        try:
            # Create questions directory if it doesn't exist
            os.makedirs(os.path.dirname(base_filename), exist_ok=True)
            
            # Process each section
            for section_num, content in structured_sections.items():
                # Save to file
                filename = f"{os.path.splitext(base_filename)[0]}_section{section_num}.txt"
                with open(filename, 'w', encoding='utf-8') as f:
                    f.write(content)
                print(f"Saved section {section_num} to {filename}")
                
            return True
            
        except Exception as e:
            print(f"Error saving questions: {str(e)}")
            return False
            
    def parse_questions_from_content(self, content: str, section: int) -> List[Dict[str, Any]]:
        """Parse questions from section content.
        
        Args:
            content: The content to parse questions from
            section: The section number
            
        Returns:
            List of dictionaries containing question text and metadata
        """
        questions = []
        # Split by question blocks
        question_blocks = re.split(r'<question>', content)
        
        for block in question_blocks[1:]:  # Skip first empty block
            if not block.strip():
                continue
                
            # Extract situation and question
            parts = re.split(r'</?question>', block, flags=re.DOTALL)
            if len(parts) < 2:
                continue
                
            question_text = parts[0].strip()
            questions.append({
                'text': question_text,
                'metadata': {
                    'section': section,
                    'source': 'transcript_processing'
                }
            })
            
        return questions

    def load_transcript(self, filepath: str) -> Optional[str]:
        """Load transcript from file"""
        import os
        # Convert to absolute path
        abs_path = os.path.abspath(filepath)
        print(f"Attempting to load transcript from: {abs_path}")
        
        # Check if file exists
        if not os.path.exists(abs_path):
            print(f"File does not exist at: {abs_path}")
            print(f"Current working directory: {os.getcwd()}")
            return None
            
        try:
            with open(abs_path, 'r', encoding='utf-8') as f:
                content = f.read()
                print(f"Successfully loaded {len(content)} characters from transcript")
                return content
        except Exception as e:
            print(f"Error loading transcript: {str(e)}")
            return None

def process_transcript(transcript_path: str, output_dir: str = "questions") -> Dict[int, str]:
    """Process a transcript file and save results to files.
    
    Args:
        transcript_path: Path to the transcript file
        output_dir: Directory to save the structured output
        
    Returns:
        Dictionary mapping section numbers to their content, or empty dict on failure
    """
    structurer = TranscriptStructurer()
    
    # Ensure output directory exists
    os.makedirs(output_dir, exist_ok=True)
    
    # Generate output path
    base_name = os.path.splitext(os.path.basename(transcript_path))[0]
    output_path = os.path.join(output_dir, f"{base_name}.txt")
    
    print(f"Processing transcript: {transcript_path}")
    transcript = structurer.load_transcript(transcript_path)
    
    if not transcript:
        print("Failed to load transcript")
        return {}
    
    print("Structuring transcript...")
    structured_sections = structurer.structure_transcript(transcript)
    
    if not structured_sections:
        print("No sections were processed successfully")
        return {}
    
    print("Saving questions...")
    if structurer.save_questions(structured_sections, output_path):
        print(f"Successfully processed {len(structured_sections)} sections")
    else:
        print("Failed to save questions")
    
    return structured_sections

if __name__ == "__main__":
    # Process a single transcript
    import sys
    
    if len(sys.argv) > 1:
        transcript_path = sys.argv[1]
    else:
        transcript_path = "transcripts/2zr8KZb1DUs.txt"
    
    process_transcript(transcript_path)