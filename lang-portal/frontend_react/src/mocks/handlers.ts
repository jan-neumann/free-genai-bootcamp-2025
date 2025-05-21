import { http, HttpResponse, delay } from 'msw';

const dashboardData = {
  lastStudySession: {
    id: 1,
    group_id: 1,
    created_at: '2025-05-20T09:03:23Z',
    study_activity_id: 1,
    group_name: 'basic_verbs',
    activity_name: 'Flashcards'
  },
  studyProgress: {
    total_words_studied: 0,
    total_available_words: 5
  },
  quickStats: {
    success_rate: 0,
    total_study_sessions: 3,
    total_active_groups: 1,
    study_streak_days: 3
  }
};

export const handlers = [
  http.get('/api/v1/dashboard/last-session', async () => {
    await delay(150);
    return HttpResponse.json(dashboardData.lastStudySession);
  }),
  
  http.get('/api/v1/dashboard/progress', async () => {
    await delay(100);
    return HttpResponse.json(dashboardData.studyProgress);
  }),
  
  http.get('/api/v1/dashboard/stats', async () => {
    await delay(100);
    return HttpResponse.json(dashboardData.quickStats);
  }),
];
