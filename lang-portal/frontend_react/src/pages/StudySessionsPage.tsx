import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { studyApi, type StudySession } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { format } from 'date-fns';
import { Clock, BookOpen, CheckCircle } from 'lucide-react';

export default function StudySessionsPage() {
  const [sessions, setSessions] = useState<StudySession[]>([]);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    const loadSessions = async () => {
      try {
        const data = await studyApi.getSessions();
        setSessions(data);
      } catch (error) {
        console.error('Error loading sessions:', error);
      } finally {
        setLoading(false);
      }
    };

    loadSessions();
  }, []);

  if (loading) {
    return <div className="container mx-auto p-6">Loading sessions...</div>;
  }

  return (
    <div className="container mx-auto p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold">Study Sessions</h1>
        <Button onClick={() => navigate('/study-activities')}>
          Start New Session
        </Button>
      </div>

      <div className="grid gap-4">
        {sessions.length === 0 ? (
          <div className="text-center py-12">
            <BookOpen className="mx-auto h-12 w-12 text-gray-400" />
            <h3 className="mt-2 text-sm font-medium text-gray-900">No sessions yet</h3>
            <p className="mt-1 text-sm text-gray-500">Get started by creating your first study session.</p>
            <div className="mt-6">
              <Button onClick={() => navigate('/study-activities')}>
                Start Studying
              </Button>
            </div>
          </div>
        ) : (
          sessions.map((session) => (
            <Card 
              key={session.id} 
              className="hover:bg-gray-50 cursor-pointer" 
              onClick={() => navigate(`/study-sessions/${session.id}`)}
            >
              <CardHeader>
                <div className="flex justify-between items-start">
                  <div>
                    <CardTitle>{session.activity_name}</CardTitle>
                    <p className="text-sm text-gray-500">{session.group_name}</p>
                  </div>
                  <div className="flex items-center space-x-2">
                    {session.end_time ? (
                      <span className="text-sm text-green-600 flex items-center">
                        <CheckCircle className="h-4 w-4 mr-1" />
                        Completed
                      </span>
                    ) : (
                      <span className="text-sm text-yellow-600 flex items-center">
                        <Clock className="h-4 w-4 mr-1" />
                        In Progress
                      </span>
                    )}
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-3 gap-4 text-sm">
                  <div>
                    <p className="text-gray-500">Date</p>
                    <p>{format(new Date(session.start_time), 'MMM d, yyyy')}</p>
                  </div>
                  <div>
                    <p className="text-gray-500">Duration</p>
                    <p>
                      {session.end_time 
                        ? `${Math.round((new Date(session.end_time).getTime() - 
                            new Date(session.start_time).getTime()) / 60000)} min`
                        : 'In Progress'}
                    </p>
                  </div>
                  <div>
                    <p className="text-gray-500">Score</p>
                    <p>
                      {session.score !== undefined 
                        ? `${Math.round(session.score * 100)}%` 
                        : 'N/A'}
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))
        )}
      </div>
    </div>
  );
}
