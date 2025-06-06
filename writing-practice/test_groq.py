import os
from dotenv import load_dotenv
from groq import Groq

# Load environment variables
load_dotenv()

# Get API key
api_key = os.getenv('GROQ_API_KEY')
print(f"API Key starts with: {api_key[:8]}..." if api_key else "No API key found")

if not api_key:
    print("Error: GROQ_API_KEY not found in environment variables")
    exit(1)

try:
    # Initialize the Groq client
    client = Groq(api_key=api_key)
    
    # Test the API with a simple request
    print("Sending test request to Groq API...")
    completion = client.chat.completions.create(
        model="llama3-8b-8192",
        messages=[
            {"role": "system", "content": "You are a helpful assistant."},
            {"role": "user", "content": "Say hello in Japanese"}
        ],
        temperature=0.5,
        max_tokens=100,
        top_p=1,
        stream=False,
        stop=None,
    )
    
    print("\nResponse received:")
    print(completion.choices[0].message.content)
    print("\nAPI key is working correctly!")
    
except Exception as e:
    print(f"\nError occurred: {str(e)}")
    if "401" in str(e):
        print("\nError: Unauthorized - The API key is invalid or doesn't have the correct permissions.")
    elif "rate limit" in str(e).lower():
        print("\nError: Rate limit exceeded. Please try again later.")
    else:
        print("\nAn unexpected error occurred. Please check your network connection and try again.")
