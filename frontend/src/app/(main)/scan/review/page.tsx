"use client";

import React, { useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { 
  ChevronLeft, Save, CheckCircle2, AlertTriangle, 
  Maximize, ZoomIn, Scan, Check
} from "lucide-react";
import Link from "next/link";

// Mock Extracted Data (Aligned with Prisma DeclarationType)
const initialFields = [
  { id: "f1", type: "NET_QUANTITY", label: "Net Quantity", value: "500", unit: "g", conf: 98, box: { top: "40%", left: "15%", width: "25%", height: "8%" } },
  { id: "f2", type: "MRP", label: "Maximum Retail Price", value: "Not Legible", unit: "₹", conf: 34, box: { top: "65%", left: "60%", width: "30%", height: "10%" } },
  { id: "f3", type: "MANUFACTURE_DATE", label: "Mfg. Date", value: "12/08/2026", unit: "", conf: 85, box: { top: "78%", left: "60%", width: "30%", height: "8%" } },
  { id: "f4", type: "MANUFACTURER", label: "Manufacturer Details", value: "NatureCo Organics Ltd, Gujarat", unit: "", conf: 92, box: { top: "88%", left: "15%", width: "70%", height: "10%" } },
];

export default function PremiumScanReview() {
  const [fields, setFields] = useState(initialFields);
  const [activeField, setActiveField] = useState<string | null>(null);
  const [isSaving, setIsSaving] = useState(false);
  const [filter, setFilter] = useState<"ALL" | "LOW_CONF">("ALL");

  const handleUpdate = (id: string, newValue: string) => {
    setFields(prev => prev.map(f => f.id === id ? { ...f, value: newValue, conf: 100 } : f));
  };

  const handleSave = () => {
    setIsSaving(true);
    setTimeout(() => setIsSaving(false), 1500);
  };

  const displayedFields = filter === "ALL" ? fields : fields.filter(f => f.conf < 80);

  // Animation Variants
  const listVariants = { hidden: { opacity: 0 }, show: { opacity: 1, transition: { staggerChildren: 0.1 } } };
  const itemVariants = { hidden: { opacity: 0, x: 20 }, show: { opacity: 1, x: 0, transition: { type: "spring", stiffness: 300, damping: 24 } } };

  return (
    <div className="min-h-full bg-background text-foreground pb-20 font-sans selection:bg-accent/30 overflow-x-hidden transition-colors duration-300">
      
      {/* App Header */}
      <motion.header initial={{ y: -20, opacity: 0 }} animate={{ y: 0, opacity: 1 }} className="bg-panel/90 backdrop-blur-md border-b border-border-color px-4 md:px-6 py-4 flex items-center justify-between sticky top-0 z-30 shadow-sm transition-colors duration-300">
        <div className="flex items-center gap-4">
          <Link href="/dashboard" className="p-2 bg-background hover:bg-border-color/50 rounded-xl transition-colors border border-border-color active:scale-95">
            <ChevronLeft className="w-5 h-5 text-muted hover:text-foreground transition-colors" />
          </Link>
          <div>
            <h1 className="text-lg md:text-xl font-black text-foreground tracking-tight">Review Extraction</h1>
            <p className="text-[10px] text-accent font-mono font-bold tracking-widest uppercase mt-0.5">INS-2026-8942 • AI Conf: 82%</p>
          </div>
        </div>
        <button onClick={handleSave} disabled={isSaving} className="flex items-center gap-2 bg-foreground text-background hover:bg-foreground/90 px-4 md:px-5 py-2 md:py-2.5 rounded-xl shadow-md transition-all active:scale-95 text-sm font-black tracking-tight disabled:opacity-50 border border-transparent">
          {isSaving ? <Scan className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
          <span className="hidden sm:inline">{isSaving ? "Verifying..." : "Finalize & Save"}</span>
        </button>
      </motion.header>

      {/* Main Workspace Workspace */}
      <div className="p-4 md:p-6 lg:p-8 grid grid-cols-1 lg:grid-cols-12 gap-6 lg:gap-8 max-w-[1800px] mx-auto">
        
        {/* LEFT COLUMN: Interactive Evidence Viewer */}
        {/* MOBILE FIX: Static on mobile with fixed height, sticky only on 'lg' desktop screens */}
        <motion.div initial={{ opacity: 0, scale: 0.95 }} animate={{ opacity: 1, scale: 1 }} className="lg:col-span-5 flex flex-col h-[45vh] min-h-[350px] lg:h-[calc(100vh-140px)] lg:sticky top-[100px] mb-4 lg:mb-0 z-20">
          <div className="bg-panel border border-border-color rounded-3xl p-2 flex flex-col h-full shadow-sm overflow-hidden relative transition-colors">
            
            <div className="h-1.5 w-full bg-hazard opacity-80 absolute top-0 left-0 z-10"></div>

            {/* Image Tools Overlay */}
            <div className="absolute top-6 right-6 z-20 flex flex-col gap-2 mt-2">
              <button className="p-2.5 bg-panel/80 backdrop-blur-md border border-border-color rounded-xl text-muted hover:text-foreground hover:border-accent/50 transition-colors shadow-sm">
                <ZoomIn className="w-4 h-4" />
              </button>
              <button className="p-2.5 bg-panel/80 backdrop-blur-md border border-border-color rounded-xl text-muted hover:text-foreground hover:border-accent/50 transition-colors shadow-sm">
                <Maximize className="w-4 h-4" />
              </button>
            </div>

            {/* Evidence Image Container */}
            <div className="flex-1 bg-background rounded-2xl border border-border-color overflow-hidden relative flex items-center justify-center group mt-2 shadow-inner">
              {/* Techy Grid Background */}
              <div className="absolute inset-0 opacity-10 dark:opacity-20" style={{ backgroundImage: 'radial-gradient(var(--foreground) 1px, transparent 1px)', backgroundSize: '24px 24px' }}></div>
              
              {/* TECHNICAL VECTOR SVG PRODUCT WIREFRAME */}
              <div className="absolute inset-8 border-2 border-border-color border-dashed rounded-xl bg-panel/50 flex flex-col items-center justify-center shadow-sm pointer-events-none">
                <svg className="w-24 h-24 md:w-32 md:h-32 text-muted opacity-40 mb-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"></path>
                  <polyline points="3.27 6.96 12 12.01 20.73 6.96"></polyline>
                  <line x1="12" y1="22.08" x2="12" y2="12"></line>
                  {/* Accent Scan Lines on Vector */}
                  <line x1="4" y1="12" x2="20" y2="12" stroke="var(--accent)" strokeWidth="0.5" strokeDasharray="1 1"></line>
                  <line x1="8" y1="16" x2="16" y2="16" stroke="var(--accent)" strokeWidth="0.5" strokeDasharray="1 1"></line>
                </svg>
                <span className="text-muted font-mono font-bold text-xs uppercase tracking-widest">Awaiting Capture</span>
              </div>

              {/* Interactive AI Bounding Boxes */}
              {fields.map((field) => (
                <div
                  key={`box-${field.id}`}
                  onMouseEnter={() => setActiveField(field.id)}
                  onMouseLeave={() => setActiveField(null)}
                  style={field.box}
                  className={`absolute rounded-lg border-2 transition-all duration-300 cursor-pointer flex items-end p-1 ${
                    activeField === field.id 
                      ? 'border-accent bg-accent/20 shadow-md z-20 scale-105' 
                      : field.conf < 80 
                        ? 'border-rose-500/50 bg-rose-500/10 z-10 animate-pulse' 
                        : 'border-emerald-500/30 bg-emerald-500/5 z-10'
                  }`}
                >
                  <span className={`text-[8px] font-bold px-1.5 py-0.5 rounded backdrop-blur-md shadow-sm ${
                    activeField === field.id ? 'bg-accent text-background' : field.conf < 80 ? 'bg-rose-500 text-white' : 'bg-emerald-500 text-white'
                  }`}>
                    {field.label}
                  </span>
                </div>
              ))}
            </div>
            
            <div className="px-4 py-3 flex justify-between items-center text-xs font-mono font-bold text-muted">
              <span>SCAN MODE: OCR_V2</span>
              <span>COORD: ABSOLUTE</span>
            </div>
          </div>
        </motion.div>

        {/* RIGHT COLUMN: Data Extraction Editor */}
        <div className="lg:col-span-7 flex flex-col relative z-10">
          
          {/* Editor Header & Tabs */}
          <div className="flex flex-col sm:flex-row sm:items-center justify-between mb-6 gap-4">
            <div>
              <h2 className="text-lg font-black text-foreground tracking-tight">Extracted Entities</h2>
              <p className="text-xs text-muted mt-1 font-medium">Review and manually correct low-confidence fields.</p>
            </div>
            
            <div className="flex bg-background p-1.5 rounded-xl border border-border-color shadow-inner w-full sm:w-auto">
              <button 
                onClick={() => setFilter("ALL")} 
                className={`flex-1 sm:flex-none px-4 py-2 text-xs font-bold rounded-lg transition-all ${filter === "ALL" ? "bg-panel border border-border-color text-foreground shadow-sm" : "text-muted hover:text-foreground"}`}
              >
                All Fields
              </button>
              <button 
                onClick={() => setFilter("LOW_CONF")} 
                className={`flex-1 sm:flex-none px-4 py-2 text-xs font-bold rounded-lg transition-all flex items-center justify-center gap-2 ${filter === "LOW_CONF" ? "bg-rose-500/10 text-rose-600 dark:text-rose-400 border border-rose-500/20 shadow-sm" : "text-muted hover:text-foreground"}`}
              >
                Needs Review <span className="bg-rose-500 text-white px-1.5 py-0.5 rounded text-[9px] shadow-sm">1</span>
              </button>
            </div>
          </div>

          {/* Animated Field List */}
          <motion.div variants={listVariants} initial="hidden" animate="show" className="space-y-4 pb-10">
            <AnimatePresence mode="popLayout">
              {displayedFields.map((field) => (
                <motion.div 
                  key={field.id}
                  variants={itemVariants}
                  layout
                  onMouseEnter={() => setActiveField(field.id)}
                  onMouseLeave={() => setActiveField(null)}
                  className={`relative p-5 rounded-2xl border transition-all duration-300 ${
                    activeField === field.id 
                      ? 'bg-panel border-accent shadow-sm scale-[1.01] z-10' 
                      : field.conf < 80 
                        ? 'bg-panel/80 border-rose-500/30 hover:border-rose-500/50' 
                        : 'bg-panel/40 border-border-color hover:border-muted'
                  }`}
                >
                  {/* Status Indicator Bar */}
                  <div className={`absolute left-0 top-0 bottom-0 w-1.5 rounded-l-2xl ${
                    field.conf === 100 ? 'bg-accent' : field.conf >= 80 ? 'bg-emerald-500' : 'bg-rose-500'
                  }`}></div>

                  <div className="pl-3">
                    <div className="flex justify-between items-start mb-4">
                      <div>
                        <div className="flex items-center gap-2">
                          <label className="text-sm font-black text-foreground uppercase tracking-wide">{field.label}</label>
                          {field.conf === 100 ? (
                            <CheckCircle2 className="w-4 h-4 text-accent" />
                          ) : field.conf >= 80 ? (
                            <CheckCircle2 className="w-4 h-4 text-emerald-500" />
                          ) : (
                            <AlertTriangle className="w-4 h-4 text-rose-500 animate-pulse" />
                          )}
                        </div>
                        <p className="text-[10px] text-muted font-mono font-bold mt-0.5">KEY: {field.type}</p>
                      </div>
                      
                      <div className={`flex items-center gap-1.5 px-2.5 py-1 rounded-lg border text-[10px] font-bold font-mono tracking-wider shadow-sm ${
                        field.conf === 100 ? 'bg-accent/10 text-accent border-accent/20' :
                        field.conf >= 80 ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20' : 
                        'bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/20'
                      }`}>
                        {field.conf}% CONF
                      </div>
                    </div>

                    <div className="relative group/input">
                      <div className="flex bg-background border border-border-color rounded-xl overflow-hidden focus-within:border-accent focus-within:ring-1 focus-within:ring-accent transition-all shadow-inner">
                        <input
                          type="text"
                          value={field.value}
                          onChange={(e) => handleUpdate(field.id, e.target.value)}
                          className={`flex-1 bg-transparent px-4 py-3.5 text-sm font-bold outline-none transition-colors ${
                            field.conf < 80 ? 'text-rose-600 dark:text-rose-400' : field.conf === 100 ? 'text-accent' : 'text-foreground'
                          }`}
                        />
                        {field.unit && (
                          <div className="flex items-center px-4 bg-panel/50 border-l border-border-color text-muted font-bold text-sm">
                            {field.unit}
                          </div>
                        )}
                      </div>
                      
                      {/* Interactive "Save Edit" button appears on modification */}
                      {field.conf < 100 && (
                        <div className="absolute -right-2 -top-2 opacity-0 group-focus-within/input:opacity-100 transition-opacity">
                          <button className="bg-accent text-background p-1.5 rounded-lg shadow-md border border-transparent hover:bg-accent/90">
                            <Check className="w-3 h-3" />
                          </button>
                        </div>
                      )}
                    </div>
                  </div>
                </motion.div>
              ))}
            </AnimatePresence>
          </motion.div>
          
        </div>
      </div>
    </div>
  );
}