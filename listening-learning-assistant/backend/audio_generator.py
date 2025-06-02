"""
Backend module for generating audio using ElevenLabs API.
"""

import os
import logging
from elevenlabs.client import ElevenLabs
# from elevenlabs.errors import APIError # For specific error handling - Commented out for diagnostics

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

AUDIO_CACHE_DIR = "listening-learning-assistant/data/audio_cache"
DEFAULT_ELEVENLABS_VOICE_ID_A = "JBFqnCBsd6RMkjVDRZzb" # Example voice for Speaker A
DEFAULT_ELEVENLABS_VOICE_ID_B = "sB1b5zUrxQVAFl2PhZFp" # Placeholder for Speaker B - USER SHOULD UPDATE with a different voice ID
                                                 # User should verify/select a preferred Japanese voice
DEFAULT_ELEVENLABS_MODEL_ID = "eleven_flash_v2_5"

def ensure_audio_cache_dir():
    os.makedirs(AUDIO_CACHE_DIR, exist_ok=True)

def generate_audio(text: str, speaker_id: str, output_filename: str) -> str | None:
    """
    Generates audio for a given text and speaker ID using ElevenLabs API.
    Saves the audio to a file in the AUDIO_CACHE_DIR.
    The output_filename MUST end with '.mp3'.

    Args:
        text (str): The text to synthesize.
        speaker_id (str): The ElevenLabs voice_id to use.
        output_filename (str): The desired filename for the output audio (e.g., 'audio_id.mp3').

    Returns:
        str | None: The full path to the saved audio file if successful, None otherwise.
    """
    ensure_audio_cache_dir()
    
    if not output_filename.endswith(".mp3"):
        logger.error(f"Output filename '{output_filename}' must end with .mp3")
        # Optionally, we could auto-correct it, but for now, let's be strict.
        # output_filename = os.path.splitext(output_filename)[0] + ".mp3" 
        return None

    output_path = os.path.join(AUDIO_CACHE_DIR, output_filename)

    api_key = os.environ.get("ELEVENLABS_API_KEY")
    if not api_key:
        logger.error("ELEVENLABS_API_KEY environment variable not set.")
        return None

    try:
        client = ElevenLabs(api_key=api_key)
        
        logger.info(f"Requesting audio from ElevenLabs for text: '{text[:30]}...' with voice_id: {speaker_id}")
        
        audio_bytes_generator = client.text_to_speech.convert(
            text=text,
            voice_id=speaker_id,
            model_id=DEFAULT_ELEVENLABS_MODEL_ID,
            output_format="mp3_44100_128" 
        )

        with open(output_path, 'wb') as f:
            for chunk in audio_bytes_generator:
                f.write(chunk)
        
        logger.info(f"Successfully saved audio to {output_path}")
        return output_path

    except Exception as e: # Catching a more general exception for now
        logger.error(f"An error occurred during ElevenLabs API call or file operation: {type(e).__name__} - {e}")
        # If you want to see if it has a 'body' attribute (like APIError often does)
        if hasattr(e, 'body') and e.body:
             logger.error(f"Error details (if available): {e.body}")
        # To see the full traceback for this specific error in your console:
        # import traceback
        # logger.error(traceback.format_exc())
        return None

if __name__ == '__main__':
    # Example usage:
    # Make sure to set your ELEVENLABS_API_KEY environment variable before running.
    # export ELEVENLABS_API_KEY="your_key_here"
    
    ensure_audio_cache_dir()
    test_text_jp = "こんにちは、世界！これはイレブンラボからのテストです。"
    # Using the default voice ID, user might want to pick a specific Japanese voice from ElevenLabs.
    test_voice_id = DEFAULT_ELEVENLABS_VOICE_ID 
    test_filename_jp = "test_elevenlabs_jp_audio.mp3"

    logger.info(f"Attempting to generate test audio: '{test_filename_jp}'")
    generated_file_path = generate_audio(test_text_jp, test_voice_id, test_filename_jp)

    if generated_file_path:
        logger.info(f"Test audio generated successfully: {generated_file_path}")
    else:
        logger.error("Failed to generate test audio.")

    # Example with a different (hypothetical) voice ID if you have one
    # test_text_en = "Hello world! This is a test from ElevenLabs."
    # test_voice_id_en = "another_voice_id_for_english" # Replace with an actual English voice ID if needed
    # test_filename_en = "test_elevenlabs_en_audio.mp3"
    # generated_file_path_en = generate_audio(test_text_en, test_voice_id_en, test_filename_en)
    # if generated_file_path_en:
    #     logger.info(f"Test English audio generated successfully: {generated_file_path_en}")
    # else:
    #     logger.error("Failed to generate test English audio.")
