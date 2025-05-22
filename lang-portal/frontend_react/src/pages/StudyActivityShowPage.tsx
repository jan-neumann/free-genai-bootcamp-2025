import { useEffect, useState } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { useBreadcrumbs } from '@/contexts/BreadcrumbContext';
import { studyApi, type StudyActivity } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { ArrowLeft, Play, BookOpen, Clock, BarChart2, Calendar } from 'lucide-react';
import { Skeleton } from '@/components/ui/skeleton';

export default function StudyActivityShowPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { setItems } = useBreadcrumbs();
  const [activity, setActivity] = useState<StudyActivity | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchActivity = async () => {
      if (!id) return;
      
      try {
        setLoading(true);
        const data = await studyApi.getActivity(parseInt(id));
        setActivity(data);
        
        // Update breadcrumb with the actual activity name
        setItems([
          { label: 'Dashboard', path: '/' },
          { label: 'Study Activities', path: '/study-activities' },
          { label: data.name }
        ]);
      } catch (err) {
        console.error('Error fetching activity:', err);
        setError('Failed to load activity. Please try again later.');
      } finally {
        setLoading(false);
      }
    };

    fetchActivity();
  }, [id, setItems]);

  if (loading) {
    return (
      <div className="space-y-6">
        <Button variant="outline" size="sm" onClick={() => navigate(-1)} className="mb-4">
          <ArrowLeft className="h-4 w-4 mr-2" /> Back to Activities
        </Button>
        
        <div className="space-y-4">
          <Skeleton className="h-8 w-3/4" />
          <Skeleton className="h-4 w-1/2" />
          <div className="flex space-x-4 pt-4">
            <Skeleton className="h-32 w-32 rounded-lg" />
            <div className="flex-1 space-y-2">
              <Skeleton className="h-6 w-1/3" />
              <Skeleton className="h-4 w-full" />
              <Skeleton className="h-4 w-2/3" />
            </div>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 pt-4">
            {[...Array(3)].map((_, i) => (
              <Skeleton key={i} className="h-32 rounded-lg" />
            ))}
          </div>
        </div>
      </div>
    );
  }

  if (error || !activity) {
    return (
      <div className="space-y-4">
        <Button variant="outline" size="sm" onClick={() => navigate(-1)}>
          <ArrowLeft className="h-4 w-4 mr-2" /> Back to Activities
        </Button>
        <div className="rounded-lg border border-destructive bg-destructive/10 p-4 text-destructive">
          <p>{error || 'Activity not found'}</p>
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
      <div className="flex items-center justify-between">
        <Button variant="outline" size="sm" onClick={() => navigate(-1)}>
          <ArrowLeft className="h-4 w-4 mr-2" /> Back to Activities
        </Button>
        <div className="flex space-x-2">
          <Button variant="outline" size="sm" asChild>
            <Link to={`/study-activities/${id}/launch`}>
              <Play className="h-4 w-4 mr-2" /> Launch
            </Link>
          </Button>
        </div>
      </div>

      <div className="bg-card rounded-lg border shadow-sm">
        <div className="p-6">
          <div className="flex flex-col md:flex-row gap-6">
            {activity.thumbnail_url ? (
              <div className="w-full md:w-48 h-48 bg-muted rounded-lg overflow-hidden flex-shrink-0">
                <img 
                  src={activity.thumbnail_url} 
                  alt={activity.name}
                  className="w-full h-full object-cover"
                />
              </div>
            ) : (
              <div className="w-full md:w-48 h-48 bg-muted rounded-lg flex items-center justify-center flex-shrink-0">
                <BookOpen className="h-12 w-12 text-muted-foreground" />
              </div>
            )}
            
            <div className="flex-1">
              <h1 className="text-3xl font-bold tracking-tight mb-2">{activity.name}</h1>
              <p className="text-muted-foreground mb-4">{activity.description}</p>
              
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mt-6">
                <div className="bg-muted/50 p-4 rounded-lg">
                  <div className="flex items-center space-x-2 text-muted-foreground">
                    <BarChart2 className="h-5 w-5" />
                    <span className="text-sm font-medium">Sessions</span>
                  </div>
                  <p className="text-2xl font-bold mt-1">{activity.total_sessions}</p>
                </div>
                
                <div className="bg-muted/50 p-4 rounded-lg">
                  <div className="flex items-center space-x-2 text-muted-foreground">
                    <Clock className="h-5 w-5" />
                    <span className="text-sm font-medium">Last Session</span>
                  </div>
                  <p className="text-lg font-medium mt-1">
                    {activity.last_session_date 
                      ? new Date(activity.last_session_date).toLocaleDateString() 
                      : 'Never'}
                  </p>
                </div>
                
                <div className="bg-muted/50 p-4 rounded-lg">
                  <div className="flex items-center space-x-2 text-muted-foreground">
                    <Calendar className="h-5 w-5" />
                    <span className="text-sm font-medium">Created</span>
                  </div>
                  <p className="text-lg font-medium mt-1">
                    {new Date().toLocaleDateString()}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Activity Details Card */}
        <div className="bg-card border rounded-lg p-6">
          <h2 className="text-xl font-semibold mb-4">Activity Details</h2>
          <div className="space-y-4">
            <div>
              <h3 className="text-sm font-medium text-muted-foreground">Description</h3>
              <p className="mt-1">{activity.description || 'No description available.'}</p>
            </div>
            <div>
              <h3 className="text-sm font-medium text-muted-foreground">Type</h3>
              <p className="mt-1">Interactive Exercise</p>
            </div>
            <div>
              <h3 className="text-sm font-medium text-muted-foreground">Difficulty</h3>
              <p className="mt-1">Beginner</p>
            </div>
          </div>
        </div>
        
        {/* Recent Sessions */}
        <div className="bg-card border rounded-lg p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-xl font-semibold">Recent Sessions</h2>
            <Button variant="ghost" size="sm" disabled={activity.total_sessions === 0}>
              View All
            </Button>
          </div>
          
          {activity.total_sessions > 0 ? (
            <div className="space-y-4">
              {[...Array(Math.min(3, activity.total_sessions))].map((_, i) => (
                <div key={i} className="flex items-center justify-between p-3 bg-muted/30 rounded-lg">
                  <div>
                    <p className="font-medium">Session {activity.total_sessions - i}</p>
                    <p className="text-sm text-muted-foreground">
                      {new Date(Date.now() - (i * 86400000)).toLocaleDateString()}
                    </p>
                  </div>
                  <Button variant="outline" size="sm">View Details</Button>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8">
              <BookOpen className="h-10 w-10 mx-auto text-muted-foreground mb-2" />
              <p className="text-muted-foreground">No sessions yet. Start your first session!</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}