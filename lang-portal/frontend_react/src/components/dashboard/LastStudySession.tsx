import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { formatDistanceToNow, parseISO } from "date-fns";
import { Activity } from "lucide-react";

interface LastStudySessionProps {
  id: number;
  group_id: number;
  created_at: string;
  study_activity_id: number;
  group_name: string;
  activity_name: string;
}

export default function LastStudySession({ 
  activity_name, 
  created_at, 
  group_name 
}: LastStudySessionProps) {
  // Safely parse the date, defaulting to now if invalid
  const lastUsedDate = created_at ? parseISO(created_at) : new Date();
  const lastUsedTime = formatDistanceToNow(lastUsedDate, { addSuffix: true });

  return (
    <Card className="w-full">
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">Last Study Session</CardTitle>
        <Activity className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{activity_name}</div>
        <p className="text-muted-foreground">Last used {lastUsedTime}</p>
        <div className="mt-4">
          <p className="text-sm">Group: {group_name}</p>
        </div>
      </CardContent>
    </Card>
  );
}
