-- Drop indexes
DROP INDEX IF EXISTS idx_word_groups_word_id;
DROP INDEX IF EXISTS idx_word_groups_group_id;

-- Drop tables in reverse order
DROP TABLE IF EXISTS word_groups;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS words; 