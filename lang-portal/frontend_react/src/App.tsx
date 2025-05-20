import { Routes, Route, Link, Navigate, useLocation } from 'react-router-dom';
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
import {
  NavigationMenu,
  NavigationMenuItem,
  NavigationMenuLink,
  NavigationMenuList,
  navigationMenuTriggerStyle,
} from "@/components/ui/navigation-menu";
import { BreadcrumbProvider } from '@/contexts/BreadcrumbContext';
import Breadcrumbs from '@/components/layout/Breadcrumbs';
// import './App.css'; // We can remove or repurpose this later

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
    <>
      <header className="sticky top-0 z-50 w-full border-b border-border/40 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="container flex h-14 max-w-screen-2xl items-center justify-center">
          <NavigationMenu>
            <NavigationMenuList>
              {navLinks.map((navLink) => (
                <NavigationMenuItem key={navLink.to}>
                  <Link to={navLink.to}>
                    <NavigationMenuLink 
                      active={location.pathname.startsWith(navLink.to)} 
                      className={navigationMenuTriggerStyle()}
                    >
                      {navLink.label}
                    </NavigationMenuLink>
                  </Link>
                </NavigationMenuItem>
              ))}
            </NavigationMenuList>
          </NavigationMenu>
        </div>
      </header>

      <BreadcrumbProvider>
        <Breadcrumbs />
        <main className="p-4 md:p-8">
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
      </BreadcrumbProvider>
    </>
  );
}

export default App;
