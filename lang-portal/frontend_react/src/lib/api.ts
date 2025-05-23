const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8081';

async function fetchApi<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`;
  const response = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({}));
    throw new Error(error.message || 'Something went wrong');
  }

  return response.json();
}

export interface StudyActivityGroup {
  id: number;
  name: string;
}

export interface StudyActivity {
  id: number;
  name: string;
  description: string;
  thumbnail_url?: string;
  total_sessions: number;
  last_session_date?: string;
  available_groups?: StudyActivityGroup[];
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

export interface StudySession {
  id: number;
  activity_id: number;
  activity_name: string;
  group_id: number;
  group_name: string;
  start_time: string;
  end_time?: string;
  score?: number;
  total_questions: number;
  correct_answers: number;
}

export interface StudySessionWord {
  id: number;
  word: string;
  translation: string;
  example: string;
  reviewed: boolean;
  correct: boolean;
}

export const studyApi = {
  getActivities: async (): Promise<PaginatedResponse<StudyActivity>> => {
    const response = await fetchApi<{ 
      items: Array<{
        id: number;
        name: string;
        description: string;
        thumbnail_url: string;
      }>;
      pagination: {
        current_page: number;
        total_pages: number;
        total_items: number;
        items_per_page: number;
      };
    }>('/api/study/activities');
    
    return {
      items: response.items.map(item => ({
        ...item,
        total_sessions: 0, // This will be updated later when we implement session tracking
        last_session_date: undefined
      })),
      total: response.pagination.total_items,
      page: response.pagination.current_page,
      limit: response.pagination.items_per_page,
      total_pages: response.pagination.total_pages
    };
  },
  
  getActivity: async (id: number): Promise<StudyActivity> => {
    const response = await fetchApi<{
      id: number;
      name: string;
      description: string;
      thumbnail_url: string;
      available_groups: Array<{
        id: number;
        name: string;
      }>;
    }>(`/api/study/activities/${id}`);
    
    return {
      id: response.id,
      name: response.name,
      description: response.description,
      thumbnail_url: response.thumbnail_url,
      total_sessions: 0, // Not provided by the backend yet
      last_session_date: undefined, // Not provided by the backend yet
      available_groups: response.available_groups
    };
  },

  // Study Sessions
  getSessions: async (): Promise<StudySession[]> => {
    return fetchApi<StudySession[]>('/api/study/sessions');
  },

  getSession: async (id: number): Promise<StudySession> => {
    return fetchApi<StudySession>(`/api/study/sessions/${id}`);
  },

  getSessionWords: async (sessionId: number): Promise<StudySessionWord[]> => {
    return fetchApi<StudySessionWord[]>(`/api/study/sessions/${sessionId}/words`);
  },

  createSession: async (activityId: number, groupId: number): Promise<StudySession> => {
    return fetchApi<StudySession>('/api/study/sessions', {
      method: 'POST',
      body: JSON.stringify({
        activity_id: activityId,
        group_id: groupId
      })
    });
  },

  updateSession: async (id: number, data: Partial<StudySession>): Promise<StudySession> => {
    return fetchApi<StudySession>(`/api/study/sessions/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(data)
    });
  }
};
