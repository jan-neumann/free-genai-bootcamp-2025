package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWord_Validate(t *testing.T) {
	tests := []struct {
		name    string
		word    Word
		wantErr bool
	}{
		{
			name: "valid word",
			word: Word{
				Japanese: "テスト",
				Romaji:   "tesuto",
				English:  "test",
				Parts:    StringSlice{"noun"},
			},
			wantErr: false,
		},
		{
			name: "empty japanese",
			word: Word{
				Japanese: "",
				Romaji:   "tesuto",
				English:  "test",
				Parts:    StringSlice{"noun"},
			},
			wantErr: true,
		},
		{
			name: "empty romaji",
			word: Word{
				Japanese: "テスト",
				Romaji:   "",
				English:  "test",
				Parts:    StringSlice{"noun"},
			},
			wantErr: true,
		},
		{
			name: "empty english",
			word: Word{
				Japanese: "テスト",
				Romaji:   "tesuto",
				English:  "",
				Parts:    StringSlice{"noun"},
			},
			wantErr: true,
		},
		{
			name: "empty parts",
			word: Word{
				Japanese: "テスト",
				Romaji:   "tesuto",
				English:  "test",
				Parts:    StringSlice{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.word.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWord_GetStudyStats(t *testing.T) {
	tests := []struct {
		name        string
		word        Word
		wantCorrect int
		wantWrong   int
	}{
		{
			name: "no reviews",
			word: Word{
				Reviews: []WordReview{},
			},
			wantCorrect: 0,
			wantWrong:   0,
		},
		{
			name: "all correct",
			word: Word{
				Reviews: []WordReview{
					{Correct: true},
					{Correct: true},
					{Correct: true},
				},
			},
			wantCorrect: 3,
			wantWrong:   0,
		},
		{
			name: "mixed results",
			word: Word{
				Reviews: []WordReview{
					{Correct: true},
					{Correct: false},
					{Correct: true},
					{Correct: false},
				},
			},
			wantCorrect: 2,
			wantWrong:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCorrect, gotWrong := tt.word.GetStudyStats()
			assert.Equal(t, tt.wantCorrect, gotCorrect)
			assert.Equal(t, tt.wantWrong, gotWrong)
		})
	}
}

func TestWord_GetSuccessRate(t *testing.T) {
	tests := []struct {
		name string
		word Word
		want float64
	}{
		{
			name: "no reviews",
			word: Word{
				Reviews: []WordReview{},
			},
			want: 0,
		},
		{
			name: "all correct",
			word: Word{
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
			word: Word{
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
			word: Word{
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
			got := tt.word.GetSuccessRate()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStringSlice_Value_Scan(t *testing.T) {
	tests := []struct {
		name    string
		input   StringSlice
		want    StringSlice
		wantErr bool
	}{
		{
			name:    "empty slice",
			input:   StringSlice{},
			want:    StringSlice{},
			wantErr: false,
		},
		{
			name:    "single item",
			input:   StringSlice{"noun"},
			want:    StringSlice{"noun"},
			wantErr: false,
		},
		{
			name:    "multiple items",
			input:   StringSlice{"noun", "verb", "adjective"},
			want:    StringSlice{"noun", "verb", "adjective"},
			wantErr: false,
		},
		{
			name:    "nil input",
			input:   nil,
			want:    StringSlice{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Value()
			value, err := tt.input.Value()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Test Scan()
			var got StringSlice
			err = got.Scan(value)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
