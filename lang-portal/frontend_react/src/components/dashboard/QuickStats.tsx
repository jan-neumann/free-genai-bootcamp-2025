import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { BarChart2, BookOpen, Flame, Target } from "lucide-react";

interface QuickStatsProps {
  success_rate: number;
  total_study_sessions: number;
  total_active_groups: number;
  study_streak_days: number;
}

interface StatItemProps {
  value: string | number;
  label: string;
  icon: React.ReactNode;
  color?: string;
}

const StatItem = ({ value, label, icon, color = 'text-primary' }: StatItemProps) => (
  <div className="flex items-start">
    <div className={`p-2 rounded-lg ${color} bg-opacity-10 mr-3`}>
      {icon}
    </div>
    <div>
      <p className="text-2xl font-bold">{value}</p>
      <p className="text-sm text-muted-foreground">{label}</p>
    </div>
  </div>
);

export default function QuickStats({ 
  success_rate, 
  total_study_sessions, 
  total_active_groups, 
  study_streak_days 
}: QuickStatsProps) {
  return (
    <Card className="w-full">
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">Quick Stats</CardTitle>
        <BarChart2 className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <StatItem 
            value={`${success_rate}%`} 
            label="Success Rate" 
            icon={<Target className="h-4 w-4" />} 
            color="text-green-500"
          />
          <StatItem 
            value={total_study_sessions} 
            label="Total Sessions" 
            icon={<BookOpen className="h-4 w-4" />} 
            color="text-blue-500"
          />
          <StatItem 
            value={total_active_groups} 
            label="Active Groups" 
            icon={<BookOpen className="h-4 w-4" />} 
            color="text-purple-500"
          />
          <StatItem 
            value={`${study_streak_days} ${study_streak_days === 1 ? 'day' : 'days'}`} 
            label="Study Streak" 
            icon={<Flame className="h-4 w-4" />} 
            color="text-orange-500"
          />
        </div>
      </CardContent>
    </Card>
  );
}
