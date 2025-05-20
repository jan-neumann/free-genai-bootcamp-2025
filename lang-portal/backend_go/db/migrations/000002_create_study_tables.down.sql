-- Drop indexes first
DROP INDEX IF EXISTS idx_word_review_items_study_session_id;
DROP INDEX IF EXISTS idx_word_review_items_word_id;
DROP INDEX IF EXISTS idx_study_sessions_study_activity_id;
DROP INDEX IF EXISTS idx_study_sessions_group_id;

-- Drop tables in reverse order of creation
DROP TABLE IF EXISTS word_review_items;
DROP TABLE IF EXISTS study_sessions;
DROP TABLE IF EXISTS study_activities; 