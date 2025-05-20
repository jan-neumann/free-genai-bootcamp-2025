import { useEffect } from 'react';
import { useBreadcrumbs } from '@/contexts/BreadcrumbContext';

export default function StudyActivitiesPage() {
  const { setItems } = useBreadcrumbs();

  useEffect(() => {
    setItems([
      { label: 'Dashboard', path: '/dashboard' },
      { label: 'Study Activities' }
    ]);
  }, [setItems]);

  return (
    <div>
      <h1>Study Activities Page</h1>
    </div>
  );
} 