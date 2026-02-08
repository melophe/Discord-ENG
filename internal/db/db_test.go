package db

import (
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := New(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Verify tables exist by querying them
	_, err = db.conn.Exec("SELECT 1 FROM users LIMIT 1")
	if err != nil {
		t.Errorf("Users table not created: %v", err)
	}

	_, err = db.conn.Exec("SELECT 1 FROM questions LIMIT 1")
	if err != nil {
		t.Errorf("Questions table not created: %v", err)
	}

	_, err = db.conn.Exec("SELECT 1 FROM answers LIMIT 1")
	if err != nil {
		t.Errorf("Answers table not created: %v", err)
	}
}

func TestGetOrCreateUser(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := New(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// First call should create user
	user, err := db.GetOrCreateUser("12345")
	if err != nil {
		t.Fatalf("Failed to get/create user: %v", err)
	}

	if user.DiscordID != "12345" {
		t.Errorf("Expected discord_id '12345', got '%s'", user.DiscordID)
	}
	if user.Difficulty != "intermediate" {
		t.Errorf("Expected difficulty 'intermediate', got '%s'", user.Difficulty)
	}
	if user.Theme != "日常会話" {
		t.Errorf("Expected theme '日常会話', got '%s'", user.Theme)
	}
	if !user.ScheduleEnabled {
		t.Error("Expected schedule_enabled true, got false")
	}

	// Second call should return existing user
	user2, err := db.GetOrCreateUser("12345")
	if err != nil {
		t.Fatalf("Failed to get existing user: %v", err)
	}
	if user2.DiscordID != user.DiscordID {
		t.Error("Got different user on second call")
	}
}

func TestUpdateUserSettings(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := New(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create user
	_, err = db.GetOrCreateUser("12345")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Update settings
	err = db.UpdateUserSettings("12345", "advanced", "プログラミング")
	if err != nil {
		t.Fatalf("Failed to update settings: %v", err)
	}

	// Verify update
	user, err := db.GetOrCreateUser("12345")
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if user.Difficulty != "advanced" {
		t.Errorf("Expected difficulty 'advanced', got '%s'", user.Difficulty)
	}
	if user.Theme != "プログラミング" {
		t.Errorf("Expected theme 'プログラミング', got '%s'", user.Theme)
	}
}

func TestSaveQuestion(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := New(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Save question
	id, err := db.SaveQuestion("これはテストです", "intermediate", "テスト")
	if err != nil {
		t.Fatalf("Failed to save question: %v", err)
	}
	if id <= 0 {
		t.Errorf("Expected positive ID, got %d", id)
	}

	// Get question
	q, err := db.GetQuestion(id)
	if err != nil {
		t.Fatalf("Failed to get question: %v", err)
	}

	if q.Japanese != "これはテストです" {
		t.Errorf("Expected japanese 'これはテストです', got '%s'", q.Japanese)
	}
	if q.Difficulty != "intermediate" {
		t.Errorf("Expected difficulty 'intermediate', got '%s'", q.Difficulty)
	}
	if q.Theme != "テスト" {
		t.Errorf("Expected theme 'テスト', got '%s'", q.Theme)
	}
}

func TestSaveAnswer(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := New(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Save question first
	qID, _ := db.SaveQuestion("テスト問題", "beginner", "テスト")

	// Save answer
	err = db.SaveAnswer("12345", qID, "This is a test", "This is a test.", 95, "Great!")
	if err != nil {
		t.Fatalf("Failed to save answer: %v", err)
	}

	// Verify through stats
	stats, err := db.GetUserStats("12345")
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats.TotalAnswers != 1 {
		t.Errorf("Expected 1 answer, got %d", stats.TotalAnswers)
	}
	if stats.HighestScore != 95 {
		t.Errorf("Expected highest score 95, got %d", stats.HighestScore)
	}
}

func TestGetUserStats(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := New(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Empty stats
	stats, err := db.GetUserStats("12345")
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats.TotalAnswers != 0 {
		t.Errorf("Expected 0 answers, got %d", stats.TotalAnswers)
	}

	// Add some answers
	qID, _ := db.SaveQuestion("テスト1", "beginner", "テスト")
	db.SaveAnswer("12345", qID, "test1", "test1", 80, "Good")
	db.SaveAnswer("12345", qID, "test2", "test2", 90, "Great")
	db.SaveAnswer("12345", qID, "test3", "test3", 100, "Perfect")

	stats, err = db.GetUserStats("12345")
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats.TotalAnswers != 3 {
		t.Errorf("Expected 3 answers, got %d", stats.TotalAnswers)
	}
	if stats.HighestScore != 100 {
		t.Errorf("Expected highest score 100, got %d", stats.HighestScore)
	}
	expectedAvg := 90.0
	if stats.AverageScore != expectedAvg {
		t.Errorf("Expected average %.1f, got %.1f", expectedAvg, stats.AverageScore)
	}
}
