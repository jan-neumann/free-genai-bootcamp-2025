from flask import request, jsonify, g
from flask_cors import cross_origin
from datetime import datetime, UTC
import math

def load(app):

  @app.route('/api/study-sessions', methods=['POST'])
  @cross_origin()
  def create_study_session():
    try:
      data = request.get_json()
      group_id = data.get('group_id')
      activity_id = data.get('activity_id')

      if not group_id or not activity_id:
        return jsonify({"error": "group_id and activity_id are required"}), 400

      cursor = app.db.cursor()
      created_at = datetime.now(UTC).strftime('%Y-%m-%d %H:%M:%S')
      
      cursor.execute('''
        INSERT INTO study_sessions (group_id, study_activity_id, created_at)
        VALUES (?, ?, ?)
      ''', (group_id, activity_id, created_at))
      
      session_id = cursor.lastrowid
      app.db.commit()
      
      return jsonify({"id": session_id, "message": "Study session created successfully"}), 201
      
    except Exception as e:
      # Rollback in case of error
      if hasattr(app.db, 'rollback'):
          app.db.rollback()
      return jsonify({"error": str(e)}), 500

  @app.route('/api/study-sessions', methods=['GET'])
  @cross_origin()
  def get_study_sessions():
    try:
      cursor = app.db.cursor()
      
      # Get pagination parameters
      page = request.args.get('page', 1, type=int)
      per_page = request.args.get('per_page', 10, type=int)
      offset = (page - 1) * per_page

      # Get total count
      cursor.execute('''
        SELECT COUNT(*) as count 
        FROM study_sessions ss
        JOIN groups g ON g.id = ss.group_id
        JOIN study_activities sa ON sa.id = ss.study_activity_id
      ''')
      total_count = cursor.fetchone()['count']

      # Get paginated sessions
      cursor.execute('''
        SELECT 
          ss.id,
          ss.group_id,
          g.name as group_name,
          sa.id as activity_id,
          sa.name as activity_name,
          ss.created_at,
          COUNT(wri.id) as review_items_count
        FROM study_sessions ss
        JOIN groups g ON g.id = ss.group_id
        JOIN study_activities sa ON sa.id = ss.study_activity_id
        LEFT JOIN word_review_items wri ON wri.study_session_id = ss.id
        GROUP BY ss.id
        ORDER BY ss.created_at DESC
        LIMIT ? OFFSET ?
      ''', (per_page, offset))
      sessions = cursor.fetchall()

      return jsonify({
        'items': [{
          'id': session['id'],
          'group_id': session['group_id'],
          'group_name': session['group_name'],
          'activity_id': session['activity_id'],
          'activity_name': session['activity_name'],
          'start_time': session['created_at'],
          'end_time': session['created_at'],  # For now, just use the same time since we don't track end time
          'review_items_count': session['review_items_count']
        } for session in sessions],
        'total': total_count,
        'page': page,
        'per_page': per_page,
        'total_pages': math.ceil(total_count / per_page)
      })
    except Exception as e:
      return jsonify({"error": str(e)}), 500

  @app.route('/api/study-sessions/<id>', methods=['GET'])
  @cross_origin()
  def get_study_session(id):
    try:
      cursor = app.db.cursor()
      
      # Get session details
      cursor.execute('''
        SELECT 
          ss.id,
          ss.group_id,
          g.name as group_name,
          sa.id as activity_id,
          sa.name as activity_name,
          ss.created_at,
          COUNT(wri.id) as review_items_count
        FROM study_sessions ss
        JOIN groups g ON g.id = ss.group_id
        JOIN study_activities sa ON sa.id = ss.study_activity_id
        LEFT JOIN word_review_items wri ON wri.study_session_id = ss.id
        WHERE ss.id = ?
        GROUP BY ss.id
      ''', (id,))
      
      session = cursor.fetchone()
      if not session:
        return jsonify({"error": "Study session not found"}), 404

      # Get pagination parameters
      page = request.args.get('page', 1, type=int)
      per_page = request.args.get('per_page', 10, type=int)
      offset = (page - 1) * per_page

      # Get the words reviewed in this session with their review status
      cursor.execute('''
        SELECT 
          w.*,
          COALESCE(SUM(CASE WHEN wri.correct = 1 THEN 1 ELSE 0 END), 0) as session_correct_count,
          COALESCE(SUM(CASE WHEN wri.correct = 0 THEN 1 ELSE 0 END), 0) as session_wrong_count
        FROM words w
        JOIN word_review_items wri ON wri.word_id = w.id
        WHERE wri.study_session_id = ?
        GROUP BY w.id
        ORDER BY w.kanji
        LIMIT ? OFFSET ?
      ''', (id, per_page, offset))
      
      words = cursor.fetchall()

      # Get total count of words
      cursor.execute('''
        SELECT COUNT(DISTINCT w.id) as count
        FROM words w
        JOIN word_review_items wri ON wri.word_id = w.id
        WHERE wri.study_session_id = ?
      ''', (id,))
      
      total_count = cursor.fetchone()['count']

      return jsonify({
        'session': {
          'id': session['id'],
          'group_id': session['group_id'],
          'group_name': session['group_name'],
          'activity_id': session['activity_id'],
          'activity_name': session['activity_name'],
          'start_time': session['created_at'],
          'end_time': session['created_at'],  # For now, just use the same time
          'review_items_count': session['review_items_count']
        },
        'words': [{
          'id': word['id'],
          'kanji': word['kanji'],
          'romaji': word['romaji'],
          'english': word['english'],
          'correct_count': word['session_correct_count'],
          'wrong_count': word['session_wrong_count']
        } for word in words],
        'total': total_count,
        'page': page,
        'per_page': per_page,
        'total_pages': math.ceil(total_count / per_page)
      })
    except Exception as e:
      return jsonify({"error": str(e)}), 500

  # todo POST /study_sessions/:id/review
  @app.route('/api/study-sessions/<int:study_session_id>/review', methods=['POST'])
  @cross_origin()
  def review_word_in_session(study_session_id):
    try:
      data = request.get_json()
      word_id = data.get('word_id')
      is_correct_input = data.get('correct')

      if word_id is None or not isinstance(word_id, int):
        return jsonify({"error": "Valid 'word_id' (integer) is required"}), 400
      if is_correct_input is None or not isinstance(is_correct_input, bool):
        return jsonify({"error": "Valid 'correct' field (boolean) is required"}), 400

      is_correct_int = 1 if is_correct_input else 0
      review_time_utc_str = datetime.now(UTC).strftime('%Y-%m-%d %H:%M:%S')

      cursor = app.db.cursor()

      # 1. Validate study_session_id and get its group_id
      cursor.execute('SELECT group_id FROM study_sessions WHERE id = ?', (study_session_id,))
      session_record = cursor.fetchone()
      if not session_record:
        return jsonify({"error": "Study session not found"}), 404
      session_group_id = session_record['group_id']

      # 2. Validate word_id
      cursor.execute('SELECT id FROM words WHERE id = ?', (word_id,))
      word_record = cursor.fetchone()
      if not word_record:
        return jsonify({"error": "Word not found"}), 404
      
      # 3. Validate if the word belongs to the session's group (optional but good practice)
      cursor.execute('SELECT 1 FROM word_groups WHERE word_id = ? AND group_id = ?', (word_id, session_group_id))
      word_in_group = cursor.fetchone()
      if not word_in_group:
          return jsonify({"error": f"Word ID {word_id} does not belong to group ID {session_group_id} of this study session."}), 400

      # 4. Insert into word_review_items
      cursor.execute('''
        INSERT INTO word_review_items (study_session_id, word_id, correct, created_at)
        VALUES (?, ?, ?, ?)
      ''', (study_session_id, word_id, is_correct_int, review_time_utc_str))
      new_review_item_id = cursor.lastrowid

      # 5. Update word_reviews (summary table)
      cursor.execute('SELECT id, correct_count, wrong_count FROM word_reviews WHERE word_id = ?', (word_id,))
      word_review_summary = cursor.fetchone()

      if word_review_summary:
        new_correct_count = word_review_summary['correct_count'] + (1 if is_correct_int == 1 else 0)
        new_wrong_count = word_review_summary['wrong_count'] + (1 if is_correct_int == 0 else 0)
        cursor.execute('''
          UPDATE word_reviews 
          SET correct_count = ?, wrong_count = ?, last_reviewed = ? 
          WHERE id = ?
        ''', (new_correct_count, new_wrong_count, review_time_utc_str, word_review_summary['id']))
      else:
        new_correct_count = 1 if is_correct_int == 1 else 0
        new_wrong_count = 1 if is_correct_int == 0 else 0
        cursor.execute('''
          INSERT INTO word_reviews (word_id, correct_count, wrong_count, last_reviewed)
          VALUES (?, ?, ?, ?)
        ''', (word_id, new_correct_count, new_wrong_count, review_time_utc_str))

      app.db.commit()
      return jsonify({
        "word_review_item_id": new_review_item_id, 
        "message": "Word review recorded successfully."
      }), 201

    except Exception as e:
      if hasattr(app.db, 'rollback'):
        app.db.rollback()
      # For more detailed debugging during development:
      # import traceback
      # print(traceback.format_exc())
      return jsonify({"error": str(e)}), 500

  @app.route('/api/study-sessions/reset', methods=['POST'])
  @cross_origin()
  def reset_study_sessions():
    try:
      cursor = app.db.cursor()
      
      # First delete all word review items since they have foreign key constraints
      cursor.execute('DELETE FROM word_review_items')
      
      # Then delete all study sessions
      cursor.execute('DELETE FROM study_sessions')
      
      app.db.commit()
      
      return jsonify({"message": "Study history cleared successfully"}), 200
    except Exception as e:
      return jsonify({"error": str(e)}), 500