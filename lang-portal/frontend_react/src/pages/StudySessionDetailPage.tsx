import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { studyApi } from '@/lib/api';
import type { StudySession, StudySessionWord } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { format } from 'date-fns';
import { ArrowLeft, CheckCircle, XCircle, Clock } from 'lucide-react';
import { DataTable } from '@/components/ui/data-table';
import type { ColumnDef } from '@tanstack/react-table';

const columns: ColumnDef<StudySessionWord>[] = [
  {
    accessorKey: 'word',
    header: 'Word',
  },
  {
    accessorKey: 'translation',
    header: 'Translation',
  },
  {
    accessorKey: 'example',
    header: 'Example',
  },
  {
    accessorKey: 'correct',
    header: 'Status',
    cell: ({ row }) => (
      <div className="flex items-center">
        {row.original.reviewed ? (
          row.original.correct ? (
            <CheckCircle className="h-4 w-4 text-green-500 mr-2" />
          ) : (
            <XCircle className="h-4 w-4 text-red-500 mr-2" />
          )
        ) : (
          <Clock className="h-4 w-4 text-gray-400 mr-2" />
        )}
        {row.original.reviewed 
          ? (row.original.correct ? 'Correct' : 'Incorrect') 
          : 'Not Reviewed'}
      </div>
    ),
  },
];

export default function StudySessionDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [session, setSession] = useState<StudySession | null>(null);
  const [words, setWords] = useState<StudySessionWord[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const loadData = async () => {
      try {
        const [sessionData, wordsData] = await Promise.all([
          studyApi.getSession(Number(id)),
          studyApi.getSessionWords(Number(id))
        ]);
        setSession(sessionData);
        setWords(wordsData);
      } catch (error) {
        console.error('Error loading session:', error);
      } finally {
        setLoading(false);
      }
    };

    loadData();
  }, [id]);

  if (loading) {
    return <div className="container mx-auto p-6">Loading session details...</div>;
  }

  if (!session) {
    return <div className="container mx-auto p-6">Session not found</div>;
  }

  return (
    <div className="container mx-auto p-6">
      <Button
        variant="ghost"
        onClick={() => navigate(-1)}
        className="mb-6"
      >
        <ArrowLeft className="h-4 w-4 mr-2" /> Back to Sessions
      </Button>

      <div className="grid gap-6">
        <Card>
          <CardHeader>
            <div className="flex justify-between items-start">
              <div>
                <CardTitle>{session.activity_name}</CardTitle>
                <p className="text-sm text-gray-500">{session.group_name}</p>
              </div>
              <div className="text-right">
                <p className="text-sm text-gray-500">Started</p>
                <p>{format(new Date(session.start_time), 'MMM d, yyyy h:mm a')}</p>
                {session.end_time && (
                  <>
                    <p className="text-sm text-gray-500 mt-2">Duration</p>
                    <p>
                      {Math.round((new Date(session.end_time).getTime() - 
                         new Date(session.start_time).getTime()) / 60000)} minutes
                    </p>
                  </>
                )}
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-3 gap-4">
              <div>
                <p className="text-sm text-gray-500">Status</p>
                <p className="flex items-center">
                  {session.end_time ? (
                    <>
                      <CheckCircle className="h-4 w-4 text-green-500 mr-2" />
                      Completed
                    </>
                  ) : (
                    <>
                      <Clock className="h-4 w-4 text-yellow-500 mr-2" />
                      In Progress
                    </>
                  )}
                </p>
              </div>
              <div>
                <p className="text-sm text-gray-500">Score</p>
                <p>
                  {session.score !== undefined 
                    ? `${Math.round(session.score * 100)}%` 
                    : 'N/A'}
                </p>
              </div>
              <div>
                <p className="text-sm text-gray-500">Words Reviewed</p>
                <p>{words.filter(w => w.reviewed).length} of {words.length}</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Words in this Session</CardTitle>
          </CardHeader>
          <CardContent>
            <DataTable
              columns={columns}
              data={words}
            />
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
