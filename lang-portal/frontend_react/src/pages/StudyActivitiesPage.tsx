import { useEffect, useState } from 'react';
import { useBreadcrumbs } from '@/contexts/BreadcrumbContext';
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import StudyActivityCard from '@/components/study/StudyActivityCard';
import { Button } from "@/components/ui/button";
import { Plus, BookOpen } from "lucide-react";
import { studyApi, type StudyActivity } from '@/lib/api';

export default function StudyActivitiesPage() {
  const { setItems } = useBreadcrumbs();
  const [activities, setActivities] = useState<StudyActivity[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setItems([
      { label: 'Dashboard', path: '/' },
      { label: 'Study Activities' }
    ]);

    // Fetch study activities
    const fetchActivities = async () => {
      try {
        const data = await studyApi.getActivities();
        // For now, we'll add mock data for total_sessions and last_session_date
        // since the backend doesn't return these fields yet
        const activitiesWithMockData = data.items.map(activity => ({
          ...activity,
          total_sessions: Math.floor(Math.random() * 20), // Mock data
          last_session_date: new Date(Date.now() - Math.floor(Math.random() * 30) * 24 * 60 * 60 * 1000).toISOString() // Mock data
        }));
        setActivities(activitiesWithMockData);
      } catch (err) {
        console.error('Error fetching study activities:', err);
        setError('Failed to load study activities. Please make sure the backend server is running.');
      } finally {
        setLoading(false);
      }
    };

    fetchActivities();
  }, [setItems]);

  if (loading) {
    return (
      <div className="space-y-6">
        <div className="flex justify-between items-center">
          <h1 className="text-3xl font-bold">Study Activities</h1>
          <Button disabled>
            <Plus className="h-4 w-4 mr-2" /> New Activity
          </Button>
        </div>
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {[...Array(6)].map((_, i) => (
            <Card key={i} className="overflow-hidden">
              <Skeleton className="h-40 w-full" />
              <CardHeader>
                <Skeleton className="h-6 w-3/4 mb-2" />
                <Skeleton className="h-4 w-full" />
                <Skeleton className="h-4 w-2/3 mt-1" />
              </CardHeader>
              <CardContent className="flex justify-between gap-2">
                <Skeleton className="h-9 w-20" />
                <Skeleton className="h-9 w-20" />
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="space-y-4">
        <div className="flex justify-between items-center">
          <h1 className="text-3xl font-bold">Study Activities</h1>
          <Button disabled>
            <Plus className="h-4 w-4 mr-2" /> New Activity
          </Button>
        </div>
        <div className="rounded-lg border border-destructive bg-destructive/10 p-4 text-destructive">
          <p>{error}</p>
          <Button 
            variant="outline" 
            size="sm" 
            className="mt-2"
            onClick={() => window.location.reload()}
          >
            Retry
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold">Study Activities</h1>
        <Button disabled>
          <Plus className="h-4 w-4 mr-2" /> New Activity
        </Button>
      </div>
      
      {activities.length > 0 ? (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {activities.map((activity) => (
            <StudyActivityCard
              key={activity.id}
              id={activity.id}
              name={activity.name}
              description={activity.description}
              thumbnailUrl={activity.thumbnail_url}
              totalSessions={activity.total_sessions}
              lastSessionDate={activity.last_session_date}
            />
          ))}
        </div>
      ) : (
        <div className="flex flex-col items-center justify-center py-12 border-2 border-dashed rounded-lg">
          <BookOpen className="h-12 w-12 text-muted-foreground mb-4" />
          <h3 className="text-lg font-medium">No study activities found</h3>
          <p className="text-muted-foreground text-sm mt-1">Get started by creating a new study activity</p>
          <Button className="mt-4">
            <Plus className="h-4 w-4 mr-2" /> Create Activity
          </Button>
        </div>
      )}
    </div>
  );
}