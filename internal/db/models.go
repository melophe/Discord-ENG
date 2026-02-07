package db

import "time"

// User represents a Discord user's settings
type User struct {
	DiscordID       string
	Difficulty      string
	Theme           string
	ScheduleEnabled bool
	CreatedAt       time.Time
}

// Question represents a quiz question
type Question struct {
	ID         int64
	Japanese   string
	Difficulty string
	Theme      string
	CreatedAt  time.Time
}

// Answer represents a user's answer to a question
type Answer struct {
	ID          int64
	DiscordID   string
	QuestionID  int64
	UserAnswer  string
	ModelAnswer string
	Score       int
	Feedback    string
	AnsweredAt  time.Time
}

// GetOrCreateUser gets a user by Discord ID, creating if not exists
func (db *DB) GetOrCreateUser(discordID string) (*User, error) {
	user := &User{DiscordID: discordID}

	row := db.conn.QueryRow(
		"SELECT difficulty, theme, schedule_enabled, created_at FROM users WHERE discord_id = ?",
		discordID,
	)

	var scheduleEnabled int
	err := row.Scan(&user.Difficulty, &user.Theme, &scheduleEnabled, &user.CreatedAt)
	if err != nil {
		// User doesn't exist, create new one
		_, err = db.conn.Exec(
			"INSERT INTO users (discord_id) VALUES (?)",
			discordID,
		)
		if err != nil {
			return nil, err
		}
		user.Difficulty = "intermediate"
		user.Theme = "日常会話"
		user.ScheduleEnabled = true
		user.CreatedAt = time.Now()
	} else {
		user.ScheduleEnabled = scheduleEnabled == 1
	}

	return user, nil
}

// UpdateUserSettings updates a user's difficulty and theme
func (db *DB) UpdateUserSettings(discordID, difficulty, theme string) error {
	_, err := db.conn.Exec(
		"UPDATE users SET difficulty = ?, theme = ? WHERE discord_id = ?",
		difficulty, theme, discordID,
	)
	return err
}

// UpdateUserSchedule updates a user's schedule setting
func (db *DB) UpdateUserSchedule(discordID string, enabled bool) error {
	enabledInt := 0
	if enabled {
		enabledInt = 1
	}
	_, err := db.conn.Exec(
		"UPDATE users SET schedule_enabled = ? WHERE discord_id = ?",
		enabledInt, discordID,
	)
	return err
}

// SaveQuestion saves a new question and returns its ID
func (db *DB) SaveQuestion(japanese, difficulty, theme string) (int64, error) {
	result, err := db.conn.Exec(
		"INSERT INTO questions (japanese, difficulty, theme) VALUES (?, ?, ?)",
		japanese, difficulty, theme,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetQuestion gets a question by ID
func (db *DB) GetQuestion(id int64) (*Question, error) {
	q := &Question{ID: id}
	row := db.conn.QueryRow(
		"SELECT japanese, difficulty, theme, created_at FROM questions WHERE id = ?",
		id,
	)
	err := row.Scan(&q.Japanese, &q.Difficulty, &q.Theme, &q.CreatedAt)
	if err != nil {
		return nil, err
	}
	return q, nil
}

// SaveAnswer saves a user's answer
func (db *DB) SaveAnswer(discordID string, questionID int64, userAnswer, modelAnswer string, score int, feedback string) error {
	_, err := db.conn.Exec(
		"INSERT INTO answers (discord_id, question_id, user_answer, model_answer, score, feedback) VALUES (?, ?, ?, ?, ?, ?)",
		discordID, questionID, userAnswer, modelAnswer, score, feedback,
	)
	return err
}

// UserStats represents a user's learning statistics
type UserStats struct {
	TotalAnswers   int
	AverageScore   float64
	HighestScore   int
	AnswersToday   int
	CurrentStreak  int
}

// GetUserStats gets statistics for a user
func (db *DB) GetUserStats(discordID string) (*UserStats, error) {
	stats := &UserStats{}

	// Total answers and average score
	row := db.conn.QueryRow(`
		SELECT COUNT(*), COALESCE(AVG(score), 0), COALESCE(MAX(score), 0)
		FROM answers WHERE discord_id = ?
	`, discordID)
	err := row.Scan(&stats.TotalAnswers, &stats.AverageScore, &stats.HighestScore)
	if err != nil {
		return nil, err
	}

	// Answers today
	row = db.conn.QueryRow(`
		SELECT COUNT(*) FROM answers
		WHERE discord_id = ? AND DATE(answered_at) = DATE('now')
	`, discordID)
	err = row.Scan(&stats.AnswersToday)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
