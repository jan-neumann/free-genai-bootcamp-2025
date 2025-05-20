import { useEffect } from 'react';
import { useBreadcrumbs } from '@/contexts/BreadcrumbContext';

export default function DashboardPage() {
  const { setItems } = useBreadcrumbs();

  useEffect(() => {
    setItems([{ label: 'Dashboard' }]);
  }, [setItems]);

  return (
    <div>
      <h1>Dashboard Page</h1>
    </div>
  );
} 