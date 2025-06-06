import os
import sys
import json
import logging
import datetime
import time
from enum import Enum, auto
from typing import Dict, Any, Optional, List, Type, TYPE_CHECKING
from io import BytesIO

import streamlit as st
import requests
from PIL import Image
from dotenv import load_dotenv

# Configure logging first
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Try to import GradingSystem, but don't fail if it's not available
try:
    logger.info("Attempting to import GradingSystem from grading_system.py")
    from grading_system import GradingSystem
    logger.info("Successfully imported GradingSystem")
    GRADING_SYSTEM_AVAILABLE = True
except ImportError as e:
    # Configure logging again in case the first attempt failed
    logging.basicConfig(level=logging.ERROR)
    logger = logging.getLogger(__name__)
    
    error_msg = f"Could not import GradingSystem: {e}"
    logger.error(error_msg, exc_info=True)
    logger.error(f"Python path: {sys.path}")
    logger.error(f"Current directory: {os.getcwd()}")
    logger.error(f"Files in directory: {os.listdir('.')}")
    
    GRADING_SYSTEM_AVAILABLE = False
    
    # Create a dummy GradingSystem class for type hints
    class GradingSystem:
        def __init__(self):
            error_msg = "GradingSystem is not available. Check your dependencies."
            logger.error(error_msg)
            raise RuntimeError(error_msg)

        def process_submission(self, *args, **kwargs):
            return {
                "success": False,
                "error": "Grading system not available. Check your dependencies.",
                "transcription": "[Not available]",
                "translation": "[Not available]",
                "grade": 0,
                "feedback": "Grading system not available. Please check the logs.",
                "suggestions": ["Install required dependencies and restart the app."]
            }

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Debug: Print current working directory and list files
cwd = os.getcwd()
logger.info(f"Current working directory: {cwd}")
logger.info(f"Files in directory: {os.listdir(cwd)}")

# Load environment variables
env_path = os.path.join(os.path.dirname(__file__), '.env')
logger.info(f"Looking for .env file at: {env_path}")

if os.path.exists(env_path):
    logger.info(".env file found, loading environment variables")
    load_dotenv(env_path)
    
    # Debug: Print all environment variables (without values for security)
    logger.info("Environment variables loaded:")
    for key in os.environ:
        logger.info(f"  {key}: {'*' * 8 if 'key' in key.lower() or 'token' in key.lower() or 'secret' in key.lower() else os.environ[key]}")
else:
    logger.warning("No .env file found")

# Check for GROQ API key
GROQ_API_KEY = os.getenv('GROQ_API_KEY')
logger.info(f"GROQ_API_KEY found: {'Yes' if GROQ_API_KEY else 'No'}")

# Initialize GradingSystem if available and API key is present
if not GRADING_SYSTEM_AVAILABLE or not GROQ_API_KEY:
    warning_msg = ""
    if not GRADING_SYSTEM_AVAILABLE:
        warning_msg = "GradingSystem is not available. "
    if not GROQ_API_KEY:
        warning_msg += "GROQ_API_KEY not found in environment variables. "
    warning_msg += "Grading features will be disabled."
    logger.warning(warning_msg)
    st.warning(warning_msg)

def log_debug(message, data=None):
    """Helper function to log debug messages"""
    if 'debug_logs' not in st.session_state:
        st.session_state.debug_logs = []
        
    timestamp = datetime.datetime.now().strftime("%H:%M:%S")
    entry = f"[{timestamp}] {message}"
    
    if data is not None:
        if isinstance(data, (dict, list)):
            entry += f"\n{json.dumps(data, indent=2, default=str)}"
        else:
            entry += f"\n{str(data)}"
    
    # Print to console for debugging
    print(entry)
    
    st.session_state.debug_logs.append(entry)
    # Keep only last 20 debug messages
    st.session_state.debug_logs = st.session_state.debug_logs[-20:]

API_BASE_URL = "http://localhost:8081/api"
GROUP_ID = 1  # Default group ID, can be made configurable

# Types
class AppState(Enum):
    SETUP = "setup"
    PRACTICE = "practice"
    REVIEW = "review"

class Word:
    def __init__(self, id: int, japanese: str, romaji: str, english: str):
        self.id = id
        self.japanese = japanese
        self.romaji = romaji
        self.english = english

class GradingResult:
    def __init__(self, transcription: str, translation: str, grade: str, feedback: str):
        self.transcription = transcription
        self.translation = translation
        self.grade = grade
        self.feedback = feedback

# Initialize session state with debug info
def init_session_state():
    if 'initialized' not in st.session_state:
        st.session_state.initialized = True
        st.session_state.app_state = 'setup'  # Start in setup state
        st.session_state.words = []
        st.session_state.current_sentence = ""
        st.session_state.debug_logs = []
        st.session_state.uploaded_image = None
        st.session_state.grading_result = None
        st.session_state.grading_system = None
        
        # Initialize grading system if available and API key is present
        if GRADING_SYSTEM_AVAILABLE and GROQ_API_KEY:
            try:
                logger.info("Initializing GradingSystem with API key")
                st.session_state.grading_system = GradingSystem()
                logger.info("GradingSystem initialized successfully")
                log_debug("Grading system initialized successfully")
                
                # Debug: Check if the GradingSystem has a valid client
                if hasattr(st.session_state.grading_system, 'groq_client') and st.session_state.grading_system.groq_client is not None:
                    logger.info("GradingSystem has a valid Groq client")
                else:
                    logger.warning("GradingSystem does not have a valid Groq client")
                    
            except Exception as e:
                error_msg = f"Failed to initialize grading system: {str(e)}"
                logger.error(error_msg, exc_info=True)
                log_debug(error_msg)
                st.session_state.grading_system = None
                
                # Don't show error if it's just about missing API key
                if "API key" not in str(e):
                    st.error("Failed to initialize the grading system. Please check your API keys and logs.")
        
        # Log if grading system is not available
        if st.session_state.grading_system is None:
            logger.info("Grading system is not available or not initialized")
            log_debug("Grading system is not available or not initialized")

# Initialize the session state
init_session_state()

def fetch_words() -> List[Word]:
    """Fetch words from the API and store them in memory."""
    url = f"{API_BASE_URL}/groups/{GROUP_ID}/raw"
    log_debug(f"Fetching words from: {url}")
    
    try:
        log_debug("Sending GET request...")
        response = requests.get(url, timeout=10)
        log_debug(f"Response status code: {response.status_code}")
        
        response.raise_for_status()
        
        data = response.json()
        log_debug(f"Response data type: {type(data)}")
        log_debug(f"Response data keys: {list(data.keys())}")
        
        words = []
        items = data.get("items", [])
        log_debug(f"Found {len(items)} items in response")
        
        if not isinstance(items, list):
            log_debug(f"Unexpected items type: {type(items)}, expected list")
            items = []
        
        for i, item in enumerate(items, 1):
            word = Word(
                id=item.get("id"),
                japanese=item.get("japanese", ""),
                romaji=item.get("romaji", ""),
                english=item.get("english", "")
            )
            words.append(word)
            log_debug(f"Word {i}: {word.japanese} ({word.romaji}) - {word.english}")
        
        log_debug(f"Successfully loaded {len(words)} words")
        return words
        
    except requests.exceptions.RequestException as e:
        st.error(f"Request failed: {str(e)}")
    except json.JSONDecodeError as e:
        st.error(f"Failed to parse JSON response: {str(e)}")
    except Exception as e:
        st.error(f"An unexpected error occurred: {str(e)}")
    
    return []

def generate_sentence(words: List[Word]) -> str:
    """
    Generate a practice sentence using the given words with Groq LLM.
    Creates a simple Japanese sentence at JLPT N5 level using the first word.
    """
    log_debug(f"Generating sentence from {len(words)} words")
    
    if not words:
        log_debug("No words provided for sentence generation")
        return "No words available for sentence generation."
    
    # Get the first word to use in the sentence
    word = words[0]
    log_debug(f"Using word for sentence generation: {word.japanese} ({word.english})")
    
    try:
        # Initialize Groq client
        groq_api_key = os.getenv('GROQ_API_KEY')
        if not groq_api_key:
            raise ValueError("GROQ_API_KEY environment variable not set")
            
        client = Groq(api_key=groq_api_key)
        
        # Create a prompt for the LLM
        prompt = f"""Generate a simple Japanese sentence using the word: {word.japanese} ({word.english}).
        
        Requirements:
        - Use only JLPT N5 level grammar and vocabulary
        - Keep the sentence short and simple (5-10 words)
        - Use the word naturally in the sentence
        - Only output the Japanese sentence, no translation or explanation
        - Use hiragana/katakana for words that would normally be written that way
        - Use kanji for common words that N5 learners should know
        
        Example output for '本 (book)': これは私の本です。
        
        Sentence:"""
        
        # Make the API call
        response = client.chat.completions.create(
            model="llama3-8b-8192",
            messages=[
                {"role": "system", "content": "You are a helpful Japanese language teacher that creates simple, natural Japanese sentences for beginners."},
                {"role": "user", "content": prompt}
            ],
            temperature=0.7,
            max_tokens=100,
            top_p=1,
            stop=None,
        )
        
        # Extract the generated sentence
        if response.choices and len(response.choices) > 0:
            sentence = response.choices[0].message.content.strip()
            # Clean up any extra quotes or whitespace
            sentence = sentence.strip('"\'').strip()
            log_debug(f"Generated sentence: {sentence}")
            return sentence
        else:
            raise ValueError("No response from Groq API")
            
    except Exception as e:
        log_debug(f"Error generating sentence with Groq: {str(e)}")
        # Fallback to a simple sentence using the word
        fallback = f"{word.japanese} は いい です。"  # "[Word] is good."
        log_debug(f"Using fallback sentence: {fallback}")
        return fallback

def setup_state():
    """Render the setup state UI."""
    log_debug("="*50)
    log_debug("RENDERING SETUP STATE")
    log_debug("-"*50)
    
    # Debug current state
    current_state = {
        'app_state': st.session_state.get('app_state'),
        'words_count': len(st.session_state.get('words', [])),
        'current_sentence': st.session_state.get('current_sentence'),
        'has_words': bool(st.session_state.get('words'))
    }
    log_debug(f"Current state: {current_state}")
    
    # Main UI
    st.title("Japanese Writing Practice")
    st.write("Welcome to Japanese Writing Practice!")
    
    # Main content
    st.write("""
    This app helps you practice writing Japanese sentences. 
    Click the button below to get started with a new sentence.
    """)
    
    # Create a form with a unique key
    form_key = "setup_form"
    log_debug(f"Creating form with key: {form_key}")
    
    # Create a single form with a submit button
    with st.form(key=form_key):
        # Add the submit button
        log_debug("Adding submit button to form")
        submit_button = st.form_submit_button("Generate Sentence", type="primary", 
                                            help="Click to generate a new practice sentence")
    
    # Handle form submission
    if submit_button:
        log_debug("Generate Sentence button clicked")
        with st.spinner("Preparing your practice session..."):
            try:
                # Clear any previous state
                for key in ['words', 'current_sentence', 'grading_result']:
                    if key in st.session_state:
                        del st.session_state[key]
                
                # Fetch new words
                log_debug("Fetching words from API...")
                words = fetch_words()
                
                if not words:
                    st.error("Failed to load words. Please try again.")
                    log_debug("No words returned from fetch_words()")
                    st.stop()
                
                log_debug(f"Successfully fetched {len(words)} words")
                
                # Generate a sentence
                log_debug("Generating practice sentence...")
                sentence = generate_sentence(words)
                log_debug(f"Generated sentence: {sentence}")
                
                # Update session state with string values
                st.session_state.words = words
                st.session_state.current_sentence = sentence
                st.session_state.app_state = 'practice'  # Use string value directly
                
                log_debug(f"Transitioning to PRACTICE state with {len(words)} words")
                log_debug(f"Generated sentence: {sentence}")
                log_debug(f"Updated session state: {st.session_state}")
                
                # Force a rerun to show the practice state
                st.rerun()
                
            except Exception as e:
                error_msg = f"Error in setup_state: {str(e)}"
                st.session_state.error = error_msg
                log_debug(error_msg)
                import traceback
                log_debug(f"Traceback: {traceback.format_exc()}")
                st.rerun()
    
    # Display any error message
    if 'error' in st.session_state:
        st.error(st.session_state.error)
        
    log_debug("Setup form rendered, waiting for button click")

def practice_state():
    """Render the practice state UI."""
    log_debug("="*50)
    log_debug("RENDERING PRACTICE STATE")
    log_debug("-"*50)
    
    # Ensure we have the required data
    if not st.session_state.get('words') or not st.session_state.get('current_sentence'):
        st.error("Missing practice data. Please go back to setup.")
        if st.button("Back to Setup"):
            st.session_state.app_state = 'setup'
            st.rerun()
        return
    
    # Display the Japanese sentence to practice writing
    st.write("### Japanese Sentence:")
    st.info(f"{st.session_state.current_sentence}")
    st.write("### Your Task:")
    st.write("Write this sentence in Japanese and upload a photo of your handwriting.")
    
    # Image upload
    uploaded_file = st.file_uploader("Upload your Japanese writing", type=["png", "jpg", "jpeg"])
    
    if uploaded_file is not None:
        # Store the uploaded image
        st.session_state.uploaded_image = uploaded_file
        
        # Display the uploaded image
        image = Image.open(uploaded_file)
        st.image(image, caption="Your writing", use_column_width=True)
    
    # Check if grading system is available
    if not st.session_state.get('grading_system'):
        st.warning("Grading system is not available. Please set GROQ_API_KEY in .env file.")
        return
    
    # Submit button
    if st.button("Submit for Review") and st.session_state.uploaded_image is not None:
        # Check if grading system is available
        if 'grading_system' not in st.session_state or st.session_state.grading_system is None:
            st.error("Grading system is not available. Please check your API key and try again.")
            return
            
        # Process the image and grade the writing
        with st.spinner("Grading your writing..."):
            try:
                # Process the image with grading system
                image = Image.open(st.session_state.uploaded_image)
                
                # Show a warning if we're running without a valid API key
                if not GROQ_API_KEY:
                    st.warning("Running in limited mode. Set GROQ_API_KEY in .env for full functionality.")
                
                # Use the full generated sentence as the expected Japanese
                expected_japanese = st.session_state.current_sentence
                
                # Store the expected Japanese for display
                st.session_state.expected_japanese = expected_japanese
                
                # Process the submission with the expected Japanese
                grading_result = st.session_state.grading_system.process_submission(
                    image, 
                    expected_japanese  # Pass the expected Japanese for direct comparison
                )
                
                # Store the result and transition to review state
                st.session_state.grading_result = grading_result
                st.session_state.app_state = 'review'
                st.rerun()
                
            except Exception as e:
                error_msg = f"Error in grading: {str(e)}"
                log_debug(error_msg, {"error": str(e), "type": type(e).__name__})
                st.error("An error occurred while grading your writing. Please try again or check your API key.")
                
                # Show more detailed error in debug mode
                if st.toggle("Show debug info"):
                    st.error(f"Error details: {str(e)}")
                    st.json({"error": str(e), "type": type(e).__name__})
    
    # Back to setup button
    if st.button("Back to Setup"):
        st.session_state.app_state = 'setup'
        st.rerun()

    # Debug section (collapsible)
    with st.expander("Debug Information"):
        st.write("### Session State")
        st.json({k: str(v) for k, v in st.session_state.items() if k != 'debug_logs'})
        st.write("Words in session:", st.session_state.words)
    
    # Help text
    if uploaded_file is None:
        st.warning("Please upload an image of your handwritten Japanese translation.")

def review_state():
    """Render the review state UI with detailed feedback."""
    log_debug("="*50)
    log_debug("RENDERING REVIEW STATE")
    log_debug("-"*50)
    
    # Debug current state
    current_state = {
        'app_state': st.session_state.get('app_state'),
        'has_grading_result': bool(st.session_state.get('grading_result')),
        'has_current_sentence': bool(st.session_state.get('current_sentence')),
        'has_words': bool(st.session_state.get('words'))
    }
    log_debug(f"Current state: {current_state}")
    
    st.title("Review Your Writing")
    
    if 'grading_result' not in st.session_state or not st.session_state.grading_result:
        st.error("No grading results found. Please submit your writing for review first.")
        if st.button("Back to Practice"):
            st.session_state.app_state = 'practice'
            st.rerun()
        return
    
    result = st.session_state.grading_result
    
    # Display the original Japanese sentence and its English translation
    col_jp, col_en = st.columns(2)
    
    with col_jp:
        st.write("### Japanese:")
        st.info(f"{st.session_state.current_sentence}")
    
    # Get the English translation from the grading result if available
    english_translation = result.get('translation', 'No translation available')
    with col_en:
        st.write("### English Translation:")
        st.info(english_translation)
    
    # Show the grade with color coding
    st.write("### Your Score:")
    grade = result.get('grade', 0)
    if isinstance(grade, (int, float)):
        if grade >= 80:
            st.success(f"**{grade}/100** - Excellent! 🎉")
        elif grade >= 60:
            st.info(f"**{grade}/100** - Good job! 👍")
        else:
            st.warning(f"**{grade}/100** - Keep practicing! 💪")
    else:
        st.write(f"**{grade}**")
    
    # Display the Japanese comparison
    st.write("---")
    st.write("### Japanese Writing Evaluation:")
    
    # Show expected vs recognized Japanese in columns
    col1, col2 = st.columns(2)
    
    with col1:
        st.write("#### Expected Japanese:")
        expected_japanese = st.session_state.get('expected_japanese', '')
        st.code(expected_japanese, language='japanese')
    
    with col2:
        st.write("#### Your Writing (Recognized):")
        japanese_text = result.get('transcription', result.get('japanese_text', ''))
        if japanese_text and not japanese_text.startswith('['):
            # Highlight differences if possible
            st.code(japanese_text, language='japanese')
            
            # Show character-by-character comparison if same length
            if len(expected_japanese) == len(japanese_text):
                st.write("Character comparison:")
                comparison_text = ""
                for exp, got in zip(expected_japanese, japanese_text):
                    if exp == got:
                        comparison_text += f"`{exp}` "
                    else:
                        comparison_text += f"<span style='color:red'>`{got}`</span> "
                st.markdown(comparison_text, unsafe_allow_html=True)
                st.write("")
        else:
            st.warning("⚠️ Could not recognize any text in the image. Please try again with clearer handwriting.")
    
    # Display feedback if available
    feedback = result.get('feedback', '')
    if feedback and not feedback.startswith('Error'):
        st.write("---")
        st.write("### Feedback:")
        st.info(feedback)
    
    # Display suggestions if available
    suggestions = result.get('suggestions', [])
    if suggestions and (not isinstance(suggestions, str) or not suggestions.startswith('Error')):
        st.write("### Suggestions for Improvement:")
        if isinstance(suggestions, str):
            st.info(suggestions)
        elif isinstance(suggestions, list) and suggestions:
            for i, suggestion in enumerate(suggestions, 1):
                st.write(f"{i}. {suggestion}")
    
    # Display the grade with color coding
    st.write("### Grade:")
    grade = result.get('grade', 0)
    if isinstance(grade, (int, float)):
        if grade >= 80:
            st.success(f"**{grade}/100** - Excellent!")
        elif grade >= 60:
            st.info(f"**{grade}/100** - Good job!")
        else:
            st.warning(f"**{grade}/100** - Keep practicing!")
    else:
        st.write(f"**{grade}**")
    
    # Display feedback if available
    explanation = result.get('explanation', '')
    if explanation and not explanation.startswith('Error'):
        st.write("### Feedback:")
        st.write(explanation)
    
    # Display suggestions if available
    suggestions = result.get('suggestions', [])
    if suggestions and (not isinstance(suggestions, str) or not suggestions.startswith('Error')):
        st.write("### Suggestions for Improvement:")
        if isinstance(suggestions, str):
            st.write(suggestions)
        elif isinstance(suggestions, list) and suggestions:
            for i, suggestion in enumerate(suggestions, 1):
                st.write(f"{i}. {suggestion}")
    
    # Display raw MangaOCR output in an expandable section
    with st.expander("View Raw OCR Output"):
        st.write("### Raw MangaOCR Output")
        
        # Show the raw transcription
        raw_transcription = result.get('transcription', 'No transcription available')
        st.code(f"Transcription: {raw_transcription}", language='text')
        
        # Show transcription metadata if available
        if 'transcription_metadata' in result:
            meta = result['transcription_metadata']
            st.write("#### Transcription Metadata")
            st.json({
                "Processing Time (seconds)": f"{meta.get('processing_time_seconds', 0):.2f}",
                "Characters Recognized": meta.get('characters_recognized', 0),
                "Characters Per Second": f"{meta.get('characters_recognized', 0) / meta.get('processing_time_seconds', 1):.1f}" if meta.get('processing_time_seconds', 0) > 0 else "N/A"
            })
        
        # Show any warnings if present
        if 'warnings' in result and result['warnings']:
            st.warning("### Warnings")
            for warning in result['warnings']:
                st.warning(f"⚠️ {warning}")
        
        # Show the raw result for debugging
        if st.checkbox("Show raw result data (for debugging)"):
            st.write("### Raw Result Data")
            st.json({k: v for k, v in result.items() if k not in ['transcription_metadata', 'warnings']})
            
        # Add a button to copy the transcription
        if st.button("📋 Copy Transcription to Clipboard"):
            st.session_state.clipboard = raw_transcription
            st.toast("Transcription copied to clipboard!")
        
        if 'clipboard' in st.session_state:
            st.text_area("Clipboard (editable)", st.session_state.clipboard, key="clipboard_editor")
    
    # Display the uploaded image if available
    if 'uploaded_image' in st.session_state and st.session_state.uploaded_image is not None:
        st.write("### Your Writing:")
        try:
            image = Image.open(st.session_state.uploaded_image)
            st.image(image, use_column_width=True, caption="Your handwritten Japanese")
        except Exception as e:
            log_debug(f"Error displaying image: {str(e)}")
    
    # Navigation buttons with clear actions
    st.markdown("---")
    st.markdown("### What would you like to do next?")
    
    col1, col2, col3 = st.columns(3)
    
    with col1:
        if st.button("⟲ Try Again", key="try_again_btn", use_container_width=True):
            st.session_state.app_state = "practice"
            st.rerun()
    
    with col2:
        if st.button("🔄 New Sentence", key="new_sentence_btn", use_container_width=True):
            # Keep the same word list but generate a new sentence
            if 'current_sentence' in st.session_state:
                del st.session_state.current_sentence
            if 'uploaded_image' in st.session_state:
                st.session_state.uploaded_image = None
            if 'grading_result' in st.session_state:
                del st.session_state.grading_result
            st.session_state.app_state = "practice"
            st.rerun()
    
    with col3:
        if st.button("🏠 New Practice Session", key="new_session_btn", type="primary", use_container_width=True):
            # Clear everything and start over
            for key in ['current_sentence', 'words', 'uploaded_image', 'grading_result']:
                if key in st.session_state:
                    del st.session_state[key]
            st.session_state.app_state = "setup"
            st.rerun()
    
    # Debug information (collapsed by default)
    with st.expander("🔍 Debug Information"):
        st.json({
            'app_state': st.session_state.app_state,
            'has_grading_result': bool(st.session_state.get('grading_result')),
            'has_current_sentence': bool(st.session_state.get('current_sentence')),
            'has_words': bool(st.session_state.get('words')),
            'has_uploaded_image': 'uploaded_image' in st.session_state,
            'grading_result_keys': list(st.session_state.grading_result.keys()) if st.session_state.get('grading_result') else None
        })

def test_backend_endpoint() -> dict:
    """Test the backend endpoint and return detailed results."""
    url = f"{API_BASE_URL}/groups/{GROUP_ID}/raw"
    result = {
        'success': False,
        'url': url,
        'status_code': None,
        'error': None,
        'response': None,
        'word_count': 0
    }
    
    try:
        log_debug(f"Testing backend endpoint: {url}")
        response = requests.get(url, timeout=10)
        result['status_code'] = response.status_code
        
        if response.status_code == 200:
            data = response.json()
            result['response'] = data
            words = data.get('items', [])
            result['word_count'] = len(words)
            result['success'] = True
            log_debug(f"Successfully retrieved {len(words)} words from backend")
        else:
            result['error'] = f"Unexpected status code: {response.status_code}"
            log_debug(f"Backend returned status {response.status_code}: {response.text}")
            
    except requests.exceptions.RequestException as e:
        result['error'] = str(e)
        log_debug(f"Request failed: {str(e)}")
    except json.JSONDecodeError as e:
        result['error'] = f"Invalid JSON response: {str(e)}"
        log_debug(f"Failed to parse JSON response: {str(e)}")
    except Exception as e:
        result['error'] = f"Unexpected error: {str(e)}"
        log_debug(f"Unexpected error: {str(e)}")
    
    return result

def check_backend_connection() -> bool:
    """Check if the backend API is accessible and return detailed status."""
    log_debug(f"Checking backend connection to: {API_BASE_URL}")
    
    # First check the health endpoint
    try:
        health_url = f"{API_BASE_URL}/health"
        log_debug(f"Checking health endpoint: {health_url}")
        response = requests.get(health_url, timeout=5)
        log_debug(f"Health check status: {response.status_code}")
        
        if response.status_code == 200:
            log_debug("Backend health check passed")
            return True
            
        log_debug(f"Health check failed with status {response.status_code}: {response.text}")
        
    except requests.exceptions.RequestException as e:
        log_debug(f"Health check request failed: {str(e)}")
    
    # If health check fails, try the actual endpoint
    log_debug("Health check failed, trying the words endpoint directly")
    result = test_backend_endpoint()
    return result['success']

# Main app
def render_sidebar():
    """Render the sidebar with debug information"""
    with st.sidebar:
        st.write("### Debug Panel")
        
        # App state info
        st.write("#### App State")
        st.json({
            'app_state': st.session_state.get('app_state', 'N/A'),
            'words_count': len(st.session_state.get('words', [])),
            'current_sentence': st.session_state.get('current_sentence', 'N/A'),
            'has_grading_result': 'grading_result' in st.session_state
        })
        
        # Debug log
        st.write("#### Debug Log")
        debug_container = st.container(height=300)
        with debug_container:
            for log in reversed(st.session_state.get('debug', [])[-10:]):
                st.code(log, language="text")
        
        # Reset button
        if st.button("🔄 Reset App", use_container_width=True):
            for key in list(st.session_state.keys()):
                del st.session_state[key]
            st.rerun()

def main():
    # Set page config - must be the first Streamlit command
    st.set_page_config(
        page_title="Japanese Writing Practice",
        page_icon="✍️",
        layout="centered",
        initial_sidebar_state="expanded"
    )
    
    # Add some custom CSS
    st.markdown("""
    <style>
    .debug-info {
        font-family: monospace;
        font-size: 0.85em;
        background-color: #f8f9fa;
        padding: 0.5rem;
        border-radius: 0.5rem;
        margin: 0.5rem 0;
        max-height: 300px;
        overflow-y: auto;
    }
    .debug-log {
        font-family: monospace;
        font-size: 0.8em;
        white-space: pre-wrap;
        background-color: #f1f3f5;
        padding: 0.5rem;
        border-radius: 0.25rem;
        margin: 0.25rem 0;
    }
    </style>
    """, unsafe_allow_html=True)
    
    try:
        # Initialize session state if not already done
        if 'initialized' not in st.session_state:
            log_debug("Initializing session state for the first time")
            init_session_state()
        
        log_debug(f"Current app state: {st.session_state.get('app_state')}")
        
        # Check backend connection
        if not check_backend_connection():
            st.error("Could not connect to the backend. Please check if the backend is running.")
            st.stop()
        
        # Always render the sidebar
        try:
            render_sidebar()
        except Exception as e:
            log_debug(f"Error rendering sidebar: {str(e)}")
        
        # Debug session state before rendering
        log_debug(f"Session state before rendering: {st.session_state}")
        
        # Get current state and normalize it
        current_state_value = st.session_state.get('app_state', 'setup')
        
        # If it's an AppState enum, get its value
        if isinstance(current_state_value, AppState):
            current_state_value = current_state_value.value
            st.session_state.app_state = current_state_value  # Update to store string value
            
        # Convert to string and normalize case for comparison
        current_state_str = str(current_state_value).lower().strip()
        
        # Map to AppState enum
        try:
            # Find the enum member that matches the string value (case-insensitive)
            current_state = next(
                state for state in AppState 
                if state.value.lower() == current_state_str
            )
        except StopIteration:
            log_debug(f"Invalid app state: {current_state_str}, defaulting to SETUP")
            current_state = AppState.SETUP
            st.session_state.app_state = current_state.value
            st.rerun()
        
        log_debug(f"Rendering state: {current_state}")
        
        # Debug log the state transition
        log_debug(f"Transitioning to state: {current_state}")
        
        # Render the appropriate state
        if current_state == AppState.SETUP:
            log_debug("Calling setup_state()")
            setup_state()
        elif current_state == AppState.PRACTICE:
            log_debug("Calling practice_state()")
            practice_state()
        elif current_state == AppState.REVIEW:
            log_debug("Calling review_state()")
            review_state()
        else:
            log_debug(f"Unknown state: {current_state}, defaulting to SETUP")
            st.session_state.app_state = AppState.SETUP.value
            st.rerun()
        
        log_debug("Rendering complete")
        
    except Exception as e:
        st.error(f"An error occurred: {str(e)}")
        log_debug(f"Error in main app: {str(e)}")
        import traceback
        log_debug(f"Traceback: {traceback.format_exc()}")
        
        # Add a button to reset the app
        if st.button("🔄 Reset App"):
            for key in list(st.session_state.keys()):
                del st.session_state[key]
            st.rerun()

if __name__ == "__main__":
    main()