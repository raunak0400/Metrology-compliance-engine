"use client";

import React from "react";
import { motion, Variants } from "framer-motion";
import { 
  History, Calendar, Filter, Search, 
  CheckCircle2, AlertOctagon, User, Box, ArrowRight, Cpu, FileText
} from "lucide-react";

// Database-Aligned Audit Logs (AuditAction Enum + Inspection Status)
const timelineData = [
  {
    date: "Today, Sep 1, 2026",
    events: [
      { id: "LOG-001", entityId: "INS-2026-8942", time: "14:32:05", action: "FINALIZE", outcome: "NON_COMPLIANT", actor: "Tushar S. (INSPECTOR)", details: "Inspection finalized. Report generated for Rule PCR2011-NETQTY-PRESENCE." },
      { id: "LOG-002", entityId: "INS-2026-8942", time: "14:30:12", action: "REVIEW", outcome: "MANUAL_REVIEW", actor: "M. Kumar (REVIEWER)", details: "Manual override approved for MRP visibility." },
      { id: "LOG-003", entityId: "ML-JOB-812", time: "14:28:00", action: "PROCESS", outcome: "SUCCEEDED", actor: "System Engine", details: "OCR and Field Extraction completed successfully. Bounding boxes mapped." },
    ]
  },
  {
    date: "Yesterday, Aug 31, 2026",
    events: [
      { id: "LOG-004", entityId: "INS-2026-8940", time: "16:45:10", action: "FINALIZE", outcome: "COMPLIANT", actor: "Tushar S. (INSPECTOR)", details: "All 5 metadata fields matched LM Rules 2011." },
      { id: "LOG-005", entityId: "RULE-SYNC", time: "09:20:00", action: "UPDATE", outcome: "INFO", actor: "System Admin", details: "Rule requirements for PCR2011-COUNTRY-OF-ORIGIN updated." },
    ]
  }
];

const containerVariants: Variants = { hidden: { opacity: 0 }, show: { opacity: 1, transition: { staggerChildren: 0.15 } } };
const itemVariants: Variants = { hidden: { opacity: 0, x: -20 }, show: { opacity: 1, x: 0, transition: { type: "spring", stiffness: 300, damping: 24 } } };

export default function HistoryTimelinePage() {
  const getIcon = (action: string, outcome: string) => {
    if (action === "PROCESS") return <Cpu className="w-4 h-4 md:w-5 md:h-5 text-accent" />;
    if (action === "UPDATE") return <FileText className="w-4 h-4 md:w-5 md:h-5 text-amber-500" />;
    if (outcome === "NON_COMPLIANT") return <AlertOctagon className="w-4 h-4 md:w-5 md:h-5 text-rose-500" />;
    if (outcome === "COMPLIANT") return <CheckCircle2 className="w-4 h-4 md:w-5 md:h-5 text-emerald-500" />;
    return <Box className="w-4 h-4 md:w-5 md:h-5 text-accent" />;
  };

  return (
    <div className="min-h-full bg-background p-4 md:p-6 lg:p-10 font-sans selection:bg-accent/30 overflow-x-hidden transition-colors duration-300 text-foreground">
      
      {/* Animated Header */}
      <motion.div initial={{ opacity: 0, y: -20 }} animate={{ opacity: 1, y: 0 }} className="flex flex-col md:flex-row md:items-center justify-between mb-8 md:mb-10 gap-4">
        <div>
          <h1 className="text-2xl md:text-3xl font-black text-foreground flex items-center gap-3 tracking-tight">
            <History className="w-6 h-6 md:w-8 md:h-8 text-accent" />
            Audit & History Logs
          </h1>
          <p className="text-xs md:text-sm text-muted mt-1 font-medium">Immutable audit trail of all system actions and rule enforcement.</p>
        </div>
        
        <div className="flex flex-col sm:flex-row items-center gap-3 w-full md:w-auto">
          <div className="relative group w-full sm:w-auto">
            <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-muted group-focus-within:text-accent transition-colors" />
            <input placeholder="Search Entity ID or Actor..." className="bg-panel border border-border-color rounded-xl pl-10 pr-4 py-3 md:py-2.5 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent transition-all w-full md:w-64 text-foreground placeholder:text-muted shadow-sm" />
          </div>
          <div className="flex gap-2 w-full sm:w-auto">
            <button className="flex-1 sm:flex-none flex items-center justify-center gap-2 bg-panel border border-border-color px-4 py-3 md:py-2.5 rounded-xl text-xs md:text-sm hover:bg-background transition-colors text-foreground font-bold shadow-sm">
              <Calendar className="w-4 h-4 text-muted" /> Date
            </button>
            <button className="flex-1 sm:flex-none flex items-center justify-center gap-2 bg-panel border border-border-color px-4 py-3 md:py-2.5 rounded-xl text-xs md:text-sm hover:bg-background transition-colors text-foreground font-bold shadow-sm">
              <Filter className="w-4 h-4 text-muted" /> Filters
            </button>
          </div>
        </div>
      </motion.div>

      {/* Vertical Timeline */}
      <div className="max-w-4xl mx-auto pb-10">
        <motion.div variants={containerVariants} initial="hidden" animate="show">
          {timelineData.map((day, dayIdx) => (
            <div key={dayIdx} className="mb-10">
              
              <motion.div variants={itemVariants} className="flex items-center gap-4 mb-6">
                <span className="text-[10px] md:text-xs font-bold uppercase tracking-widest text-muted bg-panel px-3 py-1.5 rounded-full border border-border-color shadow-sm">
                  {day.date}
                </span>
                <div className="h-px flex-1 bg-gradient-to-r from-border-color to-transparent"></div>
              </motion.div>

              <div className="space-y-6 relative before:absolute before:inset-0 before:ml-5 before:-translate-x-px md:before:mx-auto md:before:translate-x-0 before:h-full before:w-[2px] before:bg-gradient-to-b before:from-border-color before:to-transparent z-0">
                {day.events.map((ev, evIdx) => (
                  <motion.div variants={itemVariants} key={evIdx} className="relative flex items-center justify-between md:justify-normal md:odd:flex-row-reverse group">
                    
                    {/* Glowing Timeline Dot */}
                    <div className="flex items-center justify-center w-10 h-10 md:w-12 md:h-12 rounded-full border-4 border-background bg-panel shadow-sm shrink-0 md:order-1 md:group-odd:-translate-x-1/2 md:group-even:translate-x-1/2 z-10 transition-transform group-hover:scale-110 group-hover:border-accent/50">
                      {getIcon(ev.action, ev.outcome)}
                    </div>

                    {/* Event Card */}
                    <div className="w-[calc(100%-3.5rem)] md:w-[calc(50%-2.5rem)] bg-panel border border-border-color rounded-2xl shadow-sm hover:border-accent/40 transition-all hover:-translate-y-1 hover:shadow-md cursor-default relative overflow-hidden flex flex-col">
                      <div className="h-1 w-full bg-hazard opacity-80 absolute top-0 left-0"></div>
                      <div className="p-4 md:p-5 mt-1">
                        <div className="flex items-center justify-between mb-3">
                          <span className="font-mono text-[9px] md:text-[10px] text-muted tracking-wider font-bold">
                            {ev.time}
                          </span>
                          <div className="flex items-center gap-1.5">
                            <span className="text-[9px] font-bold px-2 py-0.5 rounded bg-background border border-border-color uppercase tracking-widest text-muted">
                              {ev.action}
                            </span>
                          </div>
                        </div>
                        
                        <h3 className="text-xs md:text-sm font-bold text-foreground mb-1.5 tracking-wide">{ev.entityId}</h3>
                        <p className="text-[11px] md:text-xs text-muted leading-relaxed mb-4 font-medium">
                          {ev.details}
                        </p>
                        
                        <div className="flex items-center justify-between pt-3 border-t border-border-color">
                          <div className="flex items-center gap-1.5 text-[10px] md:text-xs text-muted font-bold tracking-wide">
                            <User className="w-3 h-3 md:w-3.5 md:h-3.5 text-accent" />
                            {ev.actor}
                          </div>
                          <button className="text-[10px] md:text-xs font-bold text-accent hover:text-accent/80 flex items-center gap-1 transition-colors">
                            Details <ArrowRight className="w-3 h-3" />
                          </button>
                        </div>
                      </div>
                    </div>

                  </motion.div>
                ))}
              </div>
            </div>
          ))}
        </motion.div>
      </div>
    </div>
  );
}