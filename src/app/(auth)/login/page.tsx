"use client";

import React, { useState, useEffect } from "react";
import { motion, Variants } from "framer-motion";
import { 
  Mail, Lock, ArrowRight, ShieldCheck, 
  Fingerprint, Key, Hexagon, Cpu 
} from "lucide-react";
import { useRouter } from "next/navigation";
import BrandLogo from "@/components/BrandLogo";

type AuthRole = "INSPECTOR" | "REVIEWER" | "SUPERVISOR" | "ADMIN";

const containerVariants = {
  hidden: { opacity: 0 },
  show: { opacity: 1, transition: { staggerChildren: 0.1, delayChildren: 0.2 } }
};

const itemVariants = {
  hidden: { opacity: 0, y: 15 },
  show: { opacity: 1, y: 0, transition: { type: "spring", stiffness: 400, damping: 30 } }
};

export default function PremiumLoginPage() {
  const router = useRouter();
  const [role, setRole] = useState<AuthRole>("INSPECTOR");
  const [isLoading, setIsLoading] = useState(false);
  const [mounted, setMounted] = useState(false);

  useEffect(() => setMounted(true), []);

  const handleLogin = (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setTimeout(() => {
      router.push("/dashboard");
    }, 1200);
  };

  const roles: { id: AuthRole; label: string; icon?: any }[] = [
    { id: "INSPECTOR", label: "Inspector" },
    { id: "REVIEWER", label: "Reviewer" },
    { id: "SUPERVISOR", label: "Supervisor" },
    { id: "ADMIN", label: "Admin", icon: Key }
  ];

  if (!mounted) return null;

  return (
    <div className="min-h-screen bg-background flex flex-col lg:flex-row font-sans selection:bg-accent/30 relative overflow-hidden transition-colors duration-300">
      
      {/* Global Vector Technical Grid */}
      <div className="absolute inset-0 pointer-events-none opacity-[0.03] dark:opacity-[0.05] z-0" 
           style={{ backgroundImage: 'radial-gradient(var(--foreground) 1.5px, transparent 1.5px)', backgroundSize: '32px 32px' }} />

      {/* LEFT PANEL: Branding & Vector Art */}
      <div className="w-full lg:flex-1 relative flex flex-col justify-center lg:justify-between p-8 lg:p-14 lg:border-r border-border-color z-10 pt-12 lg:pt-14 overflow-hidden bg-panel/30">
        
        {/* Animated Vector Watermark */}
        <motion.div 
          animate={{ rotate: 360 }} 
          transition={{ repeat: Infinity, duration: 120, ease: "linear" }}
          className="absolute -top-32 -left-32 text-foreground opacity-[0.02] dark:opacity-[0.03] pointer-events-none"
        >
          <Hexagon className="w-[600px] h-[600px]" strokeWidth={0.5} />
        </motion.div>

        <motion.div initial={{ opacity: 0, x: -20 }} animate={{ opacity: 1, x: 0 }} transition={{ duration: 0.6, type: "spring" }} className="mb-8 lg:mb-12 relative z-10">
          <BrandLogo size="lg" showTagline={true} />
        </motion.div>

        <motion.div variants={containerVariants} initial="hidden" animate="show" className="hidden lg:block max-w-md relative z-10">
          <motion.div variants={itemVariants} className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full bg-background border border-border-color mb-6 shadow-sm">
            <Cpu className="w-4 h-4 text-emerald-600 dark:text-emerald-400" />
            <span className="text-[10px] font-bold text-muted tracking-widest uppercase">SIH 2026 Engine Active</span>
          </motion.div>
          <motion.h1 variants={itemVariants} className="text-4xl lg:text-5xl font-black text-foreground leading-[1.1] mb-5 tracking-tight">
            Intelligent <br /> Compliance.
          </motion.h1>
          <motion.p variants={itemVariants} className="text-muted text-base lg:text-lg leading-relaxed font-medium">
            Automated visual scanning, high-confidence rule extraction, and real-time legal metrology network telemetry.
          </motion.p>
        </motion.div>

        <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ delay: 0.8 }} className="hidden lg:flex items-center justify-between text-[10px] font-mono text-muted border-t border-border-color pt-6 font-bold tracking-widest relative z-10">
          <span>© 2026 LEGAL METROLOGY</span>
          <span className="flex items-center gap-1.5"><ShieldCheck className="w-3.5 h-3.5 text-accent" /> v2.4.1 SECURE</span>
        </motion.div>
      </div>

      {/* RIGHT PANEL: Interactive Login Form */}
      <div className="w-full lg:flex-1 flex flex-col justify-start lg:justify-center items-center px-6 pb-12 lg:p-12 relative z-10 flex-grow bg-background">
        <motion.div 
          initial={{ opacity: 0, scale: 0.95 }} animate={{ opacity: 1, scale: 1 }} transition={{ duration: 0.5, type: "spring", damping: 25 }}
          className="w-full max-w-[420px] space-y-8 bg-panel lg:bg-transparent p-8 lg:p-0 rounded-[2rem] border border-border-color lg:border-none shadow-sm lg:shadow-none relative"
        >
          
          <div className="text-center lg:text-left">
            <h2 className="text-2xl lg:text-3xl font-black text-foreground mb-2 tracking-tight">Access Portal</h2>
            <p className="text-muted text-sm font-medium">Enter your credentials to access the command center.</p>
          </div>

          <div className="bg-background p-1.5 rounded-2xl border border-border-color grid grid-cols-2 gap-1.5 shadow-inner relative">
            {roles.map((r) => {
              const isActive = role === r.id;
              return (
                <button
                  key={r.id}
                  type="button"
                  onClick={() => setRole(r.id)}
                  className={`relative flex items-center justify-center gap-2 text-xs font-black tracking-tight py-3.5 rounded-xl transition-colors outline-none z-10 ${isActive ? "text-background" : "text-muted hover:text-foreground"}`}
                >
                  {isActive && (
                    <motion.div 
                      layoutId="activeRoleTab" 
                      className="absolute inset-0 bg-foreground rounded-xl shadow-sm z-0" 
                      transition={{ type: "spring", stiffness: 400, damping: 30 }}
                    />
                  )}
                  <span className="relative z-10 flex items-center gap-2">
                    {r.icon && <r.icon className="w-3.5 h-3.5" />}
                    {r.label}
                  </span>
                </button>
              );
            })}
          </div>

          <form onSubmit={handleLogin} className="space-y-5">
            <motion.div variants={containerVariants} initial="hidden" animate="show" className="space-y-5">
              
              <motion.div variants={itemVariants} className="space-y-2 group">
                <label className="text-[10px] font-bold text-muted uppercase tracking-widest pl-1 transition-colors group-focus-within:text-accent">Email Address</label>
                <div className="relative">
                  <Mail className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-muted group-focus-within:text-accent transition-colors" />
                  <input 
                    type="email" 
                    required
                    placeholder={role === "ADMIN" ? "admin@legalmetrology.local" : "officer@legalmetrology.local"}
                    className="w-full bg-background border border-border-color rounded-xl py-4 pl-12 pr-4 text-foreground outline-none focus:border-accent focus:ring-1 focus:ring-accent transition-all placeholder:text-muted shadow-inner text-sm font-bold" 
                  />
                </div>
              </motion.div>

              <motion.div variants={itemVariants} className="space-y-2 group">
                <div className="flex justify-between items-center pl-1 pr-1">
                  <label className="text-[10px] font-bold text-muted uppercase tracking-widest transition-colors group-focus-within:text-accent">Password</label>
                  <a href="#" className="text-[10px] font-bold text-accent hover:text-accent/80 transition-colors">Forgot?</a>
                </div>
                <div className="relative">
                  <Lock className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-muted group-focus-within:text-accent transition-colors" />
                  <input 
                    type="password" 
                    required
                    placeholder="••••••••••••"
                    className="w-full bg-background border border-border-color rounded-xl py-4 pl-12 pr-4 text-foreground outline-none focus:border-accent focus:ring-1 focus:ring-accent transition-all placeholder:text-muted shadow-inner text-sm font-bold" 
                  />
                </div>
              </motion.div>

              <motion.button 
                variants={itemVariants}
                whileHover={{ scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
                type="submit" 
                disabled={isLoading}
                className="w-full bg-accent hover:bg-accent/90 text-background font-black tracking-tight py-4 rounded-xl shadow-md transition-all flex items-center justify-center gap-2 disabled:opacity-70 disabled:cursor-not-allowed text-sm mt-2 border border-transparent"
              >
                {isLoading ? (
                  <motion.div animate={{ rotate: 360 }} transition={{ repeat: Infinity, duration: 1, ease: "linear" }}>
                    <Fingerprint className="w-5 h-5" />
                  </motion.div>
                ) : (
                  <>Authenticate Identity <ArrowRight className="w-4 h-4" /></>
                )}
              </motion.button>
              
            </motion.div>
          </form>

          {/* Hackathon Helper */}
          <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.8 }} className="mt-8 p-4 border border-accent/20 bg-accent/5 rounded-2xl text-center shadow-sm flex items-start gap-3 relative overflow-hidden">
            <div className="absolute top-0 left-0 bottom-0 w-1 bg-accent"></div>
            <ShieldCheck className="w-5 h-5 text-accent shrink-0 mt-0.5" />
            <div className="text-left">
              <p className="text-[10px] text-accent mb-0.5 font-bold uppercase tracking-widest">SIH Demo Mode Active</p>
              <p className="text-xs text-muted font-medium">Click <strong className="text-foreground">Authenticate Identity</strong> to bypass validation logic.</p>
            </div>
          </motion.div>
          
        </motion.div>
      </div>
    </div>
  );
}