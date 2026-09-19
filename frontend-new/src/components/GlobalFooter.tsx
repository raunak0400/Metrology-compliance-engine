"use client";

import React from "react";
import Link from "next/link";
import { ShieldCheck, Activity, Terminal } from "lucide-react";
import BrandLogo from "@/components/BrandLogo";

export default function GlobalFooter() {
  const currentYear = new Date().getFullYear();

  // Mapped to actual application routes
  const coreModules = [
    { name: "Command Center", href: "/dashboard" },
    { name: "Live Scanner", href: "/scan" },
    { name: "Official Reports", href: "/reports/violations" },
    { name: "Audit History", href: "/history" }
  ];

  return (
    <footer className="mt-auto border-t border-border-color bg-panel transition-colors duration-300 relative z-10">
      <div className="h-1 w-full bg-hazard opacity-80"></div>
      
      <div className="max-w-[1800px] mx-auto px-6 py-10 lg:px-12">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-8 lg:gap-12">
          
          <div className="md:col-span-2 flex flex-col">
            <div className="mb-5">
              {/* Tagline Activated Here */}
              <BrandLogo size="md" showTagline={true} />
            </div>
            <p className="text-xs text-muted leading-relaxed max-w-sm mb-6 font-medium">
              Intelligent enforcement network for Legal Metrology. Automating compliance verification through high-confidence AI extraction and real-time network telemetry.
            </p>
            <div className="flex items-center gap-2 bg-background border border-border-color px-3 py-1.5 rounded-lg w-max shadow-sm mt-auto">
              <span className="relative flex h-2 w-2">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
                <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
              </span>
              <span className="text-[10px] font-mono font-bold text-emerald-600 dark:text-emerald-400 uppercase tracking-widest">Nodes Operational</span>
            </div>
          </div>

          <div>
            <h3 className="text-[10px] font-bold text-foreground uppercase tracking-widest mb-4 flex items-center gap-1.5">
              <Activity className="w-3.5 h-3.5 text-accent" /> Core Modules
            </h3>
            <ul className="space-y-3">
              {coreModules.map((link) => (
                <li key={link.name}>
                  <Link href={link.href} className="text-xs font-bold text-muted hover:text-accent transition-colors">
                    {link.name}
                  </Link>
                </li>
              ))}
            </ul>
          </div>

          <div>
            <h3 className="text-[10px] font-bold text-foreground uppercase tracking-widest mb-4 flex items-center gap-1.5">
              <ShieldCheck className="w-3.5 h-3.5 text-accent" /> Compliance
            </h3>
            <ul className="space-y-3">
              {["LM Act, 2009 Guidelines", "PCR, 2011 Framework", "Data Processing Agreement", "RBAC Security Policy"].map((link) => (
                <li key={link}>
                  <Link href="#" className="text-xs font-bold text-muted hover:text-accent transition-colors">
                    {link}
                  </Link>
                </li>
              ))}
            </ul>
          </div>
        </div>

        <div className="mt-12 pt-6 border-t border-border-color flex flex-col md:flex-row items-center justify-between gap-4">
          <p className="text-[10px] font-bold text-muted uppercase tracking-widest">
            &copy; {currentYear} Legal Metrology Department. All rights reserved.
          </p>
          <div className="flex items-center gap-4 text-[10px] font-mono font-bold text-muted">
            <span className="flex items-center gap-1"><Terminal className="w-3 h-3" /> SIH Project: 26034</span>
            <span className="w-1 h-1 rounded-full bg-border-color"></span>
            <span>v2.4.1 SECURE</span>
          </div>
        </div>
      </div>
    </footer>
  );
}