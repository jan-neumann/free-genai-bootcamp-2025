package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroup_Validate(t *testing.T) {
	tests := []struct {
		name    string
		group   Group
		wantErr bool
	}{
		{
			name: "valid group",
			group: Group{
				Name: "Test Group",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			group: Group{
				Name: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.group.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGroup_GetWordCount(t *testing.T) {
	tests := []struct {
		name  string
		group Group
		want  int
	}{
		{
			name: "no words",
			group: Group{
				Words: []Word{},
			},
			want: 0,
		},
		{
			name: "single word",
			group: Group{
				Words: []Word{
					{Japanese: "テスト"},
				},
			},
			want: 1,
		},
		{
			name: "multiple words",
			group: Group{
				Words: []Word{
					{Japanese: "テスト"},
					{Japanese: "こんにちは"},
					{Japanese: "さようなら"},
				},
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.group.GetWordCount()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGroup_GetStudyStats(t *testing.T) {
	tests := []struct {
		name         string
		group        Group
		wantSessions int
		wantReviews  int
		wantCorrect  int
	}{
		{
			name: "no sessions",
			group: Group{
				Sessions: []StudySession{},
			},
			wantSessions: 0,
			wantReviews:  0,
			wantCorrect:  0,
		},
		{
			name: "sessions with no reviews",
			group: Group{
				Sessions: []StudySession{
					{},
					{},
				},
			},
			wantSessions: 2,
			wantReviews:  0,
			wantCorrect:  0,
		},
		{
			name: "sessions with reviews",
			group: Group{
				Sessions: []StudySession{
					{
						Reviews: []WordReview{
							{Correct: true},
							{Correct: false},
						},
					},
					{
						Reviews: []WordReview{
							{Correct: true},
							{Correct: true},
						},
					},
				},
			},
			wantSessions: 2,
			wantReviews:  4,
			wantCorrect:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSessions, gotReviews, gotCorrect := tt.group.GetStudyStats()
			assert.Equal(t, tt.wantSessions, gotSessions)
			assert.Equal(t, tt.wantReviews, gotReviews)
			assert.Equal(t, tt.wantCorrect, gotCorrect)
		})
	}
}

func TestGroup_GetSuccessRate(t *testing.T) {
	tests := []struct {
		name  string
		group Group
		want  float64
	}{
		{
			name: "no reviews",
			group: Group{
				Sessions: []StudySession{},
			},
			want: 0,
		},
		{
			name: "all correct",
			group: Group{
				Sessions: []StudySession{
					{
						Reviews: []WordReview{
							{Correct: true},
							{Correct: true},
						},
					},
				},
			},
			want: 100,
		},
		{
			name: "half correct",
			group: Group{
				Sessions: []StudySession{
					{
						Reviews: []WordReview{
							{Correct: true},
							{Correct: false},
							{Correct: true},
							{Correct: false},
						},
					},
				},
			},
			want: 50,
		},
		{
			name: "all wrong",
			group: Group{
				Sessions: []StudySession{
					{
						Reviews: []WordReview{
							{Correct: false},
							{Correct: false},
						},
					},
				},
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.group.GetSuccessRate()
			assert.Equal(t, tt.want, got)
		})
	}
}
