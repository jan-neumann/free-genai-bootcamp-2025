import { Link } from 'react-router-dom';
import { useBreadcrumbs, type BreadcrumbItem } from '@/contexts/BreadcrumbContext';
import { ChevronRight } from 'lucide-react'; // Using a lucide icon for the separator

export default function Breadcrumbs() {
  const { items } = useBreadcrumbs();

  if (!items || items.length === 0) {
    return null; // Don't render anything if there are no breadcrumbs
  }

  return (
    <nav aria-label="Breadcrumb" className="container max-w-screen-2xl py-3 px-4 md:px-8 border-b border-border/40">
      <ol className="flex items-center space-x-1.5 text-sm text-muted-foreground">
        {items.map((item, index) => (
          <li key={index} className="flex items-center">
            {index > 0 && (
              <ChevronRight className="h-4 w-4 mr-1.5" />
            )}
            {item.path ? (
              <Link 
                to={item.path} 
                className="hover:text-foreground transition-colors"
              >
                {item.label}
              </Link>
            ) : (
              <span className="font-medium text-foreground">{item.label}</span>
            )}
          </li>
        ))}
      </ol>
    </nav>
  );
} 