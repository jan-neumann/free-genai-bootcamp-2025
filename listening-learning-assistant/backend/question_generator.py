import os
import sys
from pathlib import Path
import re
import json
from typing import List, Dict, Optional
from groq import Groq
from dotenv import load_dotenv
import logging
import random

# Use absolute import
from backend.vector_store import initialize_vector_store

# Load environment variables
load_dotenv()

# Configure logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')

class QuestionGenerator:
    PROMPT_TEMPLATE = """
    Create a JLPT N5 level question in Japanese.

    IMPORTANT: Your response must ONLY be a single, valid JSON object. 
    Do NOT include any of your thought process, reasoning, explanations, or any text whatsoever outside of the JSON object itself. 
    The JSON object must contain the following five string/list keys: "introduction", "conversation", "question", "options", and "correct_answer_letter".
    - "introduction", "conversation", "question", and "correct_answer_letter" must be strings.
    - "options" must be a list of four strings.
    The response must start with '{{' and end with '}}', be parsable by Python's json.loads(), and be a complete, well-formed JSON structure.

    Example of the required JSON structure:
    <JSON>
    {{
        "introduction": "男の人と女の人が話しています。",
        "conversation": "女：今朝のニュース見た？\\n男：ううん、まだ見てないよ。何かあったの？\\n女：駅の近くに新しいレストランができたって。",
        "question": "二人は何について話していますか？",
        "options": ["レストランができたこと", "天気のこと", "仕事のこと", "週末の予定"],
        "correct_answer_letter": "A"
    }}
    </JSON>

    Guidelines for the content of the JSON fields:
    1. Use only hiragana, katakana, and basic kanji (JLPT N5 level).
    2. Keep the conversation natural, simple, and appropriate for N5 listening practice.
    3. Use ONLY Japanese characters (hiragana, katakana, N5 kanji, and standard Japanese punctuation) within the Japanese text values. Do NOT use English, Chinese, or any other non-Japanese characters in these values.
    4. The conversation should be about: {topic}
    5. JSON SYNTAX IS CRITICAL: All string values must be perfectly enclosed in double quotes ("). Ensure commas correctly separate key-value pairs. "options" must be a JSON array of strings.
    6. Provide four distinct multiple-choice options for the question.
    7. Indicate the correct answer by providing its letter (A, B, C, or D) as the value for "correct_answer_letter".

    Return ONLY the JSON object enclosed between <JSON> and </JSON> tags.
    """

    def __init__(self, model_name: str = "deepseek-r1-distill-llama-70b"):
        """Initialize the question generator with Groq client and vector store"""
        api_key = os.getenv("GROQ_API_KEY")
        if not api_key:
            logging.error("GROQ_API_KEY environment variable not set")
            raise ValueError("GROQ_API_KEY environment variable not set")
            
        self.client = Groq(api_key=api_key)
        self.model_name = model_name
        self.vector_store = None
        try:
            self.vector_store = initialize_vector_store()
            logging.info("Vector store initialized successfully.")
        except Exception as e:
            logging.warning(f"Could not initialize vector store: {str(e)}")

    @staticmethod
    def _clean_text(text: str) -> str:
        """Clean text by removing unwanted tags and normalizing whitespace."""
        if not isinstance(text, str):
            return str(text)
        # Remove <think>...</think> tags and their content
        text = re.sub(r'<think>.*?</think>', '', text, flags=re.DOTALL)
        # Remove any other HTML-like tags but keep their content
        text = re.sub(r'<[^>]+>', '', text) 
        # Remove leading colons/spaces that might appear from model responses
        text = re.sub(r'^[：:\s]+', '', text)
        # Normalize whitespace: replace multiple spaces/newlines with a single space, then strip
        text = ' '.join(text.split())
        return text.strip()

    def generate_question(self, question_type: str, topic: str = None) -> Optional[Dict]:
        """
        Generate a question using RAG with vector store context.

        Args:
            question_type: Type of question to generate (e.g., 'Dialogue', 'Listening').
            topic: Optional topic to guide question generation.

        Returns:
            A dictionary with 'introduction', 'conversation', 'question' keys if successful,
            plus 'options', 'correct_answer', 'explanation', and 'raw_response'.
            Returns None if generation or parsing fails.
        """
        # logging.info(f"Starting question generation. Type: {question_type}, Topic: {topic or 'Not specified'}")
        
        effective_topic = topic if topic else 'everyday situations'
        final_prompt = self.PROMPT_TEMPLATE.format(topic=effective_topic)
        # logging.info(f"Formatted prompt for model:\n{final_prompt}")

        raw_response_content = None
        try:
            # logging.info(f"Calling Groq API with model: {self.model_name}")
            response = self.client.chat.completions.create(
                model=self.model_name,
                messages=[
                    {"role": "system", "content": "You are a Japanese language teaching assistant for JLPT N5 level. Your entire response MUST be a single, valid, complete JSON object as per the user's instructions, starting with '{{' and ending with '}}'. Do not include any other text, explanations, or thought processes outside this JSON object."},
                    {"role": "user", "content": final_prompt}
                ],
                temperature=0.3,
                max_tokens=1024 
            )

            if not response or not response.choices or not response.choices[0].message or not response.choices[0].message.content:
                logging.error("Empty or invalid response from API.")
                # logging.debug(f"Full API Response: {response}")
                return None
            
            raw_response_content = response.choices[0].message.content.strip()
            logging.info(f"Raw response content from API:\n{raw_response_content}")

        except Exception as e:
            logging.error(f"Error calling Groq API: {type(e).__name__} - {str(e)}")
            # import traceback # Uncomment for detailed stack trace
            # logging.debug(traceback.format_exc())
            return None

        if not raw_response_content:
            logging.error("No content received from API call.")
            return None

        parsed_data_dict = None
        json_str_to_parse = ""
        try:
            # Strategy 1: Look for content between <JSON> and </JSON> tags
            json_match = re.search(r'<JSON>(.*?)</JSON>', raw_response_content, re.DOTALL)
            if json_match:
                json_str_to_parse = json_match.group(1).strip()
                # logging.info(f"Extracted JSON string from <JSON> tags: {json_str_to_parse}")
            else:
                # Strategy 2: If no tags, assume the entire response is the JSON string
                # (as per stricter prompt requirements)
                # logging.info("No <JSON> tags found, assuming entire response is JSON.")
                json_str_to_parse = raw_response_content

            # Clean common non-JSON artifacts like ```json ... ``` that models sometimes add
            json_str_to_parse = re.sub(r'^```json\s*', '', json_str_to_parse, flags=re.IGNORECASE)
            json_str_to_parse = re.sub(r'\s*```$', '', json_str_to_parse)
            json_str_to_parse = json_str_to_parse.strip()

            # If after stripping tags, it doesn't start with '{', try to find the first '{'
            if not json_str_to_parse.startswith('{'):
                first_brace = json_str_to_parse.find('{')
                if first_brace != -1:
                    logging.info(f"QGEN: JSON string did not start with '{{'. Stripping leading content up to first '{{'. Original start: '{json_str_to_parse[:50]}...'" )
                    json_str_to_parse = json_str_to_parse[first_brace:]
                else:
                    logging.error(f"QGEN: No '{{' found in json_str_to_parse after tag stripping. Content: {json_str_to_parse}")
                    # Fall through, json.loads will likely fail and be caught

            # Heuristic: If the string starts with '{' and doesn't end with '}',
            # and isn't just an empty opening brace, try appending '}'
            if json_str_to_parse.startswith('{') and not json_str_to_parse.endswith('}') and len(json_str_to_parse.strip()) > 1:
                # Also try to find the last '}' in case of trailing garbage
                last_brace = json_str_to_parse.rfind('}')
                if last_brace != -1 and last_brace > json_str_to_parse.find('{'): # Ensure last brace is after first brace
                    json_str_to_parse = json_str_to_parse[:last_brace+1]
                    logging.info(f"QGEN: Attempting to fix potentially truncated/extended JSON by trimming to last '}}'. Result: '{json_str_to_parse[:50]}...{json_str_to_parse[-50:]}'")
                else:
                    logging.info("QGEN: Attempting to fix potentially truncated JSON by appending '}'.")
                    json_str_to_parse += '}'

            logging.info(f"QGEN: Attempting to parse final JSON string: '{json_str_to_parse}'")
            # Attempt to parse the (potentially cleaned) JSON string
            data = json.loads(json_str_to_parse)
            
            required_keys = ['introduction', 'conversation', 'question', 'options', 'correct_answer_letter']
            if isinstance(data, dict) and all(k in data for k in required_keys):
                options_ok = isinstance(data.get('options'), list) and len(data.get('options')) == 4 and all(isinstance(opt, str) for opt in data.get('options'))
                letter_ok = isinstance(data.get('correct_answer_letter'), str) and data.get('correct_answer_letter', '').upper() in ['A', 'B', 'C', 'D']
                
                if options_ok and letter_ok:
                    parsed_data_dict = data
                    # logging.info("Successfully parsed JSON data with all required keys and valid option/answer format.")
                else:
                    if not options_ok:
                        logging.warning(f"Parsed JSON 'options' field is not a list of 4 strings. Found: {data.get('options')}")
                    if not letter_ok:
                        logging.warning(f"Parsed JSON 'correct_answer_letter' is not a valid letter (A-D). Found: {data.get('correct_answer_letter')}")
            else:
                missing_keys = [k for k in required_keys if k not in (data.keys() if isinstance(data, dict) else [])]
                logging.warning(f"Parsed JSON does not contain all required keys. Missing: {missing_keys}. Found: {list(data.keys()) if isinstance(data, dict) else 'Not a dict'}")
                # logging.debug(f"Problematic JSON data: {data}")

        except json.JSONDecodeError as e:
            logging.error(f"JSONDecodeError: {e}. Problematic JSON string: '{json_str_to_parse}'")
            # Strategy 3: Fallback - try to find the first '{' and last '}' to isolate a potential JSON object
            if json_str_to_parse: 
                match_obj = re.search(r'{.*}', json_str_to_parse, re.DOTALL) 
                if match_obj:
                    json_substring = match_obj.group(0)
                    # logging.info(f"Attempting to parse extracted JSON substring: {json_substring}")
                    try:
                        data = json.loads(json_substring)
                        required_keys_fallback = ['introduction', 'conversation', 'question', 'options', 'correct_answer_letter']
                        if isinstance(data, dict) and all(k in data for k in required_keys_fallback):
                            options_ok_fallback = isinstance(data.get('options'), list) and len(data.get('options')) == 4 and all(isinstance(opt, str) for opt in data.get('options'))
                            letter_ok_fallback = isinstance(data.get('correct_answer_letter'), str) and data.get('correct_answer_letter', '').upper() in ['A', 'B', 'C', 'D']
                            if options_ok_fallback and letter_ok_fallback:
                                parsed_data_dict = data
                                # logging.info("Successfully parsed JSON substring with all required keys and valid option/answer format.")
                            else:
                                if not options_ok_fallback:
                                    logging.warning(f"Substring JSON 'options' field is not a list of 4 strings. Found: {data.get('options')}")
                                if not letter_ok_fallback:
                                    logging.warning(f"Substring JSON 'correct_answer_letter' is not a valid letter (A-D). Found: {data.get('correct_answer_letter')}")
                        else:
                            missing_keys_fallback = [k for k in required_keys_fallback if k not in (data.keys() if isinstance(data, dict) else [])]
                            logging.warning(f"Substring JSON does not contain all required keys after fallback. Missing: {missing_keys_fallback}. Found: {list(data.keys()) if isinstance(data, dict) else 'Not a dict'}")
                    except json.JSONDecodeError as sub_e:
                        logging.error(f"Failed to parse JSON substring after fallback: {sub_e}")
                else:
                    logging.warning("No parsable JSON object found in the string using fallback regex.")
            else:
                logging.warning("JSON string to parse was empty before attempting fallback.")
        except Exception as e: # Catch-all for other unexpected errors during parsing
            logging.error(f"An unexpected error occurred during JSON processing: {type(e).__name__} - {str(e)}")
            # import traceback # Uncomment for detailed stack trace
            # logging.debug(traceback.format_exc())
            return None # Return None if any other error occurs during parsing

        if not parsed_data_dict:
            logging.error("Failed to extract valid JSON data with required keys from the model's response after all strategies.")
            # logging.debug(f"Final raw response content that failed parsing: {raw_response_content}")
            return None
            
        try:
            intro = self._clean_text(parsed_data_dict.get('introduction', ''))
            
            # Conversation text: json.loads converts \n in JSON string to \n in Python string.
            # We need to preserve these newlines for display.
            # _clean_text normalizes whitespace, which would replace \n with a space.
            # So, we split by \n, clean each individual line, then rejoin with \n.
            conversation_raw = parsed_data_dict.get('conversation', '') 
            conversation_lines = [self._clean_text(line) for line in conversation_raw.split('\n')]
            conversation_cleaned = '\n'.join(filter(None, conversation_lines)) # filter(None, ...) removes empty strings

            question_text = self._clean_text(parsed_data_dict.get('question', ''))

            # logging.info(f"Cleaned data - Intro: '{intro}', Convo: '{conversation_cleaned}', Question: '{question_text}'")

            # Get options, default to empty strings if missing or not a list
            options_raw = parsed_data_dict.get('options', [])
            if not isinstance(options_raw, list) or len(options_raw) != 4:
                logging.warning(f"'options' field is not a list of 4 elements. Received: {options_raw}. Using placeholders.")
                options_raw = ["Option A placeholder", "Option B placeholder", "Option C placeholder", "Option D placeholder"]
            
            cleaned_options = [self._clean_text(str(opt)) for opt in options_raw] # Ensure opt is str before cleaning

            raw_letter_from_parsed = parsed_data_dict.get('correct_answer_letter', '')
            logging.info(f"QGEN: Raw correct_answer_letter from parsed_data_dict: '{raw_letter_from_parsed}' (type: {type(raw_letter_from_parsed).__name__})")

            cleaned_correct_answer_letter = self._clean_text(raw_letter_from_parsed).upper()
            logging.info(f"QGEN: correct_answer_letter after _clean_text and .upper(): '{cleaned_correct_answer_letter}'")

            # Validate correct_answer_letter again after cleaning
            if cleaned_correct_answer_letter not in ['A', 'B', 'C', 'D']:
                logging.warning(f"QGEN: Cleaned correct_answer_letter '{cleaned_correct_answer_letter}' is invalid. Setting to empty string.")
                cleaned_correct_answer_letter = ""
        
            logging.info(f"QGEN: Final cleaned_correct_answer_letter before putting in dict: '{cleaned_correct_answer_letter}'")

            # Ensure options are a list of strings
            if 0 <= ord(cleaned_correct_answer_letter) - ord('A') < len(cleaned_options):
                correct_answer_text = cleaned_options[ord(cleaned_correct_answer_letter) - ord('A')]
            else:
                # Fallback if correct_idx is somehow out of bounds (shouldn't happen with current logic)
                logging.warning(f"Initial correct_idx out of bounds for options length {len(cleaned_options)}. Defaulting to first option as correct.")
                correct_answer_text = cleaned_options[0] if cleaned_options else ""
                correct_idx = 0

            # Shuffle the options
            random.shuffle(cleaned_options)

            # Find the new index of the correct answer text in the shuffled list
            try:
                final_correct_idx = cleaned_options.index(correct_answer_text)
            except ValueError:
                # This should ideally not happen if correct_answer_text was in cleaned_options
                logging.error(f"Correct answer text '{correct_answer_text}' not found in shuffled options. Defaulting to 0.")
                final_correct_idx = 0

            final_data = {
                'introduction': intro,
                'conversation': conversation_cleaned,
                'question': question_text,
                'options': cleaned_options, 
                'correct_answer_letter': cleaned_correct_answer_letter, # Ensure this is included!
                'correct_answer': final_correct_idx, # This is the index, keep for compatibility if used elsewhere
                'explanation': "", # Placeholder for now
                'raw_response': raw_response_content  # Include the raw response for debugging
            }
            logging.info(f"QGEN: Returning final_data with correct_answer_letter: '{final_data.get('correct_answer_letter')}' and correct_answer_index: {final_data.get('correct_answer')}")
            return final_data
        except Exception as e:
            logging.error(f"Error processing parsed data into final dictionary: {type(e).__name__} - {str(e)}")
            # import traceback # Uncomment for detailed stack trace
            # logging.debug(traceback.format_exc())
            return None

# Singleton instance
question_generator = QuestionGenerator()