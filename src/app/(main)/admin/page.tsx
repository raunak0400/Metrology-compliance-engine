"use client";

import React from "react";
import { motion, Variants } from "framer-motion";
import { 
  ShieldAlert, Users, Database, Cpu, 
  MoreVertical, UserPlus, Shield, ScanLine, Key
} from "lucide-react";

// Database-Aligned Data
const systemUsers = [
  { id: "USR-001", name: "Development Admin", email: "admin@legalmetrology.local", role: "ADMIN", status: "Active" },
  { id: "USR-024", name: "Tushar Shah", email: "tushar@legalmetrology.local", role: "INSPECTOR", status: "Active" },
  { id: "USR-089", name: "Rahul Kumar", email: "rahul@legalmetrology.local", role: "REVIEWER", status: "Active" },
  { id: "USR-102", name: "Ananya Patel", email: "ananya@legalmetrology.local", role: "SUPERVISOR", status: "Offline" },
];

const activeRules = [
  { code: "PCR2011-MRP-PRESENCE", title: "Retail sale price (MRP) declaration", severity: "HIGH" },
  { code: "PCR2011-NETQTY-PRESENCE", title: "Net quantity declaration present", severity: "HIGH" },
  { code: "PCR2011-MANUFACTURER", title: "Manufacturer name and address", severity: "MEDIUM" },
];

const containerVariants: Variants = { hidden: { opacity: 0 }, show: { opacity: 1, transition: { staggerChildren: 0.1 } } };
const itemVariants: Variants = { hidden: { opacity: 0, scale: 0.95 }, show: { opacity: 1, scale: 1, transition: { type: "spring", stiffness: 300, damping: 24 } } };

export default function AdminPanelPage() {
  
  // Adjusted for Light/Dark Mode Contrast
  const getRoleBadge = (role: string) => {
    switch(role) {
      case "ADMIN": return "bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/20";
      case "SUPERVISOR": return "bg-purple-500/10 text-purple-600 dark:text-purple-400 border-purple-500/20";
      case "REVIEWER": return "bg-blue-500/10 text-blue-600 dark:text-blue-400 border-blue-500/20";
      case "INSPECTOR": return "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20";
      default: return "bg-gray-500/10 text-gray-600 dark:text-gray-400 border-gray-500/20";
    }
  };

  return (
    <div className="min-h-full bg-background p-4 md:p-6 lg:p-10 text-foreground selection:bg-accent/30 overflow-x-hidden transition-colors duration-300 font-sans">
      
      {/* Header */}
      <motion.div initial={{ opacity: 0, y: -20 }} animate={{ opacity: 1, y: 0 }} className="flex flex-col md:flex-row md:items-center justify-between mb-6 md:mb-8 gap-4">
        <div>
          <h1 className="text-2xl md:text-3xl font-black tracking-tight text-foreground flex items-center gap-3">
            <ShieldAlert className="w-6 h-6 md:w-8 md:h-8 text-accent" />
            System Administration
          </h1>
          <p className="text-xs md:text-sm text-muted mt-1 font-medium">Manage RBAC personnel access and monitor AI rule engine parameters.</p>
        </div>
        <button className="flex items-center justify-center w-full sm:w-auto gap-2 bg-foreground text-background hover:bg-foreground/90 px-5 py-3 md:py-2.5 rounded-xl text-sm font-bold shadow-md transition-all active:scale-95 border border-transparent">
          <UserPlus className="w-4 h-4" /> Provision User
        </button>
      </motion.div>

      <motion.div variants={containerVariants} initial="hidden" animate="show" className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        
        {/* Left Column: AI Engine Status */}
        <div className="space-y-6 lg:col-span-1">
          <motion.div variants={itemVariants} className="bg-panel border border-border-color rounded-2xl shadow-sm overflow-hidden flex flex-col relative">
            <div className="h-1.5 w-full bg-hazard opacity-80 absolute top-0 left-0"></div>
            
            <div className="p-5 md:p-6 mt-2">
              <div className="flex items-center justify-between mb-6">
                <div className="flex items-center gap-3">
                  <div className="p-2.5 bg-background rounded-xl border border-border-color shadow-inner">
                    <Cpu className="w-5 h-5 text-accent" />
                  </div>
                  <div>
                    <h2 className="text-xs md:text-sm font-bold text-foreground uppercase tracking-widest">Rule Engine</h2>
                    <div className="flex items-center gap-2 mt-0.5">
                      <span className="relative flex h-2 w-2">
                        <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
                        <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
                      </span>
                      <span className="text-[10px] font-mono font-bold text-emerald-600 dark:text-emerald-400">STATUS: ONLINE</span>
                    </div>
                  </div>
                </div>
              </div>

              <div className="space-y-3">
                {activeRules.map((rule) => (
                  <div key={rule.code} className="bg-background p-3 rounded-xl border border-border-color shadow-sm flex flex-col gap-1.5 transition-colors">
                    <div className="flex justify-between items-start">
                      <span className="text-[10px] font-mono text-accent font-bold bg-accent/10 px-1.5 py-0.5 rounded border border-accent/20">{rule.code}</span>
                      <span className={`text-[9px] font-bold px-1.5 py-0.5 rounded uppercase ${rule.severity === 'HIGH' ? 'bg-rose-500/10 text-rose-600 dark:text-rose-400' : 'bg-amber-500/10 text-amber-600 dark:text-amber-400'}`}>{rule.severity}</span>
                    </div>
                    <span className="text-xs text-foreground font-medium leading-tight">{rule.title}</span>
                  </div>
                ))}
              </div>
              <button className="w-full mt-5 py-2.5 bg-background hover:bg-border-color/50 border border-border-color rounded-xl text-xs font-bold text-foreground transition-colors shadow-sm flex items-center justify-center gap-2">
                <ScanLine className="w-3.5 h-3.5" /> View All 24 Rules
              </button>
            </div>
          </motion.div>

          <motion.div variants={itemVariants} className="bg-panel border border-border-color rounded-2xl p-5 md:p-6 shadow-sm">
             <div className="flex items-center gap-3 mb-4">
              <Database className="w-5 h-5 text-muted" />
              <h2 className="text-xs font-bold text-foreground uppercase tracking-widest">Database Metrics</h2>
            </div>
            <div className="space-y-4">
              <div>
                <div className="flex justify-between text-[10px] font-bold text-muted mb-1.5 uppercase tracking-wider">
                  <span>Image Blob Storage (S3)</span>
                  <span>42% (2.1 TB)</span>
                </div>
                <div className="w-full bg-background rounded-full h-2 border border-border-color shadow-inner overflow-hidden">
                  <motion.div initial={{ width: 0 }} animate={{ width: '42%' }} transition={{ duration: 1 }} className="bg-accent h-full rounded-full" />
                </div>
              </div>
              <div>
                <div className="flex justify-between text-[10px] font-bold text-muted mb-1.5 uppercase tracking-wider">
                  <span>ML Job Queue</span>
                  <span className="text-emerald-600 dark:text-emerald-400">Stable</span>
                </div>
                <div className="w-full bg-background rounded-full h-2 border border-border-color shadow-inner overflow-hidden">
                  <motion.div initial={{ width: 0 }} animate={{ width: '15%' }} transition={{ duration: 1 }} className="bg-emerald-500 h-full rounded-full" />
                </div>
              </div>
            </div>
          </motion.div>
        </div>

        {/* Right Column: User Management Table */}
        <motion.div variants={itemVariants} className="lg:col-span-2 bg-panel border border-border-color rounded-2xl shadow-sm flex flex-col overflow-hidden h-[600px] lg:h-auto relative">
          <div className="h-1.5 w-full bg-hazard opacity-80 absolute top-0 left-0 z-10"></div>
          
          <div className="px-5 py-5 mt-1.5 border-b border-border-color flex justify-between items-center bg-background/50">
            <div className="flex items-center gap-3">
              <Users className="w-5 h-5 text-muted" />
              <h2 className="text-xs md:text-sm font-bold text-foreground uppercase tracking-widest">RBAC Directory[cite: 1]</h2>
            </div>
            <span className="text-[10px] font-mono bg-background border border-border-color px-2.5 py-1 rounded-lg text-foreground shadow-sm">{systemUsers.length} Users</span>
          </div>

          <div className="overflow-x-auto flex-1 custom-scrollbar bg-panel">
            <table className="w-full text-left border-collapse min-w-[500px]">
              <thead>
                <tr className="border-b border-border-color bg-background/30">
                  <th className="px-6 py-4 text-[10px] font-bold text-muted uppercase tracking-widest">Officer</th>
                  <th className="px-6 py-4 text-[10px] font-bold text-muted uppercase tracking-widest">Role Clearance</th>
                  <th className="px-6 py-4 text-[10px] font-bold text-muted uppercase tracking-widest">Status</th>
                  <th className="px-6 py-4 text-[10px] font-bold text-muted uppercase tracking-widest text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border-color">
                {systemUsers.map((user) => (
                  <tr key={user.id} className="hover:bg-background/80 transition-colors group">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-3">
                        <div className="w-9 h-9 rounded-xl bg-background border border-border-color flex items-center justify-center text-xs font-bold text-foreground shadow-sm group-hover:border-accent/50 transition-colors">
                          {user.name.split(' ').map(n => n[0]).join('')}
                        </div>
                        <div>
                          <p className="text-sm font-bold text-foreground">{user.name}</p>
                          <p className="text-[10px] font-mono text-muted mt-0.5">{user.email}</p>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`px-2.5 py-1 rounded-md border text-[9px] md:text-[10px] font-bold uppercase tracking-widest flex w-max items-center gap-1.5 shadow-sm ${getRoleBadge(user.role)}`}>
                        {user.role === "ADMIN" && <Key className="w-3 h-3" />}
                        {user.role === "SUPERVISOR" && <Shield className="w-3 h-3" />}
                        {user.role}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className={`text-[10px] md:text-[11px] font-bold tracking-widest uppercase flex items-center gap-1.5 ${user.status === 'Active' ? 'text-emerald-600 dark:text-emerald-400' : 'text-muted'}`}>
                        <div className={`w-1.5 h-1.5 rounded-full ${user.status === 'Active' ? 'bg-emerald-500' : 'bg-muted'}`}></div>
                        {user.status}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right">
                      <button className="p-2 hover:bg-background border border-transparent hover:border-border-color rounded-lg transition-all text-muted hover:text-foreground active:scale-95">
                        <MoreVertical className="w-4 h-4" />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </motion.div>

      </motion.div>
    </div>
  );
}