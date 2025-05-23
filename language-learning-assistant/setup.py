from setuptools import setup, find_packages

setup(
    name="language-learning-assistant",
    version="0.1",
    packages=find_packages(),
    install_requires=[
        'streamlit',
        'boto3',
        'youtube-transcript-api',
        'chromadb',
        'python-dotenv',
    ],
)
