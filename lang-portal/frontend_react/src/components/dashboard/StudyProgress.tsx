import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { BookOpen } from "lucide-react";

interface StudyProgressProps {
  total_words_studied: number;
  total_available_words: number;
}

export default function StudyProgress({ 
  total_words_studied, 
  total_available_words 
}: StudyProgressProps) {
  const progress = total_available_words > 0 
    ? Math.round((total_words_studied / total_available_words) * 100) 
    : 0;

  return (
    <Card className="w-full">
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">Study Progress</CardTitle>
        <BookOpen className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          <p className="text-2xl font-bold">
            {total_words_studied} <span className="text-muted-foreground">/ {total_available_words}</span>
          </p>
          <p className="text-sm text-muted-foreground">Words studied</p>
          <div className="pt-2">
            <Progress value={progress} className="h-2" />
            <p className="text-xs text-muted-foreground mt-1 text-right">
              {progress}% complete
            </p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
