import React from 'react';
import MainLayout from './components/layout/MainLayout';
import { Toaster } from '@/components/ui/toaster';
import { TooltipProvider } from '@/components/ui/tooltip';
import { ThemeProvider } from '@/lib/theme';

function App() {
  return (
    <ThemeProvider>
      <TooltipProvider>
        <MainLayout />
        <Toaster />
      </TooltipProvider>
    </ThemeProvider>
  );
}

export default App;
