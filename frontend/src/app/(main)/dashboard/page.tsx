"use client";

import React from "react";
import { motion, Variants } from "framer-motion";
import { 
  AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
  PieChart, Pie, Cell
} from "recharts";
import { 
  ShieldCheck, AlertOctagon, Activity, PackageSearch, 
  ArrowUpRight, ArrowDownRight, Clock, MoreVertical, ScanLine, Zap
} from "lucide-react";

// Database-Aligned Mock Data
const complianceData = [
  { time: "08:00", compliant: 120, violations: 15 },
  { time: "10:00", compliant: 210, violations: 45 },
  { time: "12:00", compliant: 380, violations: 90 },
  { time: "14:00", compliant: 310, violations: 65 },
  { time: "16:00", compliant: 450, violations: 110 },
  { time: "18:00", compliant: 390, violations: 85 },
];

const categoryData = [
  { name: "Packaged Food", count: 845, color: "#f59e0b" },
  { name: "Cosmetics & Toiletries", count: 432, color: "#6366f1" },
  { name: "Consumer Electronics", count: 289, color: "#06b6d4" },
  { name: "Household & Cleaning", count: 195, color: "#ec4899" },
  { name: "General Commodity", count: 112, color: "#10b981" },
];

const recentScans = [
  { id: "INS-2026-8921", product: "Organic Honey 500g", status: "NON_COMPLIANT", time: "2m ago", conf: "94%" },
  { id: "INS-2026-8922", product: "Almond Milk 1L", status: "COMPLIANT", time: "14m ago", conf: "98%" },
  { id: "INS-2026-8923", product: "Face Wash 150ml", status: "MANUAL_REVIEW", time: "28m ago", conf: "72%" },
  { id: "INS-2026-8924", product: "LED Bulb 9W", status: "COMPLIANT", time: "45m ago", conf: "99%" },
];

const containerVariants: Variants = {
  hidden: { opacity: 0 },
  show: { opacity: 1, transition: { staggerChildren: 0.1 } }
};

const itemVariants: Variants = {
  hidden: { opacity: 0, y: 20 },
  show: { opacity: 1, y: 0, transition: { type: "spring", stiffness: 300, damping: 24 } }
};

export default function PremiumDashboard() {
  return (
    <div className="min-h-full bg-background p-4 md:p-6 lg:p-8 text-foreground overflow-x-hidden font-sans selection:bg-accent/30 transition-colors duration-300 relative">
      
      {/* Global Vector Technical Grid */}
      <div className="absolute inset-0 pointer-events-none opacity-5 dark:opacity-10 z-0" 
           style={{ backgroundImage: 'radial-gradient(var(--foreground) 1px, transparent 1px)', backgroundSize: '24px 24px' }} />

      {/* Header Section */}
      <motion.div initial={{ opacity: 0, y: -20 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.5 }} className="flex flex-col sm:flex-row justify-between items-start sm:items-center mb-6 md:mb-8 gap-4 relative z-10">
        <div>
          <h1 className="text-2xl md:text-3xl font-black tracking-tight text-foreground flex items-center gap-2">
            Command Center
          </h1>
          <div className="flex items-center gap-2 mt-1.5">
            <span className="relative flex h-2 w-2 md:h-2.5 md:w-2.5">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
              <span className="relative inline-flex rounded-full h-2 w-2 md:h-2.5 md:w-2.5 bg-emerald-500 shadow-[0_0_10px_rgba(16,185,129,0.8)]"></span>
            </span>
            <p className="text-[11px] md:text-xs font-bold text-muted uppercase tracking-widest">Live Network Telemetry</p>
          </div>
        </div>
        
        <div className="flex items-center gap-2 md:gap-3 bg-panel border border-border-color hover:border-accent/50 transition-colors px-3 py-2 md:px-4 md:py-2.5 rounded-xl shadow-sm w-full sm:w-auto justify-between sm:justify-start group cursor-default">
          <div className="flex items-center gap-2">
            <Clock className="w-4 h-4 text-accent group-hover:animate-pulse" />
            <span className="text-xs md:text-sm font-bold text-foreground">
              {new Date().toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' })}
            </span>
          </div>
          <div className="h-4 w-px bg-border-color mx-1 hidden sm:block"></div>
          <span className="text-[10px] md:text-[11px] font-mono bg-accent/10 text-accent px-2 py-1 rounded border border-accent/20 flex items-center gap-1 font-bold">
            <Zap className="w-3 h-3" /> v2.4.1
          </span>
        </div>
      </motion.div>

      {/* Animated KPI Grid */}
      <motion.div variants={containerVariants} initial="hidden" animate="show" className="grid grid-cols-2 lg:grid-cols-4 gap-3 md:gap-6 mb-6 md:mb-8 relative z-10">
        {[
          { title: "Today's Scans", value: "2,491", trend: "+14.2%", isUp: true, icon: PackageSearch, color: "text-blue-600 dark:text-blue-400", bg: "bg-blue-500/10", border: "border-blue-500/20" },
          { title: "Violations", value: "342", trend: "-5.1%", isUp: false, icon: AlertOctagon, color: "text-rose-600 dark:text-rose-400", bg: "bg-rose-500/10", border: "border-rose-500/20" },
          { title: "Compliance", value: "86.2%", trend: "+2.4%", isUp: true, icon: ShieldCheck, color: "text-emerald-600 dark:text-emerald-400", bg: "bg-emerald-500/10", border: "border-emerald-500/20" },
          { title: "AI Speed", value: "1.2s", trend: "-0.3s", isUp: true, icon: Activity, color: "text-accent", bg: "bg-accent/10", border: "border-accent/20" },
        ].map((kpi, idx) => (
          <motion.div key={idx} variants={itemVariants} whileHover={{ y: -4, transition: { duration: 0.2 } }} className="relative bg-panel border border-border-color rounded-2xl p-4 md:p-6 overflow-hidden group hover:border-accent/50 transition-colors duration-300 shadow-sm flex flex-col">
            
            {/* KPI Abstract Vector Shape */}
            <svg className="absolute -bottom-6 -right-6 w-32 h-32 text-foreground opacity-[0.03] group-hover:opacity-10 group-hover:scale-110 transition-all duration-500 pointer-events-none" viewBox="0 0 100 100" fill="currentColor">
              <circle cx="50" cy="50" r="40" />
            </svg>

            <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-3 sm:gap-0 mb-4 relative z-10">
              <div className={`p-2.5 md:p-3 rounded-xl border shadow-inner ${kpi.bg} ${kpi.border}`}>
                <kpi.icon className={`w-4 h-4 md:w-5 md:h-5 ${kpi.color}`} />
              </div>
              <div className={`flex items-center gap-1 text-[10px] md:text-[11px] font-bold px-2.5 py-1 rounded-full border shadow-sm ${kpi.isUp ? (kpi.title === "Violations" ? 'text-rose-600 dark:text-rose-400 bg-rose-500/10 border-rose-500/20' : 'text-emerald-600 dark:text-emerald-400 bg-emerald-500/10 border-emerald-500/20') : (kpi.title === "Violations" ? 'text-emerald-600 dark:text-emerald-400 bg-emerald-500/10 border-emerald-500/20' : 'text-rose-600 dark:text-rose-400 bg-rose-500/10 border-rose-500/20')}`}>
                {kpi.isUp ? <ArrowUpRight className="w-3 h-3" /> : <ArrowDownRight className="w-3 h-3" />}
                {kpi.trend}
              </div>
            </div>
            
            <div className="relative z-10 mt-auto">
              <h3 className="text-muted text-[10px] md:text-xs font-bold uppercase tracking-widest">{kpi.title}</h3>
              <p className="text-2xl md:text-3xl font-black mt-1 text-foreground tracking-tight">{kpi.value}</p>
            </div>
          </motion.div>
        ))}
      </motion.div>

      {/* Main Analytics Layout */}
      <motion.div variants={containerVariants} initial="hidden" animate="show" className="grid grid-cols-1 lg:grid-cols-3 gap-6 relative z-10">
        
        {/* Real-time Processing Chart */}
        <motion.div variants={itemVariants} className="bg-panel border border-border-color rounded-2xl shadow-sm hover:border-accent/40 transition-colors flex flex-col relative overflow-hidden lg:col-span-2">
          <div className="h-1.5 w-full bg-hazard opacity-80 absolute top-0 left-0 z-20"></div>
          
          {/* Isometric SVG Graph Watermark */}
          <svg className="absolute inset-0 w-full h-full text-foreground opacity-5 pointer-events-none z-0" preserveAspectRatio="none" viewBox="0 0 100 100" fill="none" stroke="currentColor">
            <path d="M0 80 Q 25 70, 50 80 T 100 80" strokeWidth="0.5" strokeDasharray="2 2" />
            <path d="M0 90 Q 25 80, 50 90 T 100 90" strokeWidth="0.5" strokeDasharray="2 2" />
          </svg>

          <div className="p-5 md:p-6 mt-2 flex flex-col h-full relative z-10">
            <div className="flex justify-between items-center mb-6">
              <div>
                <h3 className="text-base md:text-lg font-bold text-foreground">Processing Volume</h3>
                <p className="text-[10px] md:text-xs text-muted mt-0.5 font-medium">Live extraction metrics across all nodes</p>
              </div>
              <button className="p-2 hover:bg-background rounded-lg transition-colors text-muted hover:text-foreground focus:outline-none">
                <MoreVertical className="w-4 h-4 md:w-5 md:h-5" />
              </button>
            </div>
            <div className="h-[250px] md:h-[320px] w-full min-h-[250px] mt-4 relative right-2 md:right-0">
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={complianceData} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                  <defs>
                    <linearGradient id="colorCompliant" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#10B981" stopOpacity={0.4}/>
                      <stop offset="95%" stopColor="#10B981" stopOpacity={0}/>
                    </linearGradient>
                    <linearGradient id="colorViolations" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#F43F5E" stopOpacity={0.4}/>
                      <stop offset="95%" stopColor="#F43F5E" stopOpacity={0}/>
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" vertical={false} opacity={0.5} />
                  <XAxis dataKey="time" stroke="var(--muted)" fontSize={10} tickLine={false} axisLine={false} fontFamily="var(--font-mono)" fontWeight="bold" />
                  <YAxis stroke="var(--muted)" fontSize={10} tickLine={false} axisLine={false} fontFamily="var(--font-mono)" />
                  <Tooltip 
                    contentStyle={{ backgroundColor: 'var(--panel)', borderColor: 'var(--border)', color: 'var(--foreground)', borderRadius: '12px', fontSize: '12px', fontWeight: 'bold', boxShadow: '0 10px 25px -5px rgba(0, 0, 0, 0.1)' }}
                    itemStyle={{ fontWeight: 600 }}
                    cursor={{ stroke: 'var(--border)', strokeWidth: 1, strokeDasharray: '5 5' }}
                  />
                  <Area type="monotone" dataKey="compliant" stroke="#10B981" strokeWidth={3} fillOpacity={1} fill="url(#colorCompliant)" />
                  <Area type="monotone" dataKey="violations" stroke="#F43F5E" strokeWidth={3} fillOpacity={1} fill="url(#colorViolations)" />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          </div>
        </motion.div>

        {/* Live Scan Feed & Category Donut Stack */}
        <div className="space-y-6 flex flex-col h-full">
          
          {/* Live Feed with Node Telemetry SVG */}
          <motion.div variants={itemVariants} className="bg-panel border border-border-color rounded-2xl shadow-sm flex flex-col flex-1 hover:border-accent/40 transition-colors relative overflow-hidden">
            <div className="h-1.5 w-full bg-hazard opacity-80 absolute top-0 left-0 z-20"></div>
            
            {/* Node Telemetry SVG Watermark */}
            <svg className="absolute -bottom-4 -right-4 w-40 h-40 text-accent opacity-[0.08] pointer-events-none z-0" viewBox="0 0 100 100" fill="none" stroke="currentColor" strokeWidth="1.5">
              <circle cx="20" cy="80" r="4" fill="currentColor"/>
              <circle cx="80" cy="20" r="6" fill="currentColor"/>
              <circle cx="50" cy="50" r="3" fill="currentColor"/>
              <line x1="20" y1="80" x2="50" y2="50" strokeDasharray="2 2"/>
              <line x1="50" y1="50" x2="80" y2="20" strokeDasharray="2 2"/>
              <path d="M80 20 Q 90 50, 50 50" strokeDasharray="1 3" />
            </svg>

            <div className="p-5 mt-2 flex flex-col h-full relative z-10">
              <h3 className="text-[11px] md:text-xs font-bold text-muted uppercase tracking-widest mb-4 flex items-center justify-between">
                Live Feed
                <span className="flex items-center gap-1.5 text-[9px] md:text-[10px] bg-accent/10 border border-accent/20 text-accent px-2 py-1 rounded animate-pulse">
                  <ScanLine className="w-3 h-3" /> Auto-sync
                </span>
              </h3>
              <div className="flex-1 overflow-y-auto pr-1 flex flex-col gap-3 custom-scrollbar">
                {recentScans.map((scan, idx) => (
                  <div key={idx} className="flex justify-between items-center p-3 rounded-xl bg-background border border-border-color hover:border-muted transition-colors cursor-default group backdrop-blur-sm">
                    <div className="flex items-center gap-3">
                      <div className={`w-2 h-2 rounded-full shadow-[0_0_8px_rgba(0,0,0,0.2)] ${scan.status === 'COMPLIANT' ? 'bg-emerald-500' : scan.status === 'NON_COMPLIANT' ? 'bg-rose-500' : 'bg-amber-500'}`}></div>
                      <div>
                        <p className="text-[11px] md:text-xs font-bold text-foreground leading-tight group-hover:text-accent transition-colors">{scan.product}</p>
                        <div className="flex items-center gap-2 mt-1">
                          <span className="text-[9px] md:text-[10px] text-muted font-mono tracking-wider">{scan.id}</span>
                          <span className="text-[9px] text-accent/80 font-mono hidden sm:inline border-l border-border-color pl-2 font-bold">Conf: {scan.conf}</span>
                        </div>
                      </div>
                    </div>
                    <span className="text-[9px] md:text-[10px] text-muted font-mono bg-panel border border-border-color px-2 py-1 rounded-md">{scan.time}</span>
                  </div>
                ))}
              </div>
            </div>
          </motion.div>

          {/* Category Donut Chart with Cyber Reticle */}
          <motion.div variants={itemVariants} className="bg-panel border border-border-color rounded-2xl shadow-sm h-[200px] hover:border-accent/40 transition-colors flex flex-col relative overflow-hidden">
             <div className="h-1.5 w-full bg-hazard opacity-80 absolute top-0 left-0 z-20"></div>
             
             <div className="p-5 mt-2 flex flex-col h-full relative z-10">
               <h3 className="text-[11px] md:text-xs font-bold text-muted uppercase tracking-widest mb-2">Violations by Segment</h3>
               <div className="flex-1 w-full relative flex items-center">
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Tooltip 
                      contentStyle={{ backgroundColor: 'var(--panel)', borderColor: 'var(--border)', borderRadius: '8px', fontSize: '12px', color: 'var(--foreground)', fontWeight: 'bold' }}
                      itemStyle={{ fontWeight: 'bold' }}
                    />
                    <Pie
                      data={categoryData}
                      innerRadius={45}
                      outerRadius={65}
                      paddingAngle={5}
                      dataKey="count"
                      stroke="none"
                    >
                      {categoryData.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={entry.color} className="hover:opacity-80 transition-opacity cursor-pointer outline-none" />
                      ))}
                    </Pie>
                  </PieChart>
                </ResponsiveContainer>
                
                {/* Center KPI with Animated SVG Reticle */}
                <div className="absolute inset-0 flex flex-col items-center justify-center pointer-events-none">
                  {/* Rotating Cyber Reticle */}
                  <motion.svg animate={{ rotate: 360 }} transition={{ repeat: Infinity, duration: 25, ease: "linear" }} className="absolute inset-0 w-full h-full text-muted opacity-20" viewBox="0 0 100 100">
                     <circle cx="50" cy="50" r="35" fill="none" stroke="currentColor" strokeWidth="0.5" strokeDasharray="2 4" />
                     <path d="M50 10 L50 18 M50 90 L50 82 M10 50 L18 50 M90 50 L82 50" stroke="var(--accent)" strokeWidth="1" strokeLinecap="round" />
                  </motion.svg>
                  
                  <span className="text-xl font-black text-foreground relative z-10">1.8k</span>
                  <span className="text-[9px] text-muted font-bold uppercase tracking-wider relative z-10">Total</span>
                </div>
              </div>
            </div>
          </motion.div>

        </div>
      </motion.div>
    </div>
  );
}