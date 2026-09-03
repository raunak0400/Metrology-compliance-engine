"use client";

import React from "react";
import { motion, Variants } from "framer-motion";
import { 
  ChevronLeft, Download, Printer, AlertOctagon, 
  FileWarning, Scale, Fingerprint, Calendar, ShieldAlert,
  Eye, Zap, ScanLine
} from "lucide-react";
import Link from "next/link";

// Prisma-Aligned Mock Data
const reportData = {
  id: "REP-2026-8942",
  inspectionCode: "INS-2026-8922",
  date: "2026-09-01",
  inspector: "Tushar Shah (LM-8492)",
  product: "Almond Milk 1L",
  manufacturer: "DairyAlt Corp",
  status: "NON_COMPLIANT",
  violations: [
    {
      id: "V-001",
      ruleCode: "PCR2011-NETQTY-PRESENCE",
      severity: "HIGH",
      title: "Net Quantity Declaration Missing Standard Unit",
      description: "The extracted text indicates a numerical value ('1000') but fails to declare the standard metric unit ('ml' or 'L') as explicitly mandated by Rule 13 of the Legal Metrology Rules, 2011.",
      expected: "1000 ml / 1 L",
      actual: "Net Vol: 1000",
      conf: 94
    },
    {
      id: "V-002",
      ruleCode: "PCR2011-MRP-PRESENCE",
      severity: "MEDIUM",
      title: "Retail Sale Price (MRP) Format Invalid",
      description: "MRP is present but not in the legally expected 'MRP Rs. <amount>' or 'MRP ₹ <amount>' format inclusive of all taxes.",
      expected: "MRP ₹ 150.00 (Incl. of all taxes)",
      actual: "Price: 150",
      conf: 88
    }
  ]
};

// Vector SVG Evidence Generator based on Rule Type
const VectorEvidence = ({ type }: { type: 'QUANTITY' | 'MRP' }) => {
  if (type === 'QUANTITY') {
    return (
      <svg className="w-14 h-14 md:w-16 md:h-16 text-muted opacity-40 mb-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1" strokeLinecap="round" strokeLinejoin="round">
        <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"></path>
        <polyline points="3.27 6.96 12 12.01 20.73 6.96"></polyline>
        <line x1="12" y1="22.08" x2="12" y2="12"></line>
        <rect x="8" y="10" width="4" height="4" fill="var(--accent)" opacity="0.3" stroke="none" />
      </svg>
    );
  }
  return (
    <svg className="w-14 h-14 md:w-16 md:h-16 text-muted opacity-40 mb-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1">
      <path d="M4 5v14M8 5v14M11 5v14M15 5v14M18 5v14M22 5v14" strokeWidth="1.5" strokeDasharray="4 2"/>
      <rect x="6" y="10" width="12" height="4" fill="var(--accent)" opacity="0.3" stroke="none" />
      <line x1="4" y1="12" x2="22" y2="12" stroke="var(--accent)" strokeWidth="0.5" strokeDasharray="1 1" />
    </svg>
  );
};

// Framer Motion Variants
const containerVariants: Variants = { hidden: { opacity: 0 }, show: { opacity: 1, transition: { staggerChildren: 0.15 } } };
const itemVariants: Variants = { hidden: { opacity: 0, y: 20 }, show: { opacity: 1, y: 0, transition: { type: "spring", stiffness: 300, damping: 24 } } };

export default function PremiumViolationReport() {
  return (
    <div className="min-h-full bg-background text-foreground pb-20 font-sans selection:bg-rose-500/30 overflow-x-hidden transition-colors duration-300 relative">
      
      {/* Global Vector Technical Grid */}
      <div className="absolute inset-0 pointer-events-none opacity-5 dark:opacity-10 z-0" 
           style={{ backgroundImage: 'radial-gradient(var(--foreground) 1px, transparent 1px)', backgroundSize: '24px 24px' }} />

      {/* Action Header */}
      <motion.header initial={{ y: -20, opacity: 0 }} animate={{ y: 0, opacity: 1 }} className="bg-panel/90 backdrop-blur-md border-b border-border-color px-4 md:px-6 py-4 flex flex-col sm:flex-row sm:items-center justify-between sticky top-0 z-30 gap-4 shadow-sm">
        <div className="flex items-center gap-4">
          <Link href="/dashboard" className="p-2 bg-background hover:bg-border-color/50 rounded-xl transition-colors border border-border-color active:scale-95">
            <ChevronLeft className="w-5 h-5 text-muted" />
          </Link>
          <div>
            <h1 className="text-lg md:text-xl font-black text-foreground tracking-tight">Official Non-Compliance Dossier</h1>
            <p className="text-[10px] text-muted font-mono tracking-widest mt-0.5 font-bold">REF: {reportData.id}</p>
          </div>
        </div>
        <div className="flex items-center gap-3 w-full sm:w-auto">
          <button className="flex-1 sm:flex-none flex items-center justify-center gap-2 bg-background border border-border-color text-foreground px-4 py-2.5 rounded-xl hover:bg-border-color/50 transition-colors text-sm font-bold shadow-sm active:scale-95">
            <Printer className="w-4 h-4 text-muted" /> <span className="hidden sm:inline">Print</span>
          </button>
          <button className="flex-1 sm:flex-none flex items-center justify-center gap-2 bg-rose-600 text-white px-5 py-2.5 rounded-xl shadow-md hover:bg-rose-500 transition-all active:scale-95 text-sm font-bold border border-transparent">
            <Download className="w-4 h-4" /> Export Legal PDF
          </button>
        </div>
      </motion.header>

      {/* Report Document Body */}
      <motion.div variants={containerVariants} initial="hidden" animate="show" className="max-w-4xl mx-auto mt-6 md:mt-10 p-4 md:p-0 relative z-10">
        
        <motion.div variants={itemVariants} className="bg-panel border border-border-color rounded-3xl shadow-sm relative overflow-hidden flex flex-col">
          
          {/* Tactical Hazard Overlay for the Document */}
          <div className="h-2 w-full bg-hazard opacity-90 absolute top-0 left-0 z-20"></div>

          {/* Animated Watermark */}
          <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 pointer-events-none select-none opacity-5 flex flex-col items-center justify-center z-0 text-foreground">
            <Scale className="w-64 h-64 mb-4" />
            <div className="text-[120px] font-black tracking-tighter leading-none">VIOLATION</div>
          </div>

          {/* Dossier Header */}
          <div className="relative z-10 p-6 md:p-10 mt-2 border-b border-border-color bg-rose-500/5">
            <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-6 mb-8">
              <div className="flex items-center gap-4">
                <div className="w-14 h-14 rounded-2xl bg-rose-500/10 border border-rose-500/20 flex items-center justify-center shadow-inner">
                  <FileWarning className="w-7 h-7 text-rose-500" />
                </div>
                <div>
                  <h2 className="text-2xl font-black text-foreground tracking-tight">{reportData.product}</h2>
                  <p className="text-sm text-muted mt-1 font-bold">{reportData.manufacturer}</p>
                </div>
              </div>
              <div className="inline-flex items-center gap-2 bg-rose-500/10 text-rose-600 dark:text-rose-400 border border-rose-500/20 px-4 py-2 rounded-xl text-sm font-bold shadow-sm">
                <AlertOctagon className="w-4 h-4 animate-pulse" />
                NON-COMPLIANT
              </div>
            </div>

            {/* Meta Grid */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 md:gap-6 bg-background/80 p-5 rounded-2xl border border-border-color shadow-sm">
              <div>
                <p className="text-[10px] text-muted font-bold uppercase tracking-widest mb-1">Inspection Date</p>
                <div className="flex items-center gap-1.5 text-foreground font-bold text-sm">
                  <Calendar className="w-3.5 h-3.5 text-muted" /> {reportData.date}
                </div>
              </div>
              <div>
                <p className="text-[10px] text-muted font-bold uppercase tracking-widest mb-1">Inspector Code</p>
                <div className="flex items-center gap-1.5 text-foreground font-bold text-sm">
                  <Fingerprint className="w-3.5 h-3.5 text-accent" /> {reportData.inspector}
                </div>
              </div>
              <div>
                <p className="text-[10px] text-muted font-bold uppercase tracking-widest mb-1">Reference Code</p>
                <p className="font-mono text-sm font-bold text-accent">{reportData.inspectionCode}</p>
              </div>
              <div>
                <p className="text-[10px] text-muted font-bold uppercase tracking-widest mb-1">AI Validation</p>
                <div className="flex items-center gap-1.5 text-emerald-600 dark:text-emerald-400 font-bold text-sm">
                  <Zap className="w-3.5 h-3.5" /> Engine Active
                </div>
              </div>
            </div>
          </div>

          {/* Infractions List */}
          <div className="relative z-10 p-6 md:p-10">
            <h3 className="text-sm font-bold text-foreground uppercase tracking-widest mb-6 flex items-center gap-2">
              <ShieldAlert className="w-4 h-4 text-rose-500" /> Detected Legal Infractions
            </h3>
            
            <div className="space-y-6">
              {reportData.violations.map((violation, idx) => (
                <motion.div variants={itemVariants} key={violation.id} className="bg-background border border-border-color hover:border-rose-500/50 rounded-3xl p-5 md:p-6 transition-colors shadow-sm group">
                  
                  {/* Violation Header */}
                  <div className="flex flex-col md:flex-row justify-between items-start mb-4 gap-3">
                    <div>
                      <div className="flex items-center gap-2 mb-1.5">
                        <span className={`text-[9px] font-bold px-2 py-0.5 rounded border uppercase tracking-widest shadow-sm ${
                          violation.severity === 'HIGH' ? 'bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/20' : 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/20'
                        }`}>
                          {violation.severity} SEVERITY
                        </span>
                        <span className="text-[10px] font-mono font-bold bg-panel text-muted px-2 py-0.5 rounded-md border border-border-color shadow-sm">
                          {violation.ruleCode}
                        </span>
                      </div>
                      <h4 className="font-black text-foreground text-base md:text-lg">{violation.title}</h4>
                    </div>
                  </div>
                  
                  <p className="text-sm text-muted leading-relaxed mb-6 font-medium">
                    {violation.description}
                  </p>
                  
                  {/* Vector SVG Evidence Block */}
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4 pt-5 border-t border-border-color">
                    
                    {/* Visual Evidence with Custom Vector Wireframes */}
                    <div className="bg-panel rounded-2xl p-4 border border-border-color shadow-sm flex flex-col">
                      <p className="text-[10px] text-muted uppercase font-bold tracking-widest mb-3 flex items-center gap-1.5">
                        <Eye className="w-3.5 h-3.5" /> Captured Evidence
                      </p>
                      <div className="flex-1 bg-background rounded-xl border border-border-color flex items-center justify-center relative overflow-hidden min-h-[120px] shadow-inner">
                        <div className="absolute inset-0 opacity-10 dark:opacity-20" style={{ backgroundImage: 'radial-gradient(var(--foreground) 1px, transparent 1px)', backgroundSize: '12px 12px' }}></div>
                        
                        <div className="z-10 flex flex-col items-center group-hover:scale-105 transition-transform duration-500">
                          <VectorEvidence type={idx === 0 ? 'QUANTITY' : 'MRP'} />
                          <span className="text-[10px] text-muted font-mono font-bold tracking-widest uppercase">CROP_LABEL_0{idx+1}.SVG</span>
                        </div>

                        {/* AI Bounding Box Reticle */}
                        <div className="absolute inset-4 border-2 border-rose-500/50 border-dashed rounded-lg bg-rose-500/5 flex items-end p-1.5 z-10 pointer-events-none">
                          <div className="absolute top-0 left-0 w-3 h-3 border-t-2 border-l-2 border-rose-500 -m-[2px]" />
                          <div className="absolute bottom-0 right-0 w-3 h-3 border-b-2 border-r-2 border-rose-500 -m-[2px]" />
                          <span className="text-[9px] bg-rose-500 text-white font-bold px-1.5 py-0.5 rounded shadow-sm flex items-center gap-1">
                            <ScanLine className="w-2.5 h-2.5" /> {violation.conf}% CONF
                          </span>
                        </div>
                      </div>
                    </div>

                    {/* Data Evidence */}
                    <div className="bg-panel rounded-2xl p-4 border border-border-color shadow-sm flex flex-col justify-center space-y-4">
                      <div>
                        <p className="text-[10px] text-muted uppercase font-bold tracking-widest mb-1.5">Expected Format (Legal Rule)</p>
                        <div className="bg-emerald-500/10 border border-emerald-500/20 p-2.5 rounded-xl text-emerald-600 dark:text-emerald-400 font-mono text-xs shadow-sm font-bold">
                          {violation.expected}
                        </div>
                      </div>
                      <div>
                        <p className="text-[10px] text-muted uppercase font-bold tracking-widest mb-1.5">Actual Extraction (AI Output)</p>
                        <div className="bg-rose-500/10 border border-rose-500/20 p-2.5 rounded-xl text-rose-600 dark:text-rose-400 font-mono text-xs shadow-sm font-bold">
                          "{violation.actual}"
                        </div>
                      </div>
                    </div>

                  </div>
                </motion.div>
              ))}
            </div>
          </div>

          {/* Digital Sign-off */}
          <div className="p-6 md:p-10 border-t border-border-color bg-background/50 flex flex-col sm:flex-row justify-between items-center gap-4 relative z-10">
            <div>
              <p className="text-xs text-muted mb-1 font-bold">Digitally generated by MetrologyAI</p>
              <p className="text-[10px] font-mono text-muted">HASH: 8f4e2a9b7c1d3e5f6a8b9c0d2e4f6a8b</p>
            </div>
            <div className="text-right">
              <p className="font-sans italic font-black text-2xl text-foreground opacity-80">{reportData.inspector.split(' ')[0]}</p>
              <div className="h-px w-32 bg-border-color my-1 ml-auto"></div>
              <p className="text-[10px] font-bold text-muted uppercase tracking-widest">Authorized Signature</p>
            </div>
          </div>

        </motion.div>
      </motion.div>
    </div>
  );
}