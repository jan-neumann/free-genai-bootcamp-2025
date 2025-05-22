import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { BookOpen, Play } from "lucide-react";
import { Link } from "react-router-dom";

interface StudyActivityCardProps {
  id: number;
  name: string;
  description: string;
  thumbnailUrl?: string;
  totalSessions: number;
  lastSessionDate?: string;
}

export default function StudyActivityCard({ 
  id, 
  name, 
  description, 
  thumbnailUrl, 
  totalSessions,
  lastSessionDate 
}: StudyActivityCardProps) {
  return (
    <Card className="w-full overflow-hidden transition-all hover:shadow-md">
      {thumbnailUrl ? (
        <div className="h-40 bg-gray-100 overflow-hidden">
          <img 
            src={thumbnailUrl} 
            alt={`${name} thumbnail`} 
            className="w-full h-full object-cover"
          />
        </div>
      ) : (
        <div className="h-40 bg-muted flex items-center justify-center">
          <BookOpen className="h-12 w-12 text-muted-foreground" />
        </div>
      )}
      
      <CardHeader className="pb-2">
        <CardTitle className="text-lg">{name}</CardTitle>
        <p className="text-sm text-muted-foreground line-clamp-2">{description}</p>
      </CardHeader>
      
      <CardContent className="pb-2">
        <div className="flex items-center justify-between text-sm text-muted-foreground">
          <span>{totalSessions} session{totalSessions !== 1 ? 's' : ''}</span>
          {lastSessionDate && (
            <span>Last: {new Date(lastSessionDate).toLocaleDateString()}</span>
          )}
        </div>
      </CardContent>
      
      <CardFooter className="flex justify-between gap-2">
        <Button variant="outline" size="sm" asChild>
          <Link to={`/study-activities/${id}`}>
            <BookOpen className="h-4 w-4 mr-2" /> View
          </Link>
        </Button>
        <Button size="sm" asChild>
          <Link to={`/study-activities/${id}/launch`}>
            <Play className="h-4 w-4 mr-2" /> Launch
          </Link>
        </Button>
      </CardFooter>
    </Card>
  );
}
