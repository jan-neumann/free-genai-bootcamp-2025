import pytest
import json
import os
from app import create_app # Assuming pytest is run from backend-flask directory

@pytest.fixture
def app():
    """Create and configure a new app instance for each test."""
    base_dir = os.path.dirname(os.path.abspath(__file__)) # tests directory
    project_root = os.path.dirname(base_dir) # backend-flask directory
    # Use a unique name for the test database for groups tests
    db_path = os.path.join(project_root, 'test_groups_module.db')
    
    _app = create_app({
        'TESTING': True,
        'DATABASE': db_path
    })

    with _app.app_context():
        db_instance = _app.db
        conn = db_instance.get()
        cursor = conn.cursor()

        sql_setup_dir = os.path.join(project_root, 'sql', 'setup')

        # Tables needed for group tests, including those for related endpoints
        table_files = [
            'create_table_groups.sql',
            'create_table_words.sql',
            'create_table_word_groups.sql',
            'create_table_word_reviews.sql',
            'create_table_study_activities.sql', # For /groups/<id>/study_sessions
            'create_table_study_sessions.sql',   # For /groups/<id>/study_sessions
            'create_table_word_review_items.sql' # For /groups/<id>/study_sessions
        ]
        for table_file in table_files:
            with open(os.path.join(sql_setup_dir, table_file), 'r') as f:
                conn.executescript(f.read())
        
        # Seed data for groups
        groups_data = [
            (1, 'Adjectives', 2),
            (2, 'Verbs', 1),
            (3, 'Nouns', 0),
            (4, 'Zoo Phonics', 0) # For name sorting test
        ]
        cursor.executemany("INSERT INTO groups (id, name, words_count) VALUES (?, ?, ?)", groups_data)
        
        # Seed data for words (needed for /groups/<id>/words and /raw later)
        words_data = [
            (1, '美しい', 'utsukushii', 'beautiful', '{"type": "i-adjective"}'),
            (2, '新しい', 'atarashii', 'new', '{"type": "i-adjective"}'),
            (3, '走る', 'hashiru', 'run', '{"type": "verb"}'),
            (4, '猫', 'neko', 'cat', '{"type": "noun"}')
        ]
        cursor.executemany("INSERT INTO words (id, kanji, romaji, english, parts) VALUES (?, ?, ?, ?, ?)", words_data)

        # Seed data for word_groups
        word_groups_data = [
            (1, 1), # Word '美しい' in Group 'Adjectives'
            (2, 1), # Word '新しい' in Group 'Adjectives'
            (3, 2)  # Word '走る' in Group 'Verbs'
        ]
        cursor.executemany("INSERT INTO word_groups (word_id, group_id) VALUES (?, ?)", word_groups_data)

        # Seed data for word_reviews (for /groups/<id>/words which joins on it)
        word_reviews_data = [
            (1, 1, 5, 1, '2023-01-01 10:00:00'), # review for word '美しい'
            (2, 3, 10, 2, '2023-01-02 11:00:00') # review for word '走る'
        ]
        cursor.executemany("INSERT INTO word_reviews (id, word_id, correct_count, wrong_count, last_reviewed) VALUES (?, ?, ?, ?, ?)", word_reviews_data)
        
        # Seed data for study_activities (for /groups/<id>/study_sessions)
        cursor.execute("INSERT INTO study_activities (id, name, url, preview_url) VALUES (?, ?, ?, ?)", 
                       (1, 'Flashcards', 'http://example.com/flashcards', 'http://example.com/flash_preview.jpg'))

        # Seed data for study_sessions (for /groups/<id>/study_sessions)
        study_sessions_data = [
            (1, 1, 1, '2023-01-01 10:00:00'), # Session for Group 'Adjectives', Activity 'Flashcards'
            (2, 1, 1, '2023-01-05 12:00:00'), # Another session for Group 'Adjectives'
            (3, 2, 1, '2023-01-02 11:00:00')  # Session for Group 'Verbs'
        ]
        cursor.executemany("INSERT INTO study_sessions (id, group_id, study_activity_id, created_at) VALUES (?, ?, ?, ?)", study_sessions_data)

        # Seed data for word_review_items (for /groups/<id>/study_sessions for review_count & last_activity_time)
        word_review_items_data = [
            (1, 1, 1, 1, '2023-01-01 10:05:00'), # Session 1, Word 1, correct
            (2, 1, 2, 0, '2023-01-01 10:06:00')  # Session 1, Word 2, incorrect
        ]
        cursor.executemany("INSERT INTO word_review_items (id, study_session_id, word_id, correct, created_at) VALUES (?, ?, ?, ?, ?)", word_review_items_data)

        conn.commit()

    yield _app

    if os.path.exists(db_path):
        os.unlink(db_path)

@pytest.fixture
def client(app):
    return app.test_client()

# Tests for GET /groups
def test_get_groups_success(client):
    """Test successful retrieval of groups with default pagination and sorting."""
    response = client.get('/groups')
    assert response.status_code == 200
    data = json.loads(response.data)
    assert 'groups' in data
    assert 'total_pages' in data
    assert 'current_page' in data
    assert len(data['groups']) <= 10 # Default groups_per_page
    # Default sort is by name asc: Adjectives, Nouns, Verbs, Zoo Phonics
    assert data['groups'][0]['group_name'] == 'Adjectives'
    assert data['groups'][1]['group_name'] == 'Nouns'
    assert data['groups'][2]['group_name'] == 'Verbs'
    assert data['groups'][3]['group_name'] == 'Zoo Phonics'
    assert data['groups'][0]['word_count'] == 2

def test_get_groups_pagination(client):
    """Test pagination for groups."""
    response = client.get('/groups?page=1&groups_per_page=2') # groups_per_page is actually hardcoded to 10 in route
    data = json.loads(response.data)
    # The route uses a hardcoded groups_per_page=10, so this test will reflect that.
    # If groups_per_page was configurable via query param, this test would be different.
    assert len(data['groups']) == 4 # All 4 groups fit in the default 10 per page
    assert data['current_page'] == 1

def test_get_groups_sort_by_word_count_desc(client):
    """Test sorting groups by word_count descending."""
    response = client.get('/groups?sort_by=words_count&order=desc')
    assert response.status_code == 200
    data = json.loads(response.data)
    assert data['groups'][0]['group_name'] == 'Adjectives' # 2 words
    assert data['groups'][0]['word_count'] == 2
    assert data['groups'][1]['group_name'] == 'Verbs'      # 1 word
    assert data['groups'][1]['word_count'] == 1
    assert data['groups'][2]['word_count'] == 0 # Nouns or Zoo Phonics
    assert data['groups'][3]['word_count'] == 0 # Nouns or Zoo Phonics

def test_get_groups_sort_by_name_desc(client):
    """Test sorting groups by name descending."""
    response = client.get('/groups?sort_by=name&order=desc')
    assert response.status_code == 200
    data = json.loads(response.data)
    assert data['groups'][0]['group_name'] == 'Zoo Phonics'
    assert data['groups'][1]['group_name'] == 'Verbs'
    assert data['groups'][2]['group_name'] == 'Nouns'
    assert data['groups'][3]['group_name'] == 'Adjectives'

# Tests for GET /groups/<id>
def test_get_group_by_id_success(client):
    """Test successful retrieval of a single group by its ID."""
    response = client.get('/groups/1') # Get group 'Adjectives'
    assert response.status_code == 200
    data = json.loads(response.data)
    assert data['id'] == 1
    assert data['group_name'] == 'Adjectives'
    assert data['word_count'] == 2

def test_get_group_by_id_not_found(client):
    """Test retrieval of a non-existent group ID."""
    response = client.get('/groups/999')
    assert response.status_code == 404
    data = json.loads(response.data)
    assert data['error'] == "Group not found"

# Placeholder for future tests - will add step-by-step if requested
# def test_get_group_words_success(client):
#     pass

# def test_get_group_words_raw_success(client):
#     pass

# def test_get_group_study_sessions_success(client):
#     pass 