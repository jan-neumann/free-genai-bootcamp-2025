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

export interface StudyActivity {
  id: number;
  name: string;
  description: string;
  thumbnail_url?: string;
  total_sessions: number;
  last_session_date?: string;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
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
    return fetchApi<StudyActivity>(`/study/activities/${id}`);
  },
};
