import { useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { useBreadcrumbs } from '@/contexts/BreadcrumbContext';

export default function StudyActivityShowPage() {
  const { id } = useParams<{ id: string }>();
  const { setItems } = useBreadcrumbs();

  // In a real app, you would fetch the activity name based on the id
  const activityName = `Activity ${id}`; // Placeholder

  useEffect(() => {
    setItems([
      { label: 'Dashboard', path: '/dashboard' },
      { label: 'Study Activities', path: '/study-activities' },
      { label: activityName } // The current page, not a link
    ]);
  }, [setItems, activityName]);

  return (
    <div>
      <h1>Study Activity Show Page</h1>
      <p>Activity ID: {id}</p>
    </div>
  );
} 