import streamlit as st
from typing import Dict
import json
from collections import Counter
import re
import os
from backend.groq_chat import GroqChat
from backend.get_transcript import YouTubeTranscriptDownloader

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

def render_question(question):
    """Render a question with options"""
    # Display the question text with proper formatting
    question_text = question.get('question', '')
    
    # Split into introduction, conversation, and question parts
    parts = [p.strip() for p in question_text.split('\n\n') if p.strip()]
    
    if len(parts) >= 3:
        # Keep introduction in English
        intro = parts[0].replace('Introduction:', '').strip()
        
        # Process conversation lines
        conversation = parts[1].replace('Conversation:', '').strip()
        # Split into lines and ensure each line starts on a new line
        conversation_lines = [line.strip() for line in conversation.split('\n') if line.strip()]
        formatted_conversation = '\n'.join(conversation_lines)
        
        # Clean up question part
        question_part = parts[2].replace('Question:', '').replace('質問:', '').strip()
        
        st.subheader("Listening Practice")
        st.write(f"**Situation:** {intro}")
        st.write("**Dialogue:**")
        st.text(formatted_conversation)  # Use st.text to preserve newlines
        st.write(f"**Question:** {question_part}")
    else:
        # Fallback if format doesn't match
        st.subheader("Question")
        st.write(question_text)
    
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
    
    # Add a button to try another question with unique key
    if st.button("Next Question", key="next_question_btn"):
        # Clear the current question and feedback state
        st.session_state.current_question = None
        st.session_state.user_answer = None
        st.session_state.show_feedback = False
        # Generate a new question
        generate_new_question(question_type, topic)
        st.rerun()

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
            question = question_generator.generate_question(
                question_type=question_type,
                topic=topic if topic else None
            )
            
            if question:
                st.session_state.current_question = question
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

def main():
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