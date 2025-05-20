package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStudyActivity_Validate(t *testing.T) {
	tests := []struct {
		name     string
		activity StudyActivity
		wantErr  bool
	}{
		{
			name: "valid activity",
			activity: StudyActivity{
				Name:         "Test Activity",
				Description:  "Test Description",
				ThumbnailURL: "https://example.com/image.jpg",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			activity: StudyActivity{
				Name:         "",
				Description:  "Test Description",
				ThumbnailURL: "https://example.com/image.jpg",
			},
			wantErr: true,
		},
		{
			name: "empty description",
			activity: StudyActivity{
				Name:         "Test Activity",
				Description:  "",
				ThumbnailURL: "https://example.com/image.jpg",
			},
			wantErr: true,
		},
		{
			name: "invalid thumbnail URL",
			activity: StudyActivity{
				Name:         "Test Activity",
				Description:  "Test Description",
				ThumbnailURL: "not-a-url",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.activity.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStudySession_Validate(t *testing.T) {
	validGroup := Group{Name: "Test Group"}
	validActivity := StudyActivity{Name: "Test Activity", Description: "desc", ThumbnailURL: "https://example.com/img.jpg"}
	tests := []struct {
		name    string
		session StudySession
		wantErr bool
	}{
		{
			name: "valid session",
			session: StudySession{
				GroupID:         1,
				StudyActivityID: 1,
				Group:           validGroup,
				Activity:        validActivity,
			},
			wantErr: false,
		},
		{
			name: "zero group ID",
			session: StudySession{
				GroupID:         0,
				StudyActivityID: 1,
				Group:           validGroup,
				Activity:        validActivity,
			},
			wantErr: true,
		},
		{
			name: "zero activity ID",
			session: StudySession{
				GroupID:         1,
				StudyActivityID: 0,
				Group:           validGroup,
				Activity:        validActivity,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.session.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStudySession_GetStudyStats(t *testing.T) {
	tests := []struct {
		name        string
		session     StudySession
		wantTotal   int
		wantCorrect int
	}{
		{
			name: "no reviews",
			session: StudySession{
				Reviews: []WordReview{},
			},
			wantTotal:   0,
			wantCorrect: 0,
		},
		{
			name: "all correct",
			session: StudySession{
				Reviews: []WordReview{
					{Correct: true},
					{Correct: true},
					{Correct: true},
				},
			},
			wantTotal:   3,
			wantCorrect: 3,
		},
		{
			name: "mixed results",
			session: StudySession{
				Reviews: []WordReview{
					{Correct: true},
					{Correct: false},
					{Correct: true},
					{Correct: false},
				},
			},
			wantTotal:   4,
			wantCorrect: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTotal, gotCorrect := tt.session.GetStudyStats()
			assert.Equal(t, tt.wantTotal, gotTotal)
			assert.Equal(t, tt.wantCorrect, gotCorrect)
		})
	}
}

func TestStudySession_GetSuccessRate(t *testing.T) {
	tests := []struct {
		name    string
		session StudySession
		want    float64
	}{
		{
			name: "no reviews",
			session: StudySession{
				Reviews: []WordReview{},
			},
			want: 0,
		},
		{
			name: "all correct",
			session: StudySession{
				Reviews: []WordReview{
					{Correct: true},
					{Correct: true},
					{Correct: true},
				},
			},
			want: 100,
		},
		{
			name: "half correct",
			session: StudySession{
				Reviews: []WordReview{
					{Correct: true},
					{Correct: false},
					{Correct: true},
					{Correct: false},
				},
			},
			want: 50,
		},
		{
			name: "all wrong",
			session: StudySession{
				Reviews: []WordReview{
					{Correct: false},
					{Correct: false},
					{Correct: false},
				},
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.session.GetSuccessRate()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWordReview_Validate(t *testing.T) {
	validWord := Word{Japanese: "テスト", Romaji: "tesuto", English: "test", Parts: StringSlice{"noun"}}
	validSession := StudySession{GroupID: 1, StudyActivityID: 1, Group: Group{Name: "Test Group"}, Activity: StudyActivity{Name: "Test Activity", Description: "desc", ThumbnailURL: "https://example.com/img.jpg"}}
	tests := []struct {
		name    string
		review  WordReview
		wantErr bool
	}{
		{
			name: "valid review",
			review: WordReview{
				WordID:         1,
				StudySessionID: 1,
				Correct:        true,
				Word:           validWord,
				StudySession:   validSession,
			},
			wantErr: false,
		},
		{
			name: "zero word ID",
			review: WordReview{
				WordID:         0,
				StudySessionID: 1,
				Correct:        true,
				Word:           validWord,
				StudySession:   validSession,
			},
			wantErr: true,
		},
		{
			name: "zero session ID",
			review: WordReview{
				WordID:         1,
				StudySessionID: 0,
				Correct:        true,
				Word:           validWord,
				StudySession:   validSession,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.review.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
