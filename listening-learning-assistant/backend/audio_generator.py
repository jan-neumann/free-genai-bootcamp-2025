"""
Backend module for generating audio using MeloTTS.
"""

import requests
import os
import logging

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# MeloTTS API URL (assuming Docker container is run with -p 8888:8080)
MELOTTS_API_URL = "http://localhost:8888/convert/tts"

AUDIO_CACHE_DIR = "listening-learning-assistant/data/audio_cache"

def ensure_audio_cache_dir():
    os.makedirs(AUDIO_CACHE_DIR, exist_ok=True)

def generate_audio(text: str, speaker_id: str, output_filename: str) -> str | None:
    """
    Generates audio for a given text and speaker ID using MeloTTS.
    Saves the audio to a file in the AUDIO_CACHE_DIR.

    Args:
        text (str): The text to synthesize.
        speaker_id (str): The speaker ID for MeloTTS.
        output_filename (str): The desired filename for the output audio (e.g., 'audio_id.wav').

    Returns:
        str | None: The full path to the saved audio file if successful, None otherwise.
    """
    ensure_audio_cache_dir()
    output_path = os.path.join(AUDIO_CACHE_DIR, output_filename)

    payload = {
        "text": text,
        "language": "JP",
        "speaker_id": speaker_id, # This might need to be a specific Japanese speaker ID from MeloTTS
        # "speed": "1.0" # Optional: control speed
    }
    headers = {
        "Content-Type": "application/json"
    }

    try:
        logger.info(f"Requesting audio from MeloTTS: {MELOTTS_API_URL} with payload: {payload}")
        response = requests.post(MELOTTS_API_URL, json=payload, headers=headers, timeout=60) # Increased timeout
        response.raise_for_status()  # Raises an HTTPError for bad responses (4XX or 5XX)

        # Save the received .wav data
        with open(output_path, 'wb') as f:
            f.write(response.content)
        logger.info(f"Successfully saved audio to {output_path}")
        return output_path

    except requests.exceptions.RequestException as e:
        logger.error(f"Error calling MeloTTS API: {e}")
        if hasattr(e, 'response') and e.response is not None:
            logger.error(f"Response status: {e.response.status_code}")
            logger.error(f"Response content: {e.response.text}")
        return None
    except IOError as e:
        logger.error(f"Error saving audio file {output_path}: {e}")
        return None

# --- Placeholder for multi-speaker conversation logic ---
# This will be more complex, involving parsing, multiple TTS calls, and audio concatenation.
# For example, using pydub for concatenation.

if __name__ == '__main__':
    # Example usage (for testing the structure)
    ensure_audio_cache_dir()
    test_text = "こんにちは、世界！"
    # Example usage for testing - ensure your MeloTTS Docker container is running and configured for Japanese.
    # You might need to find a valid Japanese speaker_id for MeloTTS.
    # For now, let's assume the container's default speaker is Japanese or a generic one works with 'JP' language.
    test_text = "こんにちは、世界！"
    test_speaker_id_jp = "JP" # Using language code as placeholder, actual speaker ID might be different or handled by container default
    test_filename = "test_jp_audio.wav"
    
    generated_file = generate_audio(test_text, test_speaker_id_jp, test_filename)
    if generated_file:
        logger.info(f"Test audio generated at: {generated_file}")
    else:
        logger.error("Test audio generation failed.")
