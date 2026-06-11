import React from 'react';
import MainLayout from './components/layout/MainLayout';
import { Toaster } from '@/components/ui/toaster';
import { TooltipProvider } from '@/components/ui/tooltip';

function App() {
  return (
    <TooltipProvider>
      <MainLayout />
      <Toaster />
    </TooltipProvider>
  );
}

export default App;
