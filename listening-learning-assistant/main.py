import streamlit as st
from dotenv import load_dotenv

load_dotenv() # Load environment variables from .env file
from typing import Dict
import json
from collections import Counter
import re
import os
import uuid
from backend.groq_chat import GroqChat
from backend.get_transcript import YouTubeTranscriptDownloader
from backend.audio_generator import generate_audio, AUDIO_CACHE_DIR, ensure_audio_cache_dir, DEFAULT_ELEVENLABS_VOICE_ID_A, DEFAULT_ELEVENLABS_VOICE_ID_B
from pydub import AudioSegment
import io # For BytesIO if needed for pydub
import hashlib
import logging # Added import

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# --- Audio Playback Helper ---

def play_conversation_audio(conversation_text: str, item_id: str):
    """
    Provides a button to play the full multi-voice conversation audio.
    Audio is generated and cached if not already available.
    """
    if not conversation_text:
        st.caption("No conversation text available for audio.")
        return

    # Create a unique hash for the button key based on conversation content and item_id
    conversation_hash = hashlib.md5(conversation_text.encode('utf-8')).hexdigest()
    safe_item_id = str(item_id).replace('-', '')
    button_label = "▶️ Play Conversation"
    button_key = f"play_audio_conversation_{conversation_hash}_{safe_item_id}"

    if st.button(button_label, key=button_key):
        # Call the helper function to get/generate the combined conversation audio path
        combined_audio_path = get_or_generate_conversation_audio_path(conversation_text, item_id)
        if combined_audio_path:
            st.audio(combined_audio_path, format='audio/mp3')
            logger.info(f"Playing combined conversation audio: {combined_audio_path}")
        else:
            st.error("Could not generate or find combined conversation audio.")
            logger.error(f"Failed to get/generate conversation audio for item ID: {item_id}")

# --- Conversation Audio Generation/Retrieval Helper ---
def get_or_generate_conversation_audio_path(conversation_text: str, item_id: str) -> str | None:
    """
    Generates (if needed) and caches the full multi-voice conversation audio.
    Returns the path to the combined audio file, or None on failure.
    """
    if not conversation_text:
        logger.warning(f"No conversation text provided for conversation audio generation: item {item_id}")
        return None

    ensure_audio_cache_dir()

    conversation_hash = hashlib.md5(conversation_text.encode('utf-8')).hexdigest()
    safe_item_id = str(item_id).replace('-', '')
    combined_audio_filename = f"conv_{conversation_hash}_{safe_item_id}.mp3"
    combined_audio_cache_path = os.path.join(AUDIO_CACHE_DIR, combined_audio_filename)

    if os.path.exists(combined_audio_cache_path):
        logger.info(f"Using cached combined conversation audio: {combined_audio_cache_path}")
        return combined_audio_cache_path
    else:
        logger.info(f"Generating combined conversation audio for item ID: {item_id} (Not found in cache: {combined_audio_cache_path})")
        lines = conversation_text.strip().split('\n')
        segments = []
        default_voice = DEFAULT_ELEVENLABS_VOICE_ID_A # Fallback voice

        for i, line in enumerate(lines):
            line_strip = line.strip()
            if not line_strip:
                continue

            text_to_speak = line_strip
            current_speaker_voice = default_voice
            speaker_log_prefix = "DefaultSpeaker"
            # Use a more specific label for conversation segments for clarity in caching/logging
            segment_text_type_label = f"conversation_segment_{i}" 

            # For now, use Voice A for all speakers to avoid voice limit issues
            current_speaker_voice = DEFAULT_ELEVENLABS_VOICE_ID_A 
            speaker_log_prefix = "Speaker"

            if line_strip.startswith("女："):
                # current_speaker_voice = DEFAULT_ELEVENLABS_VOICE_ID_B # Female voice (Temporarily disabled)
                text_to_speak = line_strip.replace("女：", "", 1).strip()
                speaker_log_prefix = "女"
            elif line_strip.startswith("男："):
                # current_speaker_voice = DEFAULT_ELEVENLABS_VOICE_ID_A # Male voice (Already set as default for now)
                text_to_speak = line_strip.replace("男：", "", 1).strip()
                speaker_log_prefix = "男"
            else: # Handle lines without speaker prefixes, if any, using default voice
                text_to_speak = line_strip
            
            if not text_to_speak:
                continue

            # Use get_or_generate_audio_path for individual segments
            # Unique prefix for segments to distinguish from other audio types for the same item_id
            segment_unique_prefix = f"conv_seg_{speaker_log_prefix}_{i}" # Ensure unique prefix per segment
            segment_path = get_or_generate_audio_path(text_to_speak, current_speaker_voice, segment_text_type_label, item_id, unique_prefix=segment_unique_prefix)
            
            if segment_path and os.path.exists(segment_path):
                try:
                    audio_segment = AudioSegment.from_mp3(segment_path)
                    segments.append(audio_segment)
                    logger.info(f"Loaded segment for '{speaker_log_prefix}: {text_to_speak[:20]}...' using voice {current_speaker_voice}")
                except Exception as e:
                    logger.error(f"Error loading segment {segment_path} with pydub: {e}")
            else:
                logger.warning(f"Could not generate or find audio segment for: {text_to_speak}")

        if segments:
            try:
                combined_audio = sum(segments)
                combined_audio.export(combined_audio_cache_path, format="mp3")
                logger.info(f"Successfully combined and cached conversation audio: {combined_audio_cache_path}")
                return combined_audio_cache_path
            except Exception as e:
                logger.error(f"Error combining or exporting conversation audio: {e}")
                return None
        else:
            logger.warning("No audio segments were generated for the conversation.")
            return None

# --- Comprehensive Audio Helper ---
def generate_and_play_comprehensive_audio(question_data: dict, item_id: str):
    """
    Generates a single audio file combining intro, conversation, question, ping, and answer.
    Provides a button to play this comprehensive audio.
    """
    if not question_data:
        st.caption("No question data available for comprehensive audio.")
        return

    ensure_audio_cache_dir()

    # Create a unique hash for the comprehensive audio based on all its text parts
    # This ensures that if any part changes, a new comprehensive audio is generated.
    intro_text = question_data.get('introduction', '')
    conversation_text = question_data.get('conversation', '')
    question_text = question_data.get('question', '')
    correct_answer_text = question_data.get('correct_answer_letter', '') # Assuming this is the text to speak for the answer
    # If correct_answer_letter is 'A', 'B', etc., you might want to map it to the full option text
    # For now, let's assume it's 'A', 'B', 'C', or 'D' and we'll speak that letter.
    # If it's the full answer text, that's even better.
    # We need to clarify what 'correct_answer_letter' actually holds. For now, we'll use it directly.
    # If 'correct_answer_letter' gives 'A', and options are {'A': 'Apple', ...}, we might want to speak 'Apple'.
    # Let's assume 'correct_answer_letter' is the text to be spoken for the answer for now.

    all_text_content_for_hash = f"{intro_text}{conversation_text}{question_text}{correct_answer_text}"
    comprehensive_audio_hash = hashlib.md5(all_text_content_for_hash.encode('utf-8')).hexdigest()
    safe_item_id = str(item_id).replace('-', '')
    comprehensive_filename = f"comp_{comprehensive_audio_hash}_{safe_item_id}.mp3"
    comprehensive_cache_path = os.path.join(AUDIO_CACHE_DIR, comprehensive_filename)

    button_label = "▶️ Play Full Listening Exercise"
    button_key = f"play_comprehensive_audio_{comprehensive_audio_hash}_{safe_item_id}"

    if st.button(button_label, key=button_key):
        if os.path.exists(comprehensive_cache_path):
            logger.info(f"Playing cached comprehensive audio: {comprehensive_cache_path}")
            st.audio(comprehensive_cache_path, format='audio/mp3')
            return

        logger.info(f"Generating comprehensive audio for item ID: {item_id} (Not found in cache: {comprehensive_cache_path})")
        
        audio_segments = []
        valid_paths = True

        # 1. Introduction
        if intro_text:
            intro_path = get_or_generate_audio_path(intro_text, DEFAULT_ELEVENLABS_VOICE_ID_A, "Introduction", item_id, "comp_intro")
            if intro_path:
                audio_segments.append(AudioSegment.from_mp3(intro_path))
                audio_segments.append(AudioSegment.silent(duration=750)) # Pause after intro
            else: valid_paths = False

        # 2. Conversation
        if valid_paths and conversation_text:
            convo_path = get_or_generate_conversation_audio_path(conversation_text, item_id)
            if convo_path:
                audio_segments.append(AudioSegment.from_mp3(convo_path))
                audio_segments.append(AudioSegment.silent(duration=750)) # Pause after conversation
            else: valid_paths = False

        # 3. Question
        if valid_paths and question_text:
            question_path = get_or_generate_audio_path(question_text, DEFAULT_ELEVENLABS_VOICE_ID_A, "Question", item_id, "comp_question")
            if question_path:
                audio_segments.append(AudioSegment.from_mp3(question_path))
                audio_segments.append(AudioSegment.silent(duration=750)) # Pause after question
            else: valid_paths = False
        
        # 4. Ping Sound
        if valid_paths:
            ping_path = get_ping_sound_path()
            if ping_path: 
                audio_segments.append(AudioSegment.from_mp3(ping_path))
                audio_segments.append(AudioSegment.silent(duration=750)) # Pause after ping
            else: valid_paths = False

        # 5. Correct Answer
        actual_answer_to_speak = ""

        options_list = question_data.get('options') # This is a list of strings
        correct_answer_letter = question_data.get('correct_answer_letter', '').strip().upper()
        correct_answer_index = question_data.get('correct_answer') # This is an integer index

        logger.info(f"Comprehensive Audio: Raw correct_answer_letter: '{correct_answer_letter}', Raw correct_answer_index: {correct_answer_index}")
        logger.info(f"Comprehensive Audio: Options list: {options_list}")

        actual_answer_to_speak = None
        if isinstance(options_list, list) and options_list:
            if correct_answer_index is not None and 0 <= correct_answer_index < len(options_list):
                actual_answer_to_speak = options_list[correct_answer_index]
                logger.info(f"Comprehensive Audio: Fetched answer text using correct_answer_index ({correct_answer_index}): '{actual_answer_to_speak}'")
            elif correct_answer_letter and 'A' <= correct_answer_letter <= 'D':
                # Fallback to deriving index from letter if correct_answer_index is missing/invalid
                try:
                    letter_to_index = {'A': 0, 'B': 1, 'C': 2, 'D': 3}
                    idx = letter_to_index[correct_answer_letter]
                    if 0 <= idx < len(options_list):
                        actual_answer_to_speak = options_list[idx]
                        logger.info(f"Comprehensive Audio: Fetched answer text using correct_answer_letter '{correct_answer_letter}' (index {idx}): '{actual_answer_to_speak}'")
                    else:
                        logger.warning(f"Comprehensive Audio: Derived index {idx} from letter '{correct_answer_letter}' is out of bounds for options list.")
                except KeyError:
                    logger.warning(f"Comprehensive Audio: Could not derive a valid index from correct_answer_letter '{correct_answer_letter}'.")
            else:
                logger.warning("Comprehensive Audio: Options list is present, but couldn't determine correct answer text from index or letter.")
        
        if not actual_answer_to_speak and correct_answer_letter: # Ultimate fallback to speak letter if text not found
             logger.warning(f"Comprehensive Audio: Could not determine full answer text. Falling back to speaking the letter: '{correct_answer_letter}'")
             actual_answer_to_speak = correct_answer_letter
        elif not actual_answer_to_speak:
            logger.warning("Comprehensive Audio: correct_answer_letter is empty or no answer text could be determined.")

        if valid_paths and actual_answer_to_speak:
            logger.info(f"Comprehensive Audio: Attempting to generate audio for answer: '{actual_answer_to_speak}'")
            answer_path = get_or_generate_audio_path(actual_answer_to_speak, DEFAULT_ELEVENLABS_VOICE_ID_A, "Correct Answer", item_id, "comp_answer")
            if answer_path:
                logger.info(f"Comprehensive Audio: Generated answer audio path: {answer_path}")
                audio_segments.append(AudioSegment.from_mp3(answer_path))
            else:
                logger.error(f"Comprehensive Audio: Failed to generate audio for answer: '{actual_answer_to_speak}'")
                valid_paths = False
        elif not actual_answer_to_speak:
             logger.warning("Comprehensive Audio: No actual_answer_to_speak determined, skipping answer audio.")
        elif not valid_paths:
            logger.warning("Comprehensive Audio: valid_paths is False prior to answer, skipping answer audio generation.")

        if audio_segments and valid_paths:
            try:
                final_audio = sum(audio_segments)
                final_audio.export(comprehensive_cache_path, format="mp3")
                logger.info(f"Successfully generated and cached comprehensive audio: {comprehensive_cache_path}")
                st.audio(comprehensive_cache_path, format='audio/mp3')
            except Exception as e:
                logger.error(f"Error combining or exporting comprehensive audio: {e}")
                st.error("Could not generate comprehensive audio.")
        elif not valid_paths:
            st.error("Failed to generate one or more audio components for the comprehensive audio.")
        else:
            st.warning("No audio segments were available to create comprehensive audio.")

# --- Ping Sound Helper ---
from pydub.generators import Sine

def get_ping_sound_path() -> str | None:
    """
    Ensures a ping sound file exists and returns its path.
    Generates a simple sine wave if ping.mp3 is not found.
    """
    ensure_audio_cache_dir()
    ping_filename = "ping.mp3"
    ping_cache_path = os.path.join(AUDIO_CACHE_DIR, ping_filename)

    if os.path.exists(ping_cache_path):
        logger.info(f"Using existing ping sound: {ping_cache_path}")
        return ping_cache_path
    else:
        logger.info(f"Generating ping sound: {ping_cache_path}")
        try:
            # Generate a 440 Hz (A4) sine wave for 0.5 seconds
            ping_sound = Sine(440).to_audio_segment(duration=500, volume=-20) # volume in dBFS
            # Add a short fade out to make it sound more like a ping
            ping_sound = ping_sound.fade_out(150)
            ping_sound.export(ping_cache_path, format="mp3")
            logger.info(f"Successfully generated ping sound: {ping_cache_path}")
            return ping_cache_path
        except Exception as e:
            logger.error(f"Could not generate ping sound: {e}")
            return None

# --- Audio Generation/Retrieval Helper ---
def get_or_generate_audio_path(text_content: str, voice_id: str, text_type_label: str, item_id: str, unique_prefix: str = "item") -> str | None:
    """
    Ensures audio is generated and cached, then returns the path.
    Returns None if generation fails or text_content is empty.
    """
    if not text_content:
        logger.warning(f"No text content provided for audio generation: {text_type_label}, item {item_id}")
        return None

    ensure_audio_cache_dir()
    
    safe_item_id = str(item_id).replace('-', '') # Sanitize item_id for filename
    # Create a hash based on text content and voice_id to ensure uniqueness for different voices on same text
    text_and_voice_hash = hashlib.md5(f"{text_content}{voice_id}".encode('utf-8')).hexdigest()
    audio_filename = f"{unique_prefix}_{text_type_label.lower().replace(' ', '_')}_{text_and_voice_hash}_{safe_item_id}.mp3"
    audio_cache_path = os.path.join(AUDIO_CACHE_DIR, audio_filename)

    if os.path.exists(audio_cache_path):
        logger.info(f"Using cached audio for '{text_type_label}': {audio_cache_path}")
        return audio_cache_path
    else:
        logger.info(f"Generating audio for '{text_type_label}' (Not found in cache: {audio_cache_path}) with voice {voice_id}")
        # generate_audio is imported from backend.audio_generator
        generated_path = generate_audio(text_content, voice_id, audio_filename) 
        if generated_path and os.path.exists(generated_path):
            logger.info(f"Successfully generated audio: {generated_path}")
            return generated_path
        else:
            logger.error(f"Failed to generate audio for '{text_type_label}': {text_content[:50]}...")
            return None

# --- Audio Playback Helper --- (Original play_audio_for_text remains below this new function)
def play_audio_for_text(text_content: str, text_type_label: str, item_id: str, unique_prefix: str = "item"):
    """
    Generates (if needed) and plays audio for the given text_content.
    Uses a simple caching mechanism based on the hash of the text.
    """
    if not text_content:
        # Using st.caption for a less intrusive message if no text is available
        st.caption(f"No text available for {text_type_label.replace('_', ' ').title()} audio.")
        return

    # Use DEFAULT_ELEVENLABS_VOICE_ID_A for single speaker parts like intro, question, answer
    voice_id_to_use = DEFAULT_ELEVENLABS_VOICE_ID_A 
    
    button_label = f"▶️ Play {text_type_label}"
    # Create a unique key for the button based on content and type to avoid Streamlit duplicate key errors
    button_key_hash_content = hashlib.md5(f"{text_content}{text_type_label}{item_id}{unique_prefix}{voice_id_to_use}".encode('utf-8')).hexdigest()
    button_key = f"play_audio_{text_type_label.lower().replace(' ', '_')}_{button_key_hash_content}"

    if st.button(button_label, key=button_key):
        # Call the helper function to get/generate the audio path
        audio_path = get_or_generate_audio_path(text_content, voice_id_to_use, text_type_label, item_id, unique_prefix)
        if audio_path:
            st.audio(audio_path, format='audio/mp3')
            logger.info(f"Playing audio for '{text_type_label}': {audio_path}")
        else:
            st.error(f"Could not generate or find audio for {text_type_label}.")

# Page config
st.set_page_config(
    page_title="Japanese Learning Assistant",
    page_icon="🎌",
    layout="wide"
)

# Initialize session state
if 'transcript' not in st.session_state:
    st.session_state.transcript = None
if 'messages' not in st.session_state:
    st.session_state.messages = []
# --- Global Variables & Constants ---
HISTORY_FILE_PATH = os.getenv("HISTORY_FILE_PATH", "listening_learning_assistant/data/history.json")

def generate_unique_id():
    """Generates a unique ID for questions."""
    return str(uuid.uuid4())

if 'generated_questions_history' not in st.session_state:
    st.session_state.generated_questions_history = [] # Will be populated by load_question_history()

def render_header():
    """Render the header section"""
    st.title("🎌 Japanese Learning Assistant")
    st.markdown("""
    Transform YouTube transcripts into interactive Japanese learning experiences.
    
    This tool demonstrates:
    - Base LLM Capabilities
    - RAG (Retrieval Augmented Generation)
    - Amazon Bedrock Integration
    - Agent-based Learning Systems
    """)

def render_sidebar():
    """Render the sidebar with component selection"""
    with st.sidebar:
        st.header("Development Stages")
        
        # Main component selection
        selected_stage = st.radio(
            "Select Stage:",
            [
                "1. Chat with Groq(Qwen)",
                "2. Raw Transcript",
                "3. Structured Data",
                "4. RAG Implementation",
                "5. Interactive Learning"
            ]
        )
        
        # Stage descriptions
        stage_info = {
            "1. Chat with Groq(Qwen)": """
            **Current Focus:**
            - Basic Japanese learning
            - Understanding LLM capabilities
            - Identifying limitations
            """,
            
            "2. Raw Transcript": """
            **Current Focus:**
            - YouTube transcript download
            - Raw text visualization
            - Initial data examination
            """,
            
            "3. Structured Data": """
            **Current Focus:**
            - Text cleaning
            - Dialogue extraction
            - Data structuring
            """,
            
            "4. RAG Implementation": """
            **Current Focus:**
            - Bedrock embeddings
            - Vector storage
            - Context retrieval
            """,
            
            "5. Interactive Learning": """
            **Current Focus:**
            - Scenario generation
            - Audio synthesis
            - Interactive practice
            """
        }
        
        st.markdown("---")
        st.markdown(stage_info[selected_stage])
        
        render_question_history_sidebar() # Display question history
        
        return selected_stage

def render_chat_stage():
    """Render an improved chat interface"""
    st.header("Chat with Groq(Qwen)")

    # Initialize GroqChat instance if not in session state
    if 'groq_chat' not in st.session_state:
        st.session_state.groq_chat = GroqChat()

    # Introduction text
    st.markdown("""
    Start by exploring Groq(Qwen)'s base Japanese language capabilities. Try asking questions about Japanese grammar, 
    vocabulary, or cultural aspects.
    """)

    # Initialize chat history if not exists
    if "messages" not in st.session_state:
        st.session_state.messages = []

    # Display chat messages
    for message in st.session_state.messages:
        with st.chat_message(message["role"], avatar="🧑‍💻" if message["role"] == "user" else "🤖"):
            st.markdown(message["content"])

    # Chat input area
    if prompt := st.chat_input("Ask about Japanese language..."):
        # Process the user input
        process_message(prompt)

    # Example questions in sidebar
    with st.sidebar:
        st.markdown("### Try These Examples")
        example_questions = [
            "How do I say 'Where is the train station?' in Japanese?",
            "Explain the difference between は and が",
            "What's the polite form of 食べる?",
            "How do I count objects in Japanese?",
            "What's the difference between こんにちは and こんばんは?",
            "How do I ask for directions politely?"
        ]
        
        for q in example_questions:
            if st.button(q, use_container_width=True, type="secondary"):
                # Process the example question
                process_message(q)
                st.rerun()

    # Add a clear chat button
    if st.session_state.messages:
        if st.button("Clear Chat", type="primary"):
            st.session_state.messages = []
            st.rerun()

def format_response_with_thinking(response: str) -> str:
    """Format response to handle <think> tags specially"""
    if not response:
        return response
        
    # Split the response into parts based on <think> tags
    parts = re.split(r'(<think>.*?</think>)', response, flags=re.DOTALL)
    
    formatted_parts = []
    for part in parts:
        if part.startswith('<think>') and part.endswith('</think>'):
            # Format thinking content with a different style
            thinking_content = part[7:-8].strip()  # Remove <think> tags
            formatted_parts.append(
                f'<div style="color: #666; font-style: italic; '
                f'border-left: 3px solid #ddd; padding-left: 10px; margin: 5px 0;">'
                f'💭 {thinking_content}</div>'
            )
        else:
            formatted_parts.append(part)
    
    return ''.join(formatted_parts)

def process_message(message: str):
    """Process a message and generate a response"""
    # Add user message to state and display
    st.session_state.messages.append({"role": "user", "content": message})
    with st.chat_message("user", avatar="🧑‍💻"):
        st.markdown(message)

    # Generate and display assistant's response
    with st.chat_message("assistant", avatar="🤖"):
        response = st.session_state.groq_chat.generate_response(message)
        if response:
            # Format the response to handle <think> tags
            formatted_response = format_response_with_thinking(response)
            st.markdown(formatted_response, unsafe_allow_html=True)
            st.session_state.messages.append({"role": "assistant", "content": response})

def count_characters(text):
    """Count Japanese and total characters in text"""
    if not text:
        return 0, 0
        
    def is_japanese(char):
        return any([
            '\u4e00' <= char <= '\u9fff',  # Kanji
            '\u3040' <= char <= '\u309f',  # Hiragana
            '\u30a0' <= char <= '\u30ff',  # Katakana
        ])
    
    jp_chars = sum(1 for char in text if is_japanese(char))
    return jp_chars, len(text)

def render_transcript_stage():
    """Render the raw transcript stage"""
    st.header("Raw Transcript Processing")
    
    # URL input
    url = st.text_input(
        "YouTube URL",
        placeholder="Enter a Japanese lesson YouTube URL"
    )
    
    # Download button and processing
    if url:
        if st.button("Download Transcript"):
            try:
                with st.spinner("Downloading transcript..."):
                    downloader = YouTubeTranscriptDownloader()
                    video_id = downloader.extract_video_id(url)
                    transcript = downloader.get_transcript(url)
                    if transcript:
                        # Store the raw transcript text in session state
                        try:
                            if isinstance(transcript, list):
                                transcript_text = "\n".join([entry['text'] for entry in transcript])
                                
                                # Ensure the transcripts directory exists
                                os.makedirs("backend/transcripts", exist_ok=True)
                                
                                # Save the transcript using the downloader's save_transcript method
                                try:
                                    # Save the transcript to a temporary file first
                                    temp_success = downloader.save_transcript(transcript, f"{video_id}_temp")
                                    
                                    if temp_success:
                                        # Move the file to the correct location
                                        temp_path = f"transcripts/{video_id}_temp.txt"
                                        target_path = f"backend/transcripts/{video_id}.txt"
                                        os.rename(temp_path, target_path)
                                        st.success(f"Transcript downloaded and saved to {target_path}")
                                    else:
                                        st.warning("Transcript downloaded but there was an issue saving to file")
                                except Exception as e:
                                    st.error(f"Error saving transcript: {str(e)}")
                            else:
                                # Handle case where transcript is not in expected format
                                st.warning("Unexpected transcript format. Trying to process...")
                                transcript_text = str(transcript)
                            
                            st.session_state.transcript = transcript_text
                        except Exception as e:
                            st.error(f"Error processing transcript: {str(e)}")
                    else:
                        st.error("No transcript found for this video.")
            except Exception as e:
                st.error(f"Error downloading transcript: {str(e)}")

    col1, col2 = st.columns(2)
    
    with col1:
        st.subheader("Raw Transcript")
        if st.session_state.transcript:
            st.text_area(
                label="Raw text",
                value=st.session_state.transcript,
                height=400,
                disabled=True,
                key="transcript_area"
            )
    
        else:
            st.info("No transcript loaded yet")
    
    with col2:
        st.subheader("Transcript Stats")
        if st.session_state.transcript:
            # Calculate stats
            jp_chars, total_chars = count_characters(st.session_state.transcript)
            total_lines = len(st.session_state.transcript.split('\n'))
            
            # Display stats
            st.metric("Total Characters", total_chars)
            st.metric("Japanese Characters", jp_chars)
            st.metric("Total Lines", total_lines)
        else:
            st.info("Load a transcript to see statistics")

def render_structured_stage():
    """Render the structured data stage"""
    st.header("Structured Data Processing")
    
    col1, col2 = st.columns(2)
    
    with col1:
        st.subheader("Dialogue Extraction")
        # Placeholder for dialogue processing
        st.info("Dialogue extraction will be implemented here")
        
    with col2:
        st.subheader("Data Structure")
        # Placeholder for structured data view
        st.info("Structured data view will be implemented here")

def render_rag_stage():
    """Render the RAG implementation stage"""
    st.header("RAG System")
    
    # Query input
    query = st.text_input(
        "Test Query",
        placeholder="Enter a question about Japanese..."
    )
    
    col1, col2 = st.columns(2)
    
    with col1:
        st.subheader("Retrieved Context")
        # Placeholder for retrieved contexts
        st.info("Retrieved contexts will appear here")
        
    with col2:
        st.subheader("Generated Response")
        # Placeholder for LLM response
        st.info("Generated response will appear here")

# Initialize session state for question management
if 'current_question' not in st.session_state:
    st.session_state.current_question = None
if 'user_answer' not in st.session_state:
    st.session_state.user_answer = None
if 'show_feedback' not in st.session_state:
    st.session_state.show_feedback = False

def render_question(question: Dict):
    logger.debug(f"render_question: received question: {question}")
    logger.debug(f"render_question: received question keys: {list(question.keys()) if isinstance(question, dict) else 'Not a dict'}")
    """Render a question with options"""
    intro_text = question.get('introduction', '')
    conversation_text = question.get('conversation', '')
    actual_question_text = question.get('question', '')

    if intro_text and conversation_text and actual_question_text:
        st.subheader("Listening Practice")
        if intro_text:
            st.write(f"**Situation:** {intro_text}")
            play_audio_for_text(intro_text, "Introduction", question['id'])
        st.write("**Dialogue:**")
        # Process conversation lines to ensure newlines are respected
        conversation_lines = [line.strip() for line in conversation_text.split('\n') if line.strip()]
        formatted_conversation = '\n'.join(conversation_lines)
        st.text(formatted_conversation)  # Use st.text to preserve newlines
        play_conversation_audio(conversation_text, question['id']) # Use new function for conversation
        st.write(f"**Question:** {actual_question_text}")
        play_audio_for_text(actual_question_text, "Question Text", question['id'])

        # Add button for comprehensive audio playback
        st.markdown("---") # Add a visual separator
        generate_and_play_comprehensive_audio(question, question['id'])
        st.markdown("---") # Add a visual separator

    elif actual_question_text: # Fallback if only question text is available
        st.subheader("Question")
        st.write(actual_question_text)
        play_audio_for_text(actual_question_text, "Question Text", question['id'])

        # Add button for comprehensive audio playback (even if only question text is available)
        st.markdown("---") # Add a visual separator
        generate_and_play_comprehensive_audio(question, question['id'])
        st.markdown("---") # Add a visual separator

    else:
        st.warning("Question data is not in the expected format.")
        return None
    
    # Display options
    options = question.get('options', [])
    if not options:
        st.warning("No options available.")
        return None
    
    # Create option labels (A, B, C, D)
    option_letters = ['A', 'B', 'C', 'D']
    
    # Display radio buttons for options
    selected_option = st.radio(
        "Select your answer:",
        option_letters[:len(options)],
        format_func=lambda x: options[ord(x) - ord('A')] if ord(x) - ord('A') < len(options) else ""
    )
    
    # Check answer button with unique key
    if st.button("Check Answer", key="check_answer_btn"):
        st.session_state.user_answer = selected_option
        st.session_state.show_feedback = True
        
        # Mark this question as answered in the history
        if st.session_state.current_question and 'question_id' in st.session_state.current_question:
            current_q_id = st.session_state.current_question['question_id']
            history_updated = False
            for hist_q in st.session_state.generated_questions_history:
                if hist_q.get('question_id') == current_q_id:
                    if not hist_q.get('answered_in_main_ui', False):
                        hist_q['answered_in_main_ui'] = True
                        history_updated = True
                    break
            if history_updated:
                save_question_history()
        st.rerun()
    
    return selected_option

def render_feedback(question, user_answer, question_type, topic=None):
    """Render feedback for the user's answer"""
    if user_answer is None:
        return
        
    st.subheader("Feedback")
    
    # Get the correct answer letter
    correct_idx = question.get('correct_answer', 0)
    correct_letter = chr(ord('A') + correct_idx)
    
    if user_answer == correct_letter:
        st.success("✅ Correct!")
    else:
        st.error(f"❌ Incorrect. The correct answer is {correct_letter}.")
    
    # Show explanation if available
    explanation = question.get('explanation', '')
    if explanation:
        st.info(f"💡 **Explanation:** {explanation}")
    
    # Show raw response for debugging if available with unique key
    if 'raw_response' in question and st.checkbox("Show debug info", key="debug_info_checkbox"):
        with st.expander("Debug: Raw Response"):
            st.text(question['raw_response'])
    

def generate_new_question(question_type, topic=None):
    """Generate a new question and update session state"""
    with st.spinner(f"Generating {question_type.lower()} question..."):
        try:
            # Import here to catch import errors
            try:
                from backend.question_generator import question_generator
            except ImportError as ie:
                st.error(f"Import error: {str(ie)}")
                print(f"Import error: {str(ie)}")
                return
                
            # Generate the question
            question_data = question_generator.generate_question(
                question_type=question_type,
                topic=topic if topic else None
            )
            logger.info(f"generate_new_question: Raw question_data from generator: {question_data}")
            if not question_data or not isinstance(question_data, dict):
                st.error("Failed to generate a valid question structure. Please try again.")
                logger.error("generate_new_question: question_data from generator was None or not a dict.")
                return # Exit early if question_data is not valid

            logger.info(f"generate_new_question: Value of 'correct_answer_letter' from generator: '{question_data.get('correct_answer_letter')}'")
            # The 'if question_data:' check is now implicitly handled by the block above, 
            # but we keep it if further specific checks on a valid dict are needed.
            if question_data:
                # Add unique ID and answered status before storing
                question_data['question_id'] = str(uuid.uuid4())
                question_data['answered_in_main_ui'] = False
                
                question_data['id'] = generate_unique_id() # Ensure new questions have an ID
                logger.debug(f"generate_new_question: question_data after adding id: {question_data}")
                logger.debug(f"generate_new_question: question_data keys after adding id: {list(question_data.keys()) if isinstance(question_data, dict) else 'Not a dict'}")
                st.session_state.current_question = question_data
                # Store in history
                if 'generated_questions_history' not in st.session_state: # Defensive initialization
                    st.session_state.generated_questions_history = []
                st.session_state.generated_questions_history.insert(0, question_data) # Add to the beginning of the list
                save_question_history() # Save after adding a new question
                
                st.session_state.user_answer = None
                st.session_state.show_feedback = False
                st.rerun()
            else:
                error_msg = "Failed to generate question. The response format might be incorrect."
                st.error(error_msg)
                print(error_msg)
                
        except Exception as e:
            error_msg = f"Error generating question: {str(e)}"
            st.error(error_msg)
            print(error_msg)
            
            # Print traceback for debugging
            import traceback
            print("\n" + "="*50)
            print("Full traceback:")
            traceback.print_exc()
            print("="*50 + "\n")

def render_question_history_sidebar():
    if 'generated_questions_history' in st.session_state and st.session_state.generated_questions_history:
        with st.sidebar:
            st.markdown("---") # Separator
            st.subheader("Question History")
            # Display in reverse order (newest first)
            for i, q_item in enumerate(reversed(st.session_state.generated_questions_history)):
                question_text = q_item.get('question', 'N/A')
                # Truncate for display if too long
                display_text = (question_text[:30] + '...') if len(question_text) > 33 else question_text
                
                # Use an expander for each question to show more details
                with st.expander(f"Q{len(st.session_state.generated_questions_history) - i}: {display_text}"):
                    st.markdown(f"**Introduction:** {q_item.get('introduction', 'N/A')}")
                    play_audio_for_text(q_item.get('introduction'), "Introduction", q_item.get('id'), unique_prefix=f"hist_{i}")
                    
                    st.markdown("**Conversation:**")
                    conversation_text_hist = q_item.get('conversation', '')
                    if isinstance(conversation_text_hist, list):
                        conversation_text_hist = "\n".join(conversation_text_hist)
                    st.markdown(f"```\n{conversation_text_hist}\n```")
                    play_audio_for_text(conversation_text_hist, "Conversation", q_item.get('id'), unique_prefix=f"hist_{i}")
                    
                    st.markdown(f"**Question:** {question_text}")
                    play_audio_for_text(question_text, "Question Text", q_item.get('id'), unique_prefix=f"hist_{i}")
                    options = q_item.get('options', [])
                    correct_idx = q_item.get('correct_answer', -1)  # Or 'correct_answer_index' if that's the key
                    if options:
                        st.markdown("**Options:**")
                        question_answered = q_item.get('answered_in_main_ui', False)
                        for opt_idx, opt_text in enumerate(options):
                            if question_answered:
                                prefix = "✅" if opt_idx == correct_idx else "❌" # Show correct/incorrect if answered
                            else:
                                prefix = "➡️" # Neutral if not yet answered in main UI
                            st.markdown(f"{prefix} {chr(ord('A') + opt_idx)}. {opt_text}")
                    # Potential future enhancement: Button to reload this question
                    # if st.button(f"View Q{len(st.session_state.generated_questions_history) - i}", key=f"history_q_{i}_{q_item.get('question_id', i)}"): # Ensure unique key
                    #     st.session_state.current_question = q_item
                    #     st.session_state.user_answer = None
                    #     st.session_state.show_feedback = False
                    #     st.rerun()

def render_interactive_stage():
    """Render the interactive learning stage"""
    st.header("Interactive Learning")
    
    # Practice type selection
    col1, col2 = st.columns([2, 1])
    
    with col1:
        selected_type = st.selectbox(
            "Select Practice Type",
            ["Dialogue Practice", "Vocabulary Quiz", "Listening Exercise"]
        )
    
    with col2:
        topic = st.text_input("Topic (optional)", "")
    
    # Generate new question button
    if st.button("Generate New Question"):
        generate_new_question(selected_type, topic if topic else None)
    
    # Display current question or placeholder
    if st.session_state.current_question:
        question = st.session_state.current_question
        
        col1, col2 = st.columns([2, 1])
        
        with col1:
            # Display question and options
            user_answer = render_question(question)
            
            # Show feedback if available
            if st.session_state.show_feedback:
                render_feedback(
                    question, 
                    st.session_state.user_answer,
                    selected_type,
                    topic if topic else None
                )
        
        with col2:
            # Placeholder for audio (could be implemented later)
            st.subheader("Audio")
            st.info("Audio player will appear here")
            
            # Show additional context if available
            if 'context' in question:
                with st.expander("Additional Context"):
                    st.write(question['context'])
    else:
        st.info("Click 'Generate New Question' to start practicing!")

def load_question_history():
    """Loads question history from the JSON file into session state."""
    try:
        if os.path.exists(HISTORY_FILE_PATH):
            with open(HISTORY_FILE_PATH, 'r', encoding='utf-8') as f:
                history = json.load(f)
                # Basic validation: ensure it's a list
                if isinstance(history, list):
                    st.session_state.generated_questions_history = history
                else:
                    st.session_state.generated_questions_history = []
                    logging.warning(f"History file {HISTORY_FILE_PATH} did not contain a list.")
        else:
            st.session_state.generated_questions_history = []
    except (json.JSONDecodeError, IOError) as e:
        logging.error(f"Error loading question history from {HISTORY_FILE_PATH}: {e}")
        st.session_state.generated_questions_history = [] # Reset on error

def save_question_history():
    """Saves the current question history from session state to the JSON file."""
    try:
        # Ensure the directory exists
        os.makedirs(os.path.dirname(HISTORY_FILE_PATH), exist_ok=True)
        with open(HISTORY_FILE_PATH, 'w', encoding='utf-8') as f:
            json.dump(st.session_state.generated_questions_history, f, ensure_ascii=False, indent=4)
    except IOError as e:
        logging.error(f"Error saving question history to {HISTORY_FILE_PATH}: {e}")

def main():
    load_question_history() # Load history at the start
    render_header()
    selected_stage = render_sidebar()
    
    # Render appropriate stage
    if selected_stage == "1. Chat with Groq(Qwen)":
        render_chat_stage()
    elif selected_stage == "2. Raw Transcript":
        render_transcript_stage()
    elif selected_stage == "3. Structured Data":
        render_structured_stage()
    elif selected_stage == "4. RAG Implementation":
        render_rag_stage()
    elif selected_stage == "5. Interactive Learning":
        render_interactive_stage()
    
    # Debug section at the bottom
    with st.expander("Debug Information"):
        st.json({
            "selected_stage": selected_stage,
            "transcript_loaded": st.session_state.transcript is not None,
            "chat_messages": len(st.session_state.messages)
        })

if __name__ == "__main__":
    main()