import { useEffect, useState } from 'react';
import { useBreadcrumbs } from '@/contexts/BreadcrumbContext';
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import LastStudySession from '@/components/dashboard/LastStudySession';
import StudyProgress from '@/components/dashboard/StudyProgress';
import QuickStats from '@/components/dashboard/QuickStats';
import StartStudyingButton from '@/components/dashboard/StartStudyingButton';

interface LastStudySessionData {
  id: number;
  group_id: number;
  created_at: string;
  study_activity_id: number;
  group_name: string;
  activity_name: string;
}

interface StudyProgressData {
  total_words_studied: number;
  total_available_words: number;
}

interface QuickStatsData {
  success_rate: number;
  total_study_sessions: number;
  total_active_groups: number;
  study_streak_days: number;
}

interface DashboardData {
  lastStudySession: LastStudySessionData | null;
  studyProgress: StudyProgressData;
  quickStats: QuickStatsData;
}

export default function DashboardPage() {
  const { setItems } = useBreadcrumbs();
  const [dashboardData, setDashboardData] = useState<DashboardData>({
    lastStudySession: null,
    studyProgress: { total_words_studied: 0, total_available_words: 0 },
    quickStats: { 
      success_rate: 0, 
      total_study_sessions: 0, 
      total_active_groups: 0, 
      study_streak_days: 0 
    }
  });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setItems([{ label: 'Dashboard' }]);

    // Fetch dashboard data
    const fetchData = async () => {
      try {
        const [lastSessionRes, progressRes, statsRes] = await Promise.allSettled([
          fetch('/api/v1/dashboard/last-session'),
          fetch('/api/v1/dashboard/progress'),
          fetch('/api/v1/dashboard/stats')
        ]);

        // Handle responses
        const lastSession = lastSessionRes.status === 'fulfilled' && lastSessionRes.value.ok 
          ? await lastSessionRes.value.json() 
          : null;
          
        const progress = progressRes.status === 'fulfilled' && progressRes.value.ok
          ? await progressRes.value.json()
          : { total_words_studied: 0, total_available_words: 0 };
          
        const stats = statsRes.status === 'fulfilled' && statsRes.value.ok
          ? await statsRes.value.json()
          : { 
              success_rate: 0, 
              total_study_sessions: 0, 
              total_active_groups: 0, 
              study_streak_days: 0 
            };

        setDashboardData({
          lastStudySession: lastSession,
          studyProgress: progress,
          quickStats: stats
        });
      } catch (err) {
        console.error('Error fetching dashboard data:', err);
        setError('Failed to load dashboard data. Please try again later.');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [setItems]);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center h-64 space-y-4">
        <div className="text-destructive">{error}</div>
        <button 
          onClick={() => window.location.reload()} 
          className="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90"
        >
          Retry
        </button>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center h-64 space-y-4">
        <div className="text-destructive">{error}</div>
        <button 
          onClick={() => window.location.reload()} 
          className="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90"
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {dashboardData.lastStudySession ? (
          <LastStudySession {...dashboardData.lastStudySession} />
        ) : (
          <Card className="w-full">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Last Study Session</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-muted-foreground text-sm">No recent study sessions found</div>
            </CardContent>
          </Card>
        )}
        <StudyProgress {...dashboardData.studyProgress} />
        <QuickStats {...dashboardData.quickStats} />
      </div>
      <div className="flex justify-center">
        <StartStudyingButton />
      </div>
    </div>
  );
} 