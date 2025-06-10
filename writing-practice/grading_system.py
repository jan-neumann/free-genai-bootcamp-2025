import os
import json
import logging
import time
from typing import Dict, Any, Optional, Union
from PIL import Image
import numpy as np

# Optional imports
try:
    from groq import Groq
    GROQ_AVAILABLE = True
except ImportError:
    GROQ_AVAILABLE = False
    logging.warning("Groq package not available. Grading system will be limited.")

try:
    from manga_ocr import MangaOcr
    MANGA_OCR_AVAILABLE = True
except ImportError:
    MANGA_OCR_AVAILABLE = False
    logging.warning("MangaOCR package not available. OCR functionality will be limited.")

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class GradingSystem:
    def __init__(self, groq_api_key: Optional[str] = None):
        """Initialize the grading system with OCR and LLM components."""
        self.groq_api_key = groq_api_key or os.getenv("GROQ_API_KEY")
        self.initialized = False
        
        # Log API key status (masking the actual key for security)
        if self.groq_api_key:
            key_preview = f"{self.groq_api_key[:4]}...{self.groq_api_key[-4:]}" if len(self.groq_api_key) > 8 else "[too short]"
            logger.info(f"Groq API key found: {key_preview}")
        else:
            logger.warning("No Groq API key found. Grading system will run in limited mode.")
            self.groq_client = None
            return
            
        if not GROQ_AVAILABLE:
            logger.warning("Groq package not available. Please install it with: pip install groq")
            self.groq_client = None
            return
            
        try:
            logger.debug("Initializing Groq client...")
            self.groq_client = Groq(api_key=self.groq_api_key.strip())
            
            # Test the API key with a simple request
            logger.debug("Testing Groq API key with a simple request...")
            try:
                # Make a simple API call to validate the key
                test_response = self.groq_client.chat.completions.create(
                    model="llama3-8b-8192",
                    messages=[{"role": "user", "content": "Say hello"}],
                    max_tokens=5
                )
                self.initialized = True
                logger.info("Groq client initialized and API key validated successfully")
            except Exception as test_error:
                error_msg = f"API key validation failed: {str(test_error)}"
                if hasattr(test_error, 'response') and hasattr(test_error.response, 'text'):
                    error_msg += f"\nResponse: {test_error.response.text}"
                logger.error(error_msg)
                self.groq_client = None
                raise ValueError("Invalid or unauthorized API key. Please check your GROQ_API_KEY.") from test_error
                
        except Exception as e:
            error_msg = f"Failed to initialize Groq client: {str(e)}"
            if hasattr(e, 'response') and hasattr(e.response, 'text'):
                error_msg += f"\nResponse: {e.response.text}"
            logger.error(error_msg)
            self.groq_client = None
            raise
        
        # Initialize OCR if available
        if MANGA_OCR_AVAILABLE:
            try:
                self.ocr = MangaOcr()
                logger.info("MangaOCR initialized successfully")
            except Exception as e:
                logger.error(f"Failed to initialize MangaOCR: {str(e)}")
                self.ocr = None
        else:
            self.ocr = None
        
        logger.info("Grading system initialized")
    
    def transcribe_image(self, image: Union[Image.Image, str]) -> str:
        """
        Transcribe Japanese text from an image using MangaOCR.
        
        Args:
            image: PIL Image or file path containing Japanese text
            
        Returns:
            str: Transcribed text or error message if OCR fails
        """
        if not self.ocr:
            error_msg = "OCR functionality not available. Please install MangaOCR."
            logger.error(error_msg)
            return "[OCR not available]"
            
        try:
            # If input is a file path, open the image
            if isinstance(image, str):
                if not os.path.exists(image):
                    raise FileNotFoundError(f"Image file not found: {image}")
                image = Image.open(image)
            
            # If input is a file-like object, open it as an image
            elif hasattr(image, 'read'):
                image = Image.open(image)
            
            # Ensure we have a PIL Image at this point
            if not isinstance(image, Image.Image):
                raise ValueError(f"Expected PIL Image, file path, or file-like object, got {type(image)}")
                
            # Convert image to RGB if it's not already
            if image.mode != 'RGB':
                image = image.convert('RGB')
                
            # Perform OCR - MangaOCR can handle PIL Images directly
            text = self.ocr(image)
            logger.info(f"Transcribed text: {text}")
            return text
            
        except Exception as e:
            error_msg = f"Error in transcribing image: {str(e)}"
            logger.error(error_msg, exc_info=True)  # Log full traceback
            return f"[OCR Error: {str(e)}]"
    def translate_text(self, japanese_text: str) -> str:
        """
        Translate Japanese text to English using Groq LLM.
        
        Args:
            japanese_text: Text in Japanese to translate
            
        Returns:
            str: Translated text or error message if translation fails
        """
        if not self.groq_client or not self.groq_api_key:
            error_msg = "Translation service not available. Please check your API key and internet connection."
            logger.error(error_msg)
            return "[Translation not available]"
            
        try:
            logger.debug(f"Translating text: {japanese_text}")
            
            # Clean the input text
            clean_text = japanese_text.strip()
            if not clean_text:
                return "[No text to translate]"
                
            # System prompt to get just the translation without any additional text
            system_prompt = """
            You are a precise Japanese to English translator. 
            ONLY provide the direct English translation of the Japanese text. 
            Do not include any explanations, notes, or additional text.
            If the input is not in Japanese, return [Not Japanese].
            """
            
            response = self.groq_client.chat.completions.create(
                model="llama3-8b-8192",
                messages=[
                    {"role": "system", "content": system_prompt},
                    {"role": "user", "content": clean_text}
                ],
                temperature=0.1,  # Lower temperature for more deterministic output
                max_tokens=100,
                top_p=1.0
            )
            
            if not response.choices or not response.choices[0].message.content:
                error_msg = "Empty response from translation service"
                logger.error(error_msg)
                return "[Translation error]"
                
            # Clean up the response
            translation = response.choices[0].message.content.strip()
            
            # Remove any quotes or brackets that might have been added
            translation = translation.strip('"\'“”‘’')
            
            # If the translation is too long, it might include extra text
            if len(translation) > len(clean_text) * 3:  # If translation is more than 3x the original length
                # Try to extract just the first sentence
                first_period = translation.find('.')
                if first_period > 0:
                    translation = translation[:first_period + 1]
                else:
                    translation = translation[:100] + "..."  # Truncate if too long
            
            logger.debug(f"Translated '{clean_text}' to: {translation}")
            return translation
            
        except Exception as e:
            error_msg = f"Error in translation: {str(e)}"
            # Try to extract more detailed error information
            if hasattr(e, 'response') and hasattr(e.response, 'text'):
                try:
                    error_data = json.loads(e.response.text)
                    error_msg = f"Error in translation: {error_data.get('error', {}).get('message', str(e))}"
                except Exception as parse_error:
                    error_msg = f"Error in translation: {e.response.text}"
                    logger.debug(f"Could not parse error response: {str(parse_error)}")
            
            logger.error(error_msg, exc_info=True)
            return f"[Translation Error: {error_msg}]"
    
    def _preprocess_text(self, text: str) -> str:
        """
        Preprocess text for comparison by normalizing various aspects.
        """
        import re
        import unicodedata
        
        if not text:
            return ""
            
        # Normalize unicode (e.g., full-width to half-width)
        text = unicodedata.normalize('NFKC', text)
        
        # Remove common OCR artifacts and normalize whitespace
        text = re.sub(r'[\s\u3000]+', ' ', text)  # Convert all whitespace to single space
        text = re.sub(r'[、。]', '', text)  # Remove punctuation that might be inconsistently recognized
        text = text.strip()
        
        return text
    
    def _calculate_similarity(self, text1: str, text2: str) -> float:
        """
        Calculate a simple similarity score between two texts (0.0 to 1.0).
        """
        if not text1 or not text2:
            return 0.0
            
        # Simple character-based similarity
        set1 = set(text1)
        set2 = set(text2)
        if not set1 and not set2:
            return 1.0
        return len(set1.intersection(set2)) / len(set1.union(set2))
    
    def _convert_to_letter_grade(self, score: int) -> str:
        """Convert a numeric score (0-100) to a letter grade.
        
        Grading Scale:
        S: 90-100 - Excellent (Superior)
        A: 80-89  - Very Good
        B: 70-79  - Good
        C: 60-69  - Satisfactory
        D: 50-59  - Needs Improvement
        E: 30-49  - Poor
        F: 0-29   - Fail
        
        Args:
            score: Numeric score from 0 to 100
            
        Returns:
            str: Letter grade (S, A, B, C, D, E, F)
        """
        if not isinstance(score, (int, float)):
            return "F"
            
        if score >= 90:
            return "S"
        elif score >= 80:
            return "A"
        elif score >= 70:
            return "B"
        elif score >= 60:
            return "C"
        elif score >= 50:
            return "D"
        elif score >= 30:
            return "E"
        else:
            return "F"
            
    def grade_writing(self, user_japanese: str, expected_japanese: str) -> Dict[str, Any]:
        """
        Grade the user's Japanese writing against the expected Japanese text.
        
        Args:
            user_japanese: User's Japanese text from OCR
            expected_japanese: Expected Japanese sentence
                
        Returns:
            dict: Dictionary containing grade (letter), feedback, and error information
        """
        if not self.groq_client or not self.groq_api_key:
            error_msg = "Grading service not available. No API key or client."
            logger.error(error_msg)
            return {
                "success": False,
                "error": error_msg,
                "grade": 0,
                "feedback": "Grading service unavailable. Please check your API key and internet connection.",
                "suggestions": ["Check your GROQ_API_KEY environment variable"]
            }
        
        # Preprocess both texts for better comparison
        cleaned_user = self._preprocess_text(user_japanese)
        cleaned_expected = self._preprocess_text(expected_japanese)
        
        # Calculate basic similarity metrics
        similarity_score = self._calculate_similarity(cleaned_user, cleaned_expected)
        
        try:
            logger.debug(f"Grading Japanese text. Expected: '{expected_japanese}' (cleaned: '{cleaned_expected}'), Got: '{user_japanese}' (cleaned: '{cleaned_user}')")
            
            # Prepare the prompt for grading Japanese text with English feedback
            prompt = f"""You are a Japanese language teacher evaluating a student's handwritten Japanese.
            
            TASK: Compare the student's writing with the expected Japanese text and provide a detailed evaluation.
            
            EXPECTED JAPANESE: {expected_japanese}
            STUDENT'S WRITING: {user_japanese}
            
            IMPORTANT: The student's writing comes from OCR and may have recognition errors.
            Be forgiving of minor spacing or similar-looking character mistakes that could be OCR errors.
            
            EVALUATION CRITERIA:
            1. Overall Accuracy (50 points):
               - How closely does it match the expected text?
               - Account for likely OCR errors (e.g., つ vs っ, し vs じ, etc.)
            
            2. Grammar and Structure (30 points):
               - Correct particle usage
               - Proper verb forms and conjugations
               - Sentence structure and word order
            
            3. Character Accuracy (20 points):
               - Correct kanji/hiragana/katakana usage
               - Proper character forms (be lenient with handwriting variations)
            
            INSTRUCTIONS:
            - First, analyze if the student's writing conveys the same meaning as the expected text
            - Then, examine the specific differences and determine if they're likely OCR errors or actual mistakes
            - Provide specific feedback about any errors or areas for improvement
            - Be encouraging but accurate in your assessment
            
            RESPONSE FORMAT (JSON):
            {{
                "grade": 0-100,
                "feedback": "Detailed feedback in English explaining the evaluation",
                "suggestions": ["Specific suggestion 1", "Specific suggestion 2"],
                "is_ocr_issue": true/false  # Whether the main issues appear to be OCR-related
            }}"""
                
            logger.debug("Sending grading request to Groq API...")
            response = self.groq_client.chat.completions.create(
                model="llama3-8b-8192",
                messages=[
                    {"role": "system", "content": """You are a patient and encouraging Japanese language teacher. 
                    Provide clear, specific feedback in English. Be understanding of potential OCR errors in the student's writing.
                    Focus on both what was done well and areas for improvement."""},
                    {"role": "user", "content": prompt}
                ],
                temperature=0.3,  # Slightly higher temperature for more nuanced feedback
                max_tokens=1500,  # Increased for more detailed feedback
                response_format={"type": "json_object"}
            )
                
            if not response.choices or not response.choices[0].message.content:
                error_msg = "Empty response from grading service"
                logger.error(error_msg)
                raise ValueError(error_msg)
                
            # Extract and parse the JSON response
            try:
                result = json.loads(response.choices[0].message.content)
                logger.debug(f"Received grading result: {result}")
                
                # Adjust grade based on similarity score if there are likely OCR issues
                if result.get('is_ocr_issue', False):
                    adjusted_grade = (float(result.get('grade', 0)) + (similarity_score * 20)) / 1.2
                    result['grade'] = min(100, max(0, int(adjusted_grade)))
                    result['feedback'] += "\n\nNote: Some leniency was applied for likely OCR recognition errors."
                
                # Validate the response structure
                if not all(key in result for key in ["grade", "feedback", "suggestions"]):
                    raise ValueError("Invalid response format from grading service")
                    
                # Get numeric score and convert to letter grade
                numeric_grade = int(result.get("grade", 0))
                letter_grade = self._convert_to_letter_grade(numeric_grade)
                
                return {
                    "success": True,
                    "grade": letter_grade,
                    "numeric_grade": numeric_grade,  # Keep numeric grade for reference
                    "feedback": result.get("feedback", "No feedback provided").strip(),
                    "suggestions": [s.strip() for s in result.get("suggestions", []) if s.strip()],
                    "ocr_adjusted": result.get('is_ocr_issue', False)
                }
                    
            except json.JSONDecodeError as je:
                error_msg = f"Failed to parse grading response: {str(je)}\nResponse: {response.choices[0].message.content}"
                logger.error(error_msg)
                raise ValueError("Invalid JSON response from grading service") from je
                    
        except Exception as e:
            error_msg = f"Error in grading: {str(e)}"
            logger.error(error_msg, exc_info=True)
                
            # Provide more specific error messages for common issues
            if "401" in str(e):
                error_detail = "Authentication failed. Please check your GROQ_API_KEY."
            elif "rate limit" in str(e).lower():
                error_detail = "API rate limit exceeded. Please try again later."
            else:
                error_detail = "An unexpected error occurred."
                    
            return {
                "success": False,
                "error": error_msg,
                "grade": 0,
                "feedback": f"Grading failed. {error_detail}",
                "suggestions": [
                    "Check your internet connection",
                    "Verify your API key is valid and has sufficient credits",
                    "Try again in a few moments"
                ]
            }
    
    def process_submission(self, image: Image.Image, expected_sentence: str) -> Dict[str, Any]:
        """
        Process a complete submission: OCR -> Translation -> Grading.
            
        Args:
            image: Image containing handwritten Japanese text
            expected_sentence: Expected Japanese sentence in English
                
        Returns:
            dict: Dictionary containing all results and error information
        """
        logger.info("Starting submission processing...")
        result = {
            "success": False,
            "transcription": "",
            "translation": "",
            "grade": 0,
            "feedback": "",
            "suggestions": [],
            "error": None,
            "warnings": []
        }
        
        # Validate input
        if not image:
            result["error"] = "No image provided for processing"
            logger.error(result["error"])
            return result
                
        if not expected_sentence or not expected_sentence.strip():
            result["error"] = "No expected sentence provided for comparison"
            logger.error(result["error"])
            return result
            
        # Step 1: Transcribe the image
        logger.debug("Starting image transcription...")
        try:
            # Log image details before transcription
            logger.debug(f"Image details - Format: {image.format if hasattr(image, 'format') else 'N/A'}, Size: {image.size if hasattr(image, 'size') else 'N/A'}, Mode: {image.mode if hasattr(image, 'mode') else 'N/A'}")
            
            # Perform the transcription
            transcription_start_time = time.time()
            result["transcription"] = self.transcribe_image(image)
            transcription_time = time.time() - transcription_start_time
            
            logger.info(f"MangaOCR Transcription completed in {transcription_time:.2f} seconds")
            
            if not result["transcription"]:
                result["error"] = "Transcription returned empty result"
                logger.error(result["error"])
                return result
            
            # Log the raw transcription result
            logger.info(f"Raw transcription output: {result['transcription']}")
            
            # Check for potential issues in transcription
            if result["transcription"].startswith("["):
                warning_msg = f"Possible error in transcription: {result['transcription']}"
                result["warnings"].append(warning_msg)
                logger.warning(warning_msg)
            else:
                logger.info(f"Successfully transcribed text: {result['transcription']}")
                
            # Add debug information to the result
            result["transcription_metadata"] = {
                "processing_time_seconds": transcription_time,
                "characters_recognized": len(result["transcription"]) if result["transcription"] else 0
            }
                    
        except Exception as e:
            error_msg = f"Error in transcription: {str(e)}"
            logger.error(error_msg, exc_info=True)
            result["error"] = error_msg
            return result
            
        # Step 2: Translate the transcribed text
        logger.debug("Starting text translation...")
        try:
            result["translation"] = self.translate_text(result["transcription"])
                
            if not result["translation"]:
                result["error"] = "Translation returned empty result"
                logger.error(result["error"])
                return result
                    
            if result["translation"].startswith("["):
                result["warnings"].append("Possible error in translation")
                logger.warning(f"Translation may have issues: {result['translation']}")
                    
            logger.info(f"Translation successful: {result['translation']}")
                    
        except Exception as e:
            error_msg = f"Error in translation: {str(e)}"
            logger.error(error_msg, exc_info=True)
            result["error"] = error_msg
            return result
            
        # Step 3: Grade the Japanese writing
        logger.debug("Starting grading...")
        try:
            # Use the Japanese transcription and expected Japanese for grading
            grade_result = self.grade_writing(
                user_japanese=result["transcription"].strip(),
                expected_japanese=expected_sentence.strip()
            )
                
            if not grade_result.get("success", False):
                result["error"] = grade_result.get("error", "Grading failed")
                logger.error(f"Grading failed: {result['error']}")
                return result
                    
            # Update result with grading information
            numeric_grade = grade_result.get("numeric_grade", 0)
            result.update({
                "success": True,
                "grade": grade_result.get("grade", "F"),  # Letter grade (S, A, B, C, D, E, F)
                "numeric_grade": min(100, max(0, int(numeric_grade))),  # Ensure numeric grade is 0-100
                "feedback": grade_result.get("feedback", "No feedback provided").strip(),
                "suggestions": [s for s in grade_result.get("suggestions", []) if s]
            })
                
            logger.info(f"Grading complete. Grade: {result['grade']} (Score: {result['numeric_grade']}/100)")
                
        except Exception as e:
            error_msg = f"Error in grading: {str(e)}"
            logger.error(error_msg, exc_info=True)
            result["error"] = error_msg
            
        # Clean up any empty fields
        result["suggestions"] = [s for s in result["suggestions"] if s]
        if not result["suggestions"] and not result["error"]:
            result["suggestions"] = ["Great job! Keep practicing!"]
                
        logger.info("Submission processing completed successfully")
        return result
