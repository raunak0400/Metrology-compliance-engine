"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useTheme } from "next-themes";
import { AnimatePresence, motion } from "framer-motion";
import { 
  LayoutDashboard, Camera, ClipboardCheck, AlertTriangle, 
  Search, History, LineChart, ShieldAlert, 
  Menu, X, Sun, Moon, Languages
} from "lucide-react";
import GlobalFooter from "@/components/GlobalFooter";
import BrandLogo from "@/components/BrandLogo";

export default function MainLayout({ children }: { children: React.ReactNode }) {
  const [isMobileOpen, setIsMobileOpen] = useState(false);
  const [mounted, setMounted] = useState(false);
  const [isHindi, setIsHindi] = useState(false);
  const { theme, setTheme } = useTheme();
  const pathname = usePathname();

  useEffect(() => {
    setMounted(true);
    
    // Inject Google Translate Script silently
    (window as any).googleTranslateElementInit = () => {
      new (window as any).google.translate.TranslateElement({
        pageLanguage: 'en',
        includedLanguages: 'en,hi',
        autoDisplay: false
      }, 'google_translate_element');
    };

    if (!document.querySelector('script[src*="translate_a/element.js"]')) {
      const addScript = document.createElement('script');
      addScript.src = '//translate.google.com/translate_a/element.js?cb=googleTranslateElementInit';
      addScript.async = true;
      document.body.appendChild(addScript);
    }
  }, []);

  const toggleLanguage = () => {
    const select = document.querySelector('.goog-te-combo') as HTMLSelectElement;
    if (select) {
      select.value = isHindi ? 'en' : 'hi';
      select.dispatchEvent(new Event('change'));
      setIsHindi(!isHindi);
    } else {
      alert("Translation engine is still loading. Please try again in a moment.");
    }
  };

  const navLinks = [
    { name: "Dashboard", href: "/dashboard", icon: LayoutDashboard },
    { name: "Scan & Upload", href: "/scan", icon: Camera },
    { name: "Scan Review", href: "/scan/review", icon: ClipboardCheck },
    { name: "Violation Reports", href: "/reports/violations", icon: AlertTriangle },
    { name: "Repository", href: "/repository", icon: Search },
    { name: "Global History", href: "/history", icon: History },
    { name: "Analytics Engine", href: "/analytics", icon: LineChart },
    { name: "Admin Panel", href: "/admin", icon: ShieldAlert },
  ];

  const toggleTheme = () => setTheme(theme === "dark" ? "light" : "dark");

  return (
    <div className="flex h-screen bg-background text-foreground overflow-hidden font-sans selection:bg-accent/30 transition-colors duration-300">
      
      <div id="google_translate_element" className="hidden"></div>

      <style jsx global>{`
        .goog-te-banner-frame { display: none !important; }
        .skiptranslate { display: none !important; }
        body { top: 0 !important; }
        #goog-gt-tt { display: none !important; }
      `}</style>

      {/* DESKTOP SIDEBAR */}
      <aside className="hidden md:flex flex-col w-[260px] lg:w-72 bg-panel border-r border-border-color shadow-2xl relative z-20 transition-colors duration-300">
        <div className="h-2 w-full bg-hazard opacity-90"></div>
        
        {/* Brand Logo Integration - Desktop */}
        <div className="p-6 border-b border-border-color">
          {/* Tagline Activated Here */}
          <BrandLogo size="md" showTagline={true} />
        </div>

        <nav className="flex-1 overflow-y-auto py-6 px-4 space-y-1.5 custom-scrollbar">
          {navLinks.map((link) => {
            const isActive = link.href === "/dashboard" || link.href === "/scan" || link.href === "/admin" || link.href === "/history" || link.href === "/repository" || link.href === "/analytics"
              ? pathname === link.href 
              : pathname.startsWith(link.href);
            return (
              <Link key={link.name} href={link.href} className={`relative flex items-center gap-3 px-3 py-3.5 rounded-xl text-sm font-bold transition-all duration-200 ${isActive ? 'bg-background border border-border-color text-accent shadow-sm' : 'text-foreground/70 hover:bg-background hover:text-foreground border border-transparent'}`}>
                <link.icon className={`w-5 h-5 ${isActive ? 'text-accent' : 'text-muted'}`} />
                <span className="translate-text">{link.name}</span>
                {isActive && <div className="absolute left-0 w-1 h-8 bg-accent rounded-r-full"></div>}
              </Link>
            );
          })}
        </nav>

        <div className="p-4 border-t border-border-color bg-background/50">
          {mounted && (
            <div className="flex gap-2 mb-4">
              <button onClick={toggleTheme} className="flex-1 flex justify-between items-center bg-panel border border-border-color p-1 rounded-xl shadow-inner relative overflow-hidden group">
                <div className={`flex items-center justify-center w-1/2 py-2 rounded-lg z-10 transition-colors ${theme === 'light' ? 'text-accent font-bold' : 'text-muted'}`}>
                  <Sun className="w-4 h-4" />
                </div>
                <div className={`flex items-center justify-center w-1/2 py-2 rounded-lg z-10 transition-colors ${theme === 'dark' ? 'text-accent font-bold' : 'text-muted'}`}>
                  <Moon className="w-4 h-4" />
                </div>
                <div className={`absolute left-1 top-1 bottom-1 w-[calc(50%-4px)] bg-background border border-border-color rounded-lg shadow-sm z-0 transition-transform duration-300 ${theme === 'dark' ? 'translate-x-full' : 'translate-x-0'}`} />
              </button>

              <button onClick={toggleLanguage} className="flex-1 flex justify-between items-center bg-panel border border-border-color p-1 rounded-xl shadow-inner relative overflow-hidden group">
                <div className={`flex items-center justify-center w-1/2 py-2 rounded-lg z-10 transition-colors ${!isHindi ? 'text-accent font-bold' : 'text-muted'}`}>
                  EN
                </div>
                <div className={`flex items-center justify-center w-1/2 py-2 rounded-lg z-10 transition-colors ${isHindi ? 'text-accent font-bold' : 'text-muted'}`}>
                  HI
                </div>
                <div className={`absolute left-1 top-1 bottom-1 w-[calc(50%-4px)] bg-background border border-border-color rounded-lg shadow-sm z-0 transition-transform duration-300 ${isHindi ? 'translate-x-full' : 'translate-x-0'}`} />
              </button>
            </div>
          )}

          <div className="flex items-center gap-3 px-2">
            <div className="w-9 h-9 rounded-xl bg-background border border-border-color flex items-center justify-center font-bold text-sm shadow-sm">TS</div>
            <div className="flex-1 overflow-hidden">
              <p className="text-xs font-bold truncate">Tushar Shah</p>
              <p className="text-[10px] font-mono text-muted truncate">ID: LM-8492</p>
            </div>
          </div>
        </div>
      </aside>

      {/* MOBILE HEADER */}
      <div className="md:hidden fixed top-0 left-0 right-0 h-16 bg-panel border-b border-border-color flex justify-between items-center px-4 z-40 transition-colors duration-300">
        <BrandLogo size="sm" />
        
        <button onClick={() => setIsMobileOpen(true)} className="p-2 bg-background rounded-xl border border-border-color active:scale-95 transition-transform">
          <Menu className="w-5 h-5 text-foreground" />
        </button>
      </div>

      {/* MOBILE SLIDE-OUT MENU */}
      <AnimatePresence>
        {isMobileOpen && (
          <>
            <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }} onClick={() => setIsMobileOpen(false)} className="md:hidden fixed inset-0 bg-black/60 backdrop-blur-sm z-40" />
            <motion.aside initial={{ x: "100%" }} animate={{ x: 0 }} exit={{ x: "100%" }} transition={{ type: "spring", damping: 25, stiffness: 200 }} className="md:hidden fixed top-0 right-0 bottom-0 w-[80%] max-w-sm bg-panel border-l border-border-color shadow-2xl z-50 flex flex-col">
              <div className="h-2 w-full bg-hazard opacity-90"></div>
              <div className="p-5 border-b border-border-color flex justify-between items-center">
                <span className="text-xs font-bold text-muted uppercase tracking-widest">Navigation</span>
                <button onClick={() => setIsMobileOpen(false)} className="p-1.5 bg-background rounded-lg border border-border-color active:scale-90">
                  <X className="w-4 h-4 text-foreground" />
                </button>
              </div>
              <nav className="flex-1 overflow-y-auto p-4 space-y-2">
                {navLinks.map((link) => {
                  const isActive = link.href === "/dashboard" || link.href === "/scan" || link.href === "/admin" || link.href === "/history" || link.href === "/repository" || link.href === "/analytics"
                    ? pathname === link.href 
                    : pathname.startsWith(link.href);
                  return (
                    <Link key={link.name} href={link.href} onClick={() => setIsMobileOpen(false)} className={`flex items-center gap-3 px-4 py-4 rounded-xl text-sm font-bold transition-all ${isActive ? 'bg-background border border-border-color text-accent shadow-sm' : 'text-foreground/70 active:bg-background border border-transparent'}`}>
                      <link.icon className={`w-5 h-5 ${isActive ? 'text-accent' : 'text-muted'}`} />
                      {link.name}
                    </Link>
                  );
                })}
              </nav>
              {mounted && (
                <div className="p-6 border-t border-border-color flex justify-between items-center gap-4">
                  <button onClick={toggleTheme} className="flex-1 p-3 flex justify-center items-center bg-background border border-border-color rounded-xl shadow-inner active:scale-90 transition-transform">
                    {theme === 'dark' ? <Sun className="w-5 h-5 text-accent" /> : <Moon className="w-5 h-5 text-accent" />}
                  </button>
                  <button onClick={toggleLanguage} className="flex-1 p-3 flex justify-center items-center gap-2 bg-background border border-border-color rounded-xl shadow-inner active:scale-90 transition-transform font-bold text-accent text-sm">
                    <Languages className="w-5 h-5" /> {isHindi ? 'EN' : 'HI'}
                  </button>
                </div>
              )}
            </motion.aside>
          </>
        )}
      </AnimatePresence>

      {/* MAIN CONTENT AREA */}
      <main className="flex-1 flex flex-col h-full overflow-y-auto bg-background transition-colors duration-300 custom-scrollbar relative z-10 pt-16 md:pt-0">
        
        {/* Auto-Translation Legal Disclaimer */}
        <AnimatePresence>
          {isHindi && (
            <motion.div initial={{ height: 0, opacity: 0 }} animate={{ height: 'auto', opacity: 1 }} exit={{ height: 0, opacity: 0 }} className="bg-amber-500/10 border-b border-amber-500/20 px-4 py-2 flex items-center justify-center text-center z-50">
              <span className="text-[10px] md:text-xs font-bold text-amber-600 dark:text-amber-400 flex items-center gap-2">
                <AlertTriangle className="w-3.5 h-3.5 shrink-0" />
                This site is auto-translated by a third-party tool. Translations may not be 100% legally accurate.
              </span>
            </motion.div>
          )}
        </AnimatePresence>

        <div className="flex-1">
          {children}
        </div>
        
        <GlobalFooter />
      </main>
    </div>
  );
}