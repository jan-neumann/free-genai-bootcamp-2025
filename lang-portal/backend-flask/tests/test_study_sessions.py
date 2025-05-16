import pytest
import json
import os
from app import create_app # Assuming pytest is run from backend-flask directory
from datetime import datetime, timedelta, UTC

@pytest.fixture
def app():
    """Create and configure a new app instance for each test."""
    base_dir = os.path.dirname(os.path.abspath(__file__))
    project_root = os.path.dirname(base_dir)
    db_path = os.path.join(project_root, 'test_study_sessions_module.db') # Unique DB name
    
    _app = create_app({
        'TESTING': True,
        'DATABASE': db_path
    })

    with _app.app_context():
        db_instance = _app.db
        conn = db_instance.get()
        cursor = conn.cursor()
        sql_setup_dir = os.path.join(project_root, 'sql', 'setup')

        table_files = [
            'create_table_groups.sql',
            'create_table_study_activities.sql',
            'create_table_words.sql',
            'create_table_word_groups.sql',
            'create_table_study_sessions.sql',
            'create_table_word_review_items.sql',
            'create_table_word_reviews.sql'
        ]
        for table_file in table_files:
            with open(os.path.join(sql_setup_dir, table_file), 'r') as f:
                conn.executescript(f.read())
        
        # Seed data
        cursor.execute("INSERT INTO groups (id, name, words_count) VALUES (?, ?, ?)", (1, 'Test Group 1', 2))
        cursor.execute("INSERT INTO groups (id, name, words_count) VALUES (?, ?, ?)", (2, 'Test Group 2', 1))
        cursor.execute("INSERT INTO study_activities (id, name, url, preview_url) VALUES (?, ?, ?, ?)", (1, 'Flashcards', 'http://example.com/flashcards', 'http://example.com/flash_preview.jpg'))
        cursor.execute("INSERT INTO study_activities (id, name, url, preview_url) VALUES (?, ?, ?, ?)", (2, 'Quiz', 'http://example.com/quiz', 'http://example.com/quiz_preview.jpg'))

        words_data = [
            (1, '単語1', 'tango1', 'word1', '{"type": "noun"}'),
            (2, '単語2', 'tango2', 'word2', '{"type": "verb"}'),
            (3, '単語3', 'tango3', 'word3', '{"type": "adj"}')
        ]
        cursor.executemany("INSERT INTO words (id, kanji, romaji, english, parts) VALUES (?, ?, ?, ?, ?)", words_data)

        word_groups_data = [
            (1, 1), # word1 in group1
            (2, 1), # word2 in group1
            (3, 2)  # word3 in group2
        ]
        cursor.executemany("INSERT INTO word_groups (word_id, group_id) VALUES (?, ?)", word_groups_data)

        # Timestamps for study sessions
        now = datetime.now(UTC)
        time1 = (now - timedelta(days=2)).strftime('%Y-%m-%d %H:%M:%S')
        time2 = (now - timedelta(days=1)).strftime('%Y-%m-%d %H:%M:%S')
        time3 = now.strftime('%Y-%m-%d %H:%M:%S')

        study_sessions_data = [
            (1, 1, 1, time1), # Group 1, Activity 1
            (2, 1, 2, time2), # Group 1, Activity 2
            (3, 2, 1, time3)  # Group 2, Activity 1
        ]
        cursor.executemany("INSERT INTO study_sessions (id, group_id, study_activity_id, created_at) VALUES (?, ?, ?, ?)", study_sessions_data)

        # Seed word_review_items for session 1
        cursor.execute("INSERT INTO word_review_items (study_session_id, word_id, correct, created_at) VALUES (?, ?, ?, ?)", (1, 1, 1, time1))
        cursor.execute("INSERT INTO word_review_items (study_session_id, word_id, correct, created_at) VALUES (?, ?, ?, ?)", (1, 2, 0, (datetime.fromisoformat(time1.replace('Z','')) + timedelta(minutes=5)).strftime('%Y-%m-%d %H:%M:%S')))
        
        # Seed word_review_items for session 3
        cursor.execute("INSERT INTO word_review_items (study_session_id, word_id, correct, created_at) VALUES (?, ?, ?, ?)", (3, 3, 1, time3))

        # Seed word_reviews (summary)
        # Word 1: 1 correct from session 1
        cursor.execute("INSERT INTO word_reviews (word_id, correct_count, wrong_count, last_reviewed) VALUES (?, ?, ?, ?)", (1, 1, 0, time1))
        # Word 2: 1 wrong from session 1
        cursor.execute("INSERT INTO word_reviews (word_id, correct_count, wrong_count, last_reviewed) VALUES (?, ?, ?, ?)", (2, 0, 1, (datetime.fromisoformat(time1.replace('Z','')) + timedelta(minutes=5)).strftime('%Y-%m-%d %H:%M:%S')))
        # Word 3 is not in word_reviews yet, will be created by POST /review test

        conn.commit()
    yield _app
    if os.path.exists(db_path):
        os.unlink(db_path)

@pytest.fixture
def client(app):
    return app.test_client()

# === Tests for POST /api/study-sessions (from previous session) ===
def test_create_study_session_success(client):
    response = client.post('/api/study-sessions',
                           data=json.dumps({'group_id': 1, 'activity_id': 1}),
                           content_type='application/json')
    assert response.status_code == 201
    data = json.loads(response.data)
    assert 'id' in data
    assert isinstance(data['id'], int)
    assert data['message'] == "Study session created successfully"

def test_create_study_session_missing_group_id(client):
    response = client.post('/api/study-sessions', data=json.dumps({'activity_id': 1}), content_type='application/json')
    assert response.status_code == 400
    assert json.loads(response.data)['error'] == "group_id and activity_id are required"

def test_create_study_session_missing_activity_id(client):
    response = client.post('/api/study-sessions', data=json.dumps({'group_id': 1}), content_type='application/json')
    assert response.status_code == 400
    assert json.loads(response.data)['error'] == "group_id and activity_id are required"

def test_create_study_session_empty_payload(client):
    response = client.post('/api/study-sessions', data=json.dumps({}), content_type='application/json')
    assert response.status_code == 400
    assert json.loads(response.data)['error'] == "group_id and activity_id are required"

# === Tests for GET /api/study-sessions ===
def test_get_study_sessions_success(client):
    response = client.get('/api/study-sessions')
    assert response.status_code == 200
    data = json.loads(response.data)
    assert 'items' in data
    assert 'total' in data
    assert 'page' in data
    assert 'per_page' in data
    assert 'total_pages' in data
    assert data['total'] == 3
    assert len(data['items']) == 3 # Default per_page is 10, all 3 fit
    # Default order is created_at DESC
    assert data['items'][0]['id'] == 3 # Most recent
    assert data['items'][1]['id'] == 2
    assert data['items'][2]['id'] == 1
    assert data['items'][0]['group_name'] == 'Test Group 2'
    assert data['items'][0]['activity_name'] == 'Flashcards'
    # Check review_items_count for session 1 (created above with 2 items)
    session1_data = next((item for item in data['items'] if item['id'] == 1), None)
    assert session1_data is not None
    assert session1_data['review_items_count'] == 2

def test_get_study_sessions_pagination(client):
    response = client.get('/api/study-sessions?page=1&per_page=2')
    assert response.status_code == 200
    data = json.loads(response.data)
    assert len(data['items']) == 2
    assert data['total'] == 3
    assert data['page'] == 1
    assert data['per_page'] == 2
    assert data['total_pages'] == 2
    assert data['items'][0]['id'] == 3

    response_page2 = client.get('/api/study-sessions?page=2&per_page=2')
    assert response_page2.status_code == 200
    data_page2 = json.loads(response_page2.data)
    assert len(data_page2['items']) == 1
    assert data_page2['items'][0]['id'] == 1

# === Tests for GET /api/study-sessions/<id> ===
def test_get_study_session_by_id_success(client):
    response = client.get('/api/study-sessions/1') # Session with 2 review items
    assert response.status_code == 200
    data = json.loads(response.data)
    assert 'session' in data
    assert 'words' in data
    assert data['session']['id'] == 1
    assert data['session']['group_name'] == 'Test Group 1'
    assert data['session']['activity_name'] == 'Flashcards'
    assert data['session']['review_items_count'] == 2
    assert len(data['words']) == 2
    word1_data = next((w for w in data['words'] if w['id'] == 1), None)
    assert word1_data['correct_count'] == 1
    assert word1_data['wrong_count'] == 0

def test_get_study_session_by_id_not_found(client):
    response = client.get('/api/study-sessions/999')
    assert response.status_code == 404
    assert json.loads(response.data)['error'] == "Study session not found"

# === Tests for POST /api/study-sessions/<study_session_id>/review ===
def test_review_word_in_session_success_new_review_summary(client):
    """Word 3 (ID 3) is in Group 2. Session 3 is for Group 2. Word 3 has no word_reviews entry yet."""
    response = client.post('/api/study-sessions/3/review', 
                           data=json.dumps({'word_id': 3, 'correct': True}),
                           content_type='application/json')
    assert response.status_code == 201
    data = json.loads(response.data)
    assert 'word_review_item_id' in data
    assert data['message'] == "Word review recorded successfully."
    # Verify word_reviews table got updated/created
    with client.application.app_context():
        cursor = client.application.db.cursor()
        cursor.execute("SELECT correct_count, wrong_count FROM word_reviews WHERE word_id = 3")
        review_summary = cursor.fetchone()
        assert review_summary is not None
        assert review_summary['correct_count'] == 1
        assert review_summary['wrong_count'] == 0

def test_review_word_in_session_success_update_review_summary(client):
    """Word 1 (ID 1) already has a review summary. Session 1 is for Group 1 (which has Word 1)."""
    # Initial state: Word 1 has correct_count=1, wrong_count=0
    response = client.post('/api/study-sessions/1/review', 
                           data=json.dumps({'word_id': 1, 'correct': False}), # New review is incorrect
                           content_type='application/json')
    assert response.status_code == 201
    with client.application.app_context():
        cursor = client.application.db.cursor()
        cursor.execute("SELECT correct_count, wrong_count FROM word_reviews WHERE word_id = 1")
        review_summary = cursor.fetchone()
        assert review_summary is not None
        assert review_summary['correct_count'] == 1 # Original correct count
        assert review_summary['wrong_count'] == 1 # Incremented wrong count

def test_review_word_in_session_invalid_session_id(client):
    response = client.post('/api/study-sessions/999/review', data=json.dumps({'word_id': 1, 'correct': True}), content_type='application/json')
    assert response.status_code == 404
    assert json.loads(response.data)['error'] == "Study session not found"

def test_review_word_in_session_invalid_word_id(client):
    response = client.post('/api/study-sessions/1/review', data=json.dumps({'word_id': 999, 'correct': True}), content_type='application/json')
    assert response.status_code == 404
    assert json.loads(response.data)['error'] == "Word not found"

def test_review_word_in_session_word_not_in_group(client):
    """Word 3 is in Group 2. Session 1 is for Group 1."""
    response = client.post('/api/study-sessions/1/review', 
                           data=json.dumps({'word_id': 3, 'correct': True}), 
                           content_type='application/json')
    assert response.status_code == 400
    assert "does not belong to group ID 1" in json.loads(response.data)['error']

def test_review_word_in_session_missing_word_id(client):
    response = client.post('/api/study-sessions/1/review', data=json.dumps({'correct': True}), content_type='application/json')
    assert response.status_code == 400
    assert json.loads(response.data)['error'] == "Valid 'word_id' (integer) is required"

def test_review_word_in_session_missing_correct_flag(client):
    response = client.post('/api/study-sessions/1/review', data=json.dumps({'word_id': 1}), content_type='application/json')
    assert response.status_code == 400
    assert json.loads(response.data)['error'] == "Valid 'correct' field (boolean) is required"

# === Tests for POST /api/study-sessions/reset ===
def test_reset_study_sessions_success(client):
    # First, ensure there's some data
    with client.application.app_context():
        cursor = client.application.db.cursor()
        # Use an explicit alias for COUNT(*) to avoid ambiguity with sqlite3.Row
        cursor.execute("SELECT COUNT(*) AS total_count FROM study_sessions")
        assert cursor.fetchone()['total_count'] > 0
        cursor.execute("SELECT COUNT(*) AS total_count FROM word_review_items")
        assert cursor.fetchone()['total_count'] > 0

    response = client.post('/api/study-sessions/reset')
    assert response.status_code == 200
    assert json.loads(response.data)['message'] == "Study history cleared successfully"

    # Verify tables are empty
    with client.application.app_context():
        cursor = client.application.db.cursor()
        cursor.execute("SELECT COUNT(*) AS total_count FROM study_sessions")
        assert cursor.fetchone()['total_count'] == 0
        cursor.execute("SELECT COUNT(*) AS total_count FROM word_review_items")
        assert cursor.fetchone()['total_count'] == 0

# Future tests to consider:
# - Test GET /api/study-sessions (listing, pagination)
# - Test GET /api/study-sessions/<id> (retrieving a specific session)
# - Test GET /api/study-sessions/<id> for a non-existent session (should be 404)
# - Test POST /api/study-sessions/reset
# - Test database interactions more deeply (e.g., verify data integrity after creation).
# - Test with invalid data types for group_id and activity_id. 