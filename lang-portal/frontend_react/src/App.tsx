import { Routes, Route, Link, Navigate, useLocation } from 'react-router-dom';
import type { LinkProps } from 'react-router-dom';
import DashboardPage from '@/pages/DashboardPage';
import StudyActivitiesPage from '@/pages/StudyActivitiesPage';
import StudyActivityShowPage from '@/pages/StudyActivityShowPage';
import StudyActivityLaunchPage from '@/pages/StudyActivityLaunchPage';
import WordsPage from '@/pages/WordsPage';
import WordShowPage from '@/pages/WordShowPage';
import WordGroupsPage from '@/pages/WordGroupsPage';
import WordGroupShowPage from '@/pages/WordGroupShowPage';
import SessionsPage from '@/pages/SessionsPage';
import SettingsPage from '@/pages/SettingsPage';
import { cn } from '@/lib/utils';
import { BreadcrumbProvider } from '@/contexts/BreadcrumbContext';
import Breadcrumbs from '@/components/layout/Breadcrumbs';

// NavigationMenu components
const NavigationMenu = ({ children }: { children: React.ReactNode }) => (
  <nav className="flex items-center space-x-1">{children}</nav>
);

const NavigationMenuList = ({ children }: { children: React.ReactNode }) => (
  <ul className="flex flex-row space-x-1">{children}</ul>
);

const NavigationMenuItem = ({ children }: { children: React.ReactNode }) => (
  <li>{children}</li>
);

type NavigationMenuLinkProps = LinkProps & {
  active?: boolean;
  children: React.ReactNode;
};

const NavigationMenuLink = ({
  active,
  className,
  children,
  ...props
}: NavigationMenuLinkProps) => (
  <Link
    className={cn(
      'inline-flex items-center justify-center rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50 hover:bg-accent hover:text-accent-foreground h-9 py-2 px-4',
      active ? 'bg-accent text-accent-foreground' : 'text-foreground/60',
      className
    )}
    {...props}
  >
    {children}
  </Link>
);

function App() {
  const location = useLocation();

  const navLinks = [
    { to: "/dashboard", label: "Dashboard" },
    { to: "/study-activities", label: "Study Activities" },
    { to: "/words", label: "Words" },
    { to: "/word-groups", label: "Word Groups" },
    { to: "/sessions", label: "Sessions" },
    { to: "/settings", label: "Settings" },
  ];

  return (
    <BreadcrumbProvider>
      <div className="min-h-screen bg-background flex flex-col items-center">
        <header className="sticky top-0 z-50 w-full border-b border-border/40 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
          <div className="w-full max-w-[1400px] mx-auto px-4">
            <div className="flex h-14 items-center justify-between">
              <NavigationMenu>
              <NavigationMenuList>
                {navLinks.map((navLink) => (
                  <NavigationMenuItem key={navLink.to}>
                    <NavigationMenuLink 
                      to={navLink.to}
                      active={location.pathname.startsWith(navLink.to)}
                    >
                      {navLink.label}
                    </NavigationMenuLink>
                  </NavigationMenuItem>
                ))}
              </NavigationMenuList>
              </NavigationMenu>
            </div>
          </div>
        </header>
        
        <main className="w-full max-w-[1400px] px-4 py-6 mx-auto">
          <Breadcrumbs />
          <Routes>
            <Route path="/" element={<Navigate to="/dashboard" replace />} />
            <Route path="/dashboard" element={<DashboardPage />} />
            <Route path="/study-activities" element={<StudyActivitiesPage />} />
            <Route path="/study-activities/:id" element={<StudyActivityShowPage />} />
            <Route path="/study-activities/:id/launch" element={<StudyActivityLaunchPage />} />
            <Route path="/words" element={<WordsPage />} />
            <Route path="/words/:id" element={<WordShowPage />} />
            <Route path="/word-groups" element={<WordGroupsPage />} />
            <Route path="/word-groups/:id" element={<WordGroupShowPage />} />
            <Route path="/sessions" element={<SessionsPage />} />
            <Route path="/settings" element={<SettingsPage />} />
            {/* Add a catch-all for 404 if desired */}
            {/* <Route path="*" element={<div>Page Not Found</div>} /> */}
          </Routes>
        </main>
        
        <footer className="w-full border-t border-border/40 mt-auto">
          <div className="w-full max-w-[1400px] mx-auto px-4 py-4">
            <p className="text-center text-sm text-muted-foreground">
              Language Learning Portal © {new Date().getFullYear()}
            </p>
          </div>
        </footer>
      </div>
    </BreadcrumbProvider>
  );
}

export default App;
