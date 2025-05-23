from youtube_transcript_api import YouTubeTranscriptApi
from typing import Optional, List, Dict
import os

class YouTubeTranscriptDownloader:
    def __init__(self, languages: List[str] = ["ja", "en"]):
        self.languages = languages

    def extract_video_id(self, url: str) -> Optional[str]:
        """
        Extract video ID from YouTube URL
        
        Args:
            url (str): YouTube URL
            
        Returns:
            Optional[str]: Video ID if found, None otherwise
        """
        if "v=" in url:
            return url.split("v=")[1][:11]
        elif "youtu.be/" in url:
            return url.split("youtu.be/")[1][:11]
        return None

    def get_transcript(self, video_id: str) -> Optional[list]:
        """
        Download YouTube Transcript
        
        Args:
            video_id (str): YouTube video ID or URL
            
        Returns:
            List of transcript entries or None if failed
        """
        if "youtube.com" in video_id or "youtu.be" in video_id:
            video_id = self.extract_video_id(video_id)
        
        if not video_id:
            print("Invalid video ID or URL")
            return None
    
        print(f"Downloading transcript for video ID: {video_id}")
    
        try:
            # Try direct approach first
            try:
                return YouTubeTranscriptApi.get_transcript(
                    video_id, 
                    languages=self.languages
                )
            except Exception as e:
                print(f"Direct approach failed: {str(e)}")
                
            # If direct approach fails, try listing available transcripts
            try:
                transcript_list = list(YouTubeTranscriptApi.list_transcripts(video_id))
                print("Available transcripts:")
                for t in transcript_list:
                    print(f"- {t.language} ({t.language_code})")
                
                # Try to find a working transcript
                for transcript in transcript_list:
                    try:
                        print(f"Trying {transcript.language} transcript...")
                        return transcript.fetch()
                    except Exception as e:
                        print(f"Failed to fetch {transcript.language}: {str(e)}")
                        continue
                        
            except Exception as e:
                print(f"Error listing transcripts: {str(e)}")
                
        except Exception as e:
            print(f"Error getting transcript: {str(e)}")
        
        print("No working transcript found")
        return None

    def save_transcript(self, transcript: list, filename: str) -> bool:
        """
        Save transcript to file
        
        Args:
            transcript: List of transcript entries
            filename: Base filename (without path or extension)
            
        Returns:
            bool: True if successful, False otherwise
        """
        if not transcript:
            print("No transcript data to save")
            return False
    
        os.makedirs("transcripts", exist_ok=True)
        filepath = os.path.join("transcripts", f"{filename}.txt")
        
        try:
            with open(filepath, 'w', encoding='utf-8') as f:
                for entry in transcript:
                    f.write(f"{entry['text']}\n")
            print(f"Transcript saved to {filepath}")
            return True
        except Exception as e:
            print(f"Error saving transcript: {str(e)}")
            return False

def main(video_url, print_transcript=False):
    # Initialize downloader
    downloader = YouTubeTranscriptDownloader()
    
    # Get transcript
    video_id = downloader.extract_video_id(video_url) if "youtube.com" in video_url or "youtu.be" in video_url else video_url
    transcript = downloader.get_transcript(video_id)
    
    if transcript:
        # Save transcript
        if downloader.save_transcript(transcript, video_id):
            print(f"Transcript saved successfully to transcripts/{video_id}.txt")
            if print_transcript:
                for entry in transcript:
                    print(entry['text'])
        else:
            print("Failed to save transcript")
    else:
        print("Failed to get transcript")

if __name__ == "__main__":
    video_id = "https://www.youtube.com/watch?v=sY7L5cfCWno&list=PLkGU7DnOLgRMl-h4NxxrGbK-UdZHIXzKQ"  # Extract from URL: XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
    transcript = main(video_id, print_transcript=True)
        