"use client";

import React from "react";
import { 
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
  LineChart, Line
} from "recharts";
import { MapPin, TrendingUp, AlertTriangle, ShieldAlert, Download } from "lucide-react";

// Mock Aggregated SQL Data
const regionalData = [
  { region: "Ahmedabad", violations: 450, compliance: 1200 },
  { region: "Surat", violations: 320, compliance: 980 },
  { region: "Vadodara", violations: 210, compliance: 850 },
  { region: "Rajkot", violations: 180, compliance: 640 },
  { region: "Bhavnagar", violations: 90, compliance: 420 },
];

const trendData = [
  { month: "Apr", rate: 12.5 },
  { month: "May", rate: 14.2 },
  { month: "Jun", rate: 11.1 },
  { month: "Jul", rate: 9.8 },
  { month: "Aug", rate: 8.4 },
  { month: "Sep", rate: 6.2 }, // Current
];

const repeatOffenders = [
  { id: "MFG-1042", name: "Apex Foods Ltd.", total: 14, lastViolation: "2026-08-29", severity: "High", penalty: "Pending" },
  { id: "MFG-0991", name: "DairyAlt Corp", total: 8, lastViolation: "2026-08-15", severity: "Medium", penalty: "Issued" },
  { id: "MFG-2204", name: "Global Cosmetics", total: 6, lastViolation: "2026-08-10", severity: "Medium", penalty: "Paid" },
  { id: "MFG-3112", name: "NatureCo Organics", total: 5, lastViolation: "2026-07-22", severity: "Low", penalty: "Warning" },
];

export default function AnalyticsPage() {
  return (
    <div className="min-h-full bg-background p-6 lg:p-10 text-foreground selection:bg-accent/30 font-sans transition-colors duration-300">
      
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between mb-8 gap-4">
        <div>
          <h1 className="text-2xl md:text-3xl font-black tracking-tight text-foreground flex items-center gap-3">
            <TrendingUp className="w-6 h-6 md:w-8 md:h-8 text-accent" />
            Analytics & Network Intelligence
          </h1>
          <p className="text-xs md:text-sm text-muted mt-1 font-medium">Aggregated legal metrology enforcement data and entity tracking.</p>
        </div>
        <button className="flex items-center justify-center gap-2 bg-panel border border-border-color text-foreground px-4 py-2.5 rounded-xl hover:bg-background transition-colors text-sm font-bold shadow-sm">
          <Download className="w-4 h-4" />
          Export Dataset
        </button>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-6">
        
        {/* Regional Distribution Chart */}
        <div className="lg:col-span-2 bg-panel border border-border-color rounded-2xl shadow-sm relative overflow-hidden flex flex-col">
          <div className="h-1.5 w-full bg-hazard opacity-80 absolute top-0 left-0"></div>
          
          <div className="p-6 mt-2">
            <div className="flex justify-between items-center mb-6">
              <h2 className="text-xs md:text-sm font-bold text-foreground uppercase tracking-wider flex items-center gap-2">
                <MapPin className="w-4 h-4 text-accent" />
                Regional Infraction Density
              </h2>
              <select className="bg-background border border-border-color rounded-lg px-3 py-1.5 text-xs font-bold text-foreground outline-none cursor-pointer shadow-sm">
                <option>Gujarat State</option>
                <option>National View</option>
              </select>
            </div>
            <div className="h-[280px] w-full">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={regionalData} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" vertical={false} opacity={0.5} />
                  <XAxis dataKey="region" stroke="var(--muted)" fontSize={11} tickLine={false} axisLine={false} fontFamily="var(--font-sans)" fontWeight="bold" />
                  <YAxis stroke="var(--muted)" fontSize={11} tickLine={false} axisLine={false} fontFamily="var(--font-mono)" />
                  <Tooltip 
                    contentStyle={{ backgroundColor: 'var(--panel)', borderColor: 'var(--border)', color: 'var(--foreground)', borderRadius: '8px', fontSize: '12px', fontWeight: 'bold' }}
                    cursor={{ fill: 'var(--border)', opacity: 0.4 }}
                  />
                  <Bar dataKey="violations" name="Violations" fill="#F43F5E" radius={[4, 4, 0, 0]} barSize={32} />
                  <Bar dataKey="compliance" name="Compliant" fill="#10B981" radius={[4, 4, 0, 0]} barSize={32} />
                </BarChart>
              </ResponsiveContainer>
            </div>
          </div>
        </div>

        {/* Macro Trend Line */}
        <div className="bg-panel border border-border-color rounded-2xl shadow-sm relative overflow-hidden flex flex-col">
          <div className="h-1.5 w-full bg-hazard opacity-80 absolute top-0 left-0"></div>
          
          <div className="p-6 mt-2 flex flex-col h-full">
            <h2 className="text-xs md:text-sm font-bold text-foreground uppercase tracking-wider mb-6 flex items-center gap-2">
              <ShieldAlert className="w-4 h-4 text-accent" />
              Violation Rate Trajectory
            </h2>
            <div className="h-[220px] w-full mb-4">
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={trendData} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" vertical={false} opacity={0.5} />
                  <XAxis dataKey="month" stroke="var(--muted)" fontSize={11} tickLine={false} axisLine={false} fontFamily="var(--font-sans)" fontWeight="bold" />
                  <YAxis stroke="var(--muted)" fontSize={11} tickLine={false} axisLine={false} tickFormatter={(val) => `${val}%`} fontFamily="var(--font-mono)" />
                  <Tooltip 
                    contentStyle={{ backgroundColor: 'var(--panel)', borderColor: 'var(--border)', color: 'var(--foreground)', borderRadius: '8px', fontSize: '12px', fontWeight: 'bold' }}
                  />
                  <Line type="monotone" dataKey="rate" stroke="var(--accent)" strokeWidth={3} dot={{ r: 4, fill: 'var(--accent)', strokeWidth: 2, stroke: 'var(--panel)' }} activeDot={{ r: 6 }} />
                </LineChart>
              </ResponsiveContainer>
            </div>
            <div className="bg-accent/10 border border-accent/20 p-3 rounded-xl flex items-start gap-3 mt-auto">
               <TrendingUp className="w-5 h-5 text-accent shrink-0" />
               <p className="text-xs text-foreground leading-relaxed font-medium">
                 AI implementation has reduced overall non-compliance by <span className="font-bold text-accent">6.3%</span> across the network over the last 6 months.
               </p>
            </div>
          </div>
        </div>

      </div>

      {/* High-Density Data Table: Repeat Offenders */}
      <div className="bg-panel border border-border-color rounded-2xl shadow-sm flex flex-col overflow-hidden relative">
        <div className="h-1.5 w-full bg-hazard opacity-80 absolute top-0 left-0 z-10"></div>
        
        <div className="px-6 py-5 mt-1.5 border-b border-border-color flex justify-between items-center bg-background/50">
          <div className="flex items-center gap-2">
            <AlertTriangle className="w-5 h-5 text-rose-500" />
            <h2 className="text-xs md:text-sm font-bold text-foreground uppercase tracking-wider">High-Risk Manufacturers (Repeat Offenders)</h2>
          </div>
          <span className="text-[10px] font-mono font-bold bg-rose-500/10 text-rose-600 dark:text-rose-400 border border-rose-500/20 px-2 py-1 rounded shadow-sm">Action Required</span>
        </div>

        <div className="overflow-x-auto custom-scrollbar">
          <table className="w-full text-left border-collapse min-w-[700px]">
            <thead>
              <tr className="border-b border-border-color bg-background/30">
                <th className="px-6 py-4 text-[10px] font-bold text-muted uppercase tracking-widest">Entity ID</th>
                <th className="px-6 py-4 text-[10px] font-bold text-muted uppercase tracking-widest">Manufacturer Name</th>
                <th className="px-6 py-4 text-[10px] font-bold text-muted uppercase tracking-widest text-center">Total Violations</th>
                <th className="px-6 py-4 text-[10px] font-bold text-muted uppercase tracking-widest">Risk Level</th>
                <th className="px-6 py-4 text-[10px] font-bold text-muted uppercase tracking-widest">Last Detection</th>
                <th className="px-6 py-4 text-[10px] font-bold text-muted uppercase tracking-widest">Legal Status</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border-color">
              {repeatOffenders.map((entity, idx) => (
                <tr key={idx} className="hover:bg-background/80 transition-colors group">
                  <td className="px-6 py-4 whitespace-nowrap font-mono text-[11px] text-muted">
                    {entity.id}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap font-bold text-sm text-foreground">
                    {entity.name}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-center">
                    <span className="font-mono text-sm font-bold text-rose-600 dark:text-rose-400">{entity.total}</span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className={`text-[10px] font-bold px-2.5 py-1 rounded-md border uppercase tracking-wider shadow-sm ${
                      entity.severity === 'High' ? 'bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/20' :
                      entity.severity === 'Medium' ? 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/20' :
                      'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20'
                    }`}>
                      {entity.severity}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap font-mono text-[11px] text-muted">
                    {entity.lastViolation}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className="text-[11px] font-bold text-foreground uppercase tracking-wide">
                      {entity.penalty}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

    </div>
  );
}