# Free GenAI Bootcamp 2025

A collection of language learning tools powered by AI, developed during the Free GenAI Bootcamp 2025.

## 🚀 Features

### 1. Listening Learning Assistant
An interactive tool that helps language learners improve their listening and comprehension skills through AI-generated questions and answers.

### 2. Lang Portal
A centralized portal for accessing various language learning resources and tools.

### 3. Vocabulary Importer
A tool to import and manage vocabulary lists for language learning.

## 🛠️ Setup

1. **Prerequisites**
   - Python 3.8+
   - pip (Python package manager)
   - Virtual environment (recommended)

2. **Installation**
   ```bash
   # Clone the repository
   git clone https://github.com/yourusername/free-genai-bootcamp-2025.git
   cd free-genai-bootcamp-2025
   
   # Create and activate virtual environment
   python -m venv .venv
   source .venv/bin/activate  # On Windows: .venv\Scripts\activate
   
   # Install dependencies
   pip install -r requirements.txt
   ```

3. **Environment Variables**
   Create a `.env` file in the project root with the following variables:
   ```
   OPENAI_API_KEY=your_openai_api_key
   ```

## 🚀 Running the Application

### Listening Learning Assistant
```bash
# Navigate to the assistant directory
cd listening-learning-assistant

# Run the Streamlit app
/Users/janalexanderneumann/Dev/free-genai-bootcamp-2025/.venv/bin/streamlit run main.py
```

## 📝 Project Structure

```
free-genai-bootcamp-2025/
├── listening-learning-assistant/  # Listening comprehension tool
│   ├── main.py                   # Main Streamlit application
│   ├── backend/                  # Backend logic and API handlers
│   └── ...
├── lang-portal/                  # Language learning portal
├── vocabulary-importer/          # Vocabulary management tool
├── .env.example                 # Example environment variables
├── requirements.txt             # Python dependencies
└── README.md                    # This file
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

Built with ❤️ during the Free GenAI Bootcamp 2025
