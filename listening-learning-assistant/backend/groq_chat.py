import os
from dotenv import load_dotenv
from groq import Groq
from typing import Optional, Dict, Any

# Load environment variables from .env file
load_dotenv()

# Model ID
MODEL_ID = "qwen-qwq-32b"  # or 'llama2-70b-4096'

class GroqChat:
    def __init__(self, model_id: str = MODEL_ID):
        """Initialize Groq chat client"""
        api_key = os.getenv("GROQ_API_KEY")
        if not api_key:
            raise ValueError("GROQ_API_KEY not found in environment variables")
            
        self.client = Groq(api_key=api_key)
        self.model_id = model_id

    def generate_response(self, message: str, temperature: float = 0.7) -> Optional[str]:
        """Generate a response using Groq"""
        try:
            completion = self.client.chat.completions.create(
                messages=[{
                    "role": "user",
                    "content": message
                }],
                model=self.model_id,
                temperature=temperature,
            )
            return completion.choices[0].message.content
            
        except Exception as e:
            print(f"Error generating response: {str(e)}")
            return None

# Example usage
if __name__ == "__main__":
    chat = GroqChat()
    while True:
        user_input = input("You: ")
        if user_input.lower() in ['/exit', 'quit', 'exit']:
            break
        response = chat.generate_response(user_input)
        print("Bot:", response)