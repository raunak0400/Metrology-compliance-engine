"use client";

import React, { useState, useMemo } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { Search, ArrowUpDown, Filter, X, Box, Calendar, AlertOctagon, CheckCircle2, Clock } from "lucide-react";

// 1. Aligned strictly with Prisma Schema Enums
type InspectionOutcome = "COMPLIANT" | "NON_COMPLIANT" | "MANUAL_REVIEW" | "PENDING";

// 2. Aligned with Prisma Inspection Model
type Inspection = {
  code: string; 
  productName: string;
  category: string;
  createdAt: string; 
  outcome: InspectionOutcome; 
  inspectorName: string;
  notes?: string;
};

// 3. Seeded with Prisma-aligned dummy data
const defaultData: Inspection[] = [
  { code: "INS-2026-8921", productName: "Organic Honey 500g", category: "Packaged Food", createdAt: "2026-09-01", outcome: "COMPLIANT", inspectorName: "Tushar S.", notes: "All mandatory declarations present per PCR2011." },
  { code: "INS-2026-8922", productName: "Almond Milk 1L", category: "Packaged Food", createdAt: "2026-09-01", outcome: "NON_COMPLIANT", inspectorName: "Tushar S.", notes: "PCR2011-NETQTY-PRESENCE: Missing standard unit." },
  { code: "INS-2026-8923", productName: "Face Wash 150ml", category: "Cosmetics & Toiletries", createdAt: "2026-08-31", outcome: "MANUAL_REVIEW", inspectorName: "Rahul K.", notes: "PCR2011-MRP-PRESENCE: OCR confidence below threshold." },
  { code: "INS-2026-8924", productName: "LED Bulb 9W", category: "Consumer Electronics", createdAt: "2026-08-31", outcome: "COMPLIANT", inspectorName: "Tushar S.", notes: "Manufacturer and MRP clearly visible." },
  { code: "INS-2026-8925", productName: "Detergent Powder 2kg", category: "Household & Cleaning", createdAt: "2026-08-30", outcome: "NON_COMPLIANT", inspectorName: "Ananya P.", notes: "PCR2011-CONSUMER-CARE: Contact details missing." },
];

export default function RepositoryPage() {
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedScan, setSelectedScan] = useState<Inspection | null>(null);

  const filteredData = useMemo(() => {
    return defaultData.filter((item) => 
      item.productName.toLowerCase().includes(searchQuery.toLowerCase()) ||
      item.code.toLowerCase().includes(searchQuery.toLowerCase())
    );
  }, [searchQuery]);

  const getOutcomeStyles = (outcome: InspectionOutcome) => {
    switch(outcome) {
      case "COMPLIANT": return { bg: "bg-emerald-500/10", border: "border-emerald-500/20", text: "text-emerald-600 dark:text-emerald-400", icon: CheckCircle2 };
      case "NON_COMPLIANT": return { bg: "bg-rose-500/10", border: "border-rose-500/20", text: "text-rose-600 dark:text-rose-400", icon: AlertOctagon };
      case "MANUAL_REVIEW": return { bg: "bg-amber-500/10", border: "border-amber-500/20", text: "text-amber-600 dark:text-amber-400", icon: Clock };
      default: return { bg: "bg-gray-500/10", border: "border-gray-500/20", text: "text-gray-600 dark:text-gray-400", icon: Box };
    }
  };

  return (
    <div className="min-h-full bg-background p-4 md:p-6 lg:p-8 text-foreground selection:bg-accent/30 relative font-sans transition-colors duration-300">
      
      {/* Mobile-Optimized Header */}
      <motion.div initial={{ opacity: 0, y: -10 }} animate={{ opacity: 1, y: 0 }} className="flex flex-col md:flex-row md:items-center justify-between mb-6 md:mb-8 gap-4">
        <div>
          <h1 className="text-2xl md:text-3xl font-black tracking-tight text-foreground">Inspection Archive</h1>
          <p className="text-xs md:text-sm text-muted mt-1 font-medium">Search and filter all historical compliance records.</p>
        </div>
        
        <div className="flex flex-col sm:flex-row items-center gap-3 w-full md:w-auto">
          <div className="relative group w-full sm:w-auto">
            <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-muted group-focus-within:text-accent transition-colors" />
            <input
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search products or INS codes..."
              className="bg-panel border border-border-color rounded-xl pl-10 pr-4 py-3 md:py-2.5 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent transition-all w-full md:w-72 text-foreground placeholder:text-muted shadow-sm"
            />
          </div>
          <button className="flex items-center justify-center gap-2 bg-panel border border-border-color px-4 py-3 md:py-2.5 rounded-xl text-sm hover:bg-background transition-colors shadow-sm w-full sm:w-auto text-foreground font-bold active:scale-95">
            <Filter className="w-4 h-4 text-muted" />
            <span>Filters</span>
          </button>
        </div>
      </motion.div>

      {/* MOBILE VIEW: Animated Cards (Hidden on Desktop) */}
      <div className="md:hidden flex flex-col gap-3 pb-20">
        <AnimatePresence>
          {filteredData.map((row, idx) => {
            const style = getOutcomeStyles(row.outcome);
            return (
              <motion.div 
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
                exit={{ opacity: 0, scale: 0.95 }}
                transition={{ delay: idx * 0.05 }}
                key={row.code}
                onClick={() => setSelectedScan(row)}
                className="bg-panel border border-border-color rounded-2xl p-4 shadow-sm active:scale-[0.98] transition-transform relative overflow-hidden"
              >
                <div className="h-1 w-full bg-hazard opacity-80 absolute top-0 left-0"></div>
                <div className="flex justify-between items-start mb-3 mt-1">
                  <div>
                    <h3 className="font-bold text-foreground text-sm leading-tight">{row.productName}</h3>
                    <p className="text-[10px] text-muted font-mono mt-0.5">{row.code}</p>
                  </div>
                  <div className={`p-1.5 rounded-lg border shadow-sm ${style.bg} ${style.border}`}>
                    <style.icon className={`w-4 h-4 ${style.text}`} />
                  </div>
                </div>
                <div className="flex items-center justify-between mt-4 pt-3 border-t border-border-color">
                  <span className="text-[10px] text-muted font-bold bg-background border border-border-color px-2 py-1 rounded shadow-sm">{row.category}</span>
                  <span className="text-[10px] text-muted font-mono font-bold">{row.createdAt}</span>
                </div>
              </motion.div>
            );
          })}
        </AnimatePresence>
      </div>

      {/* DESKTOP VIEW: Premium Data Table (Hidden on Mobile) */}
      <div className="hidden md:block bg-panel border border-border-color rounded-2xl overflow-hidden shadow-sm relative">
        <div className="h-1.5 w-full bg-hazard opacity-80 absolute top-0 left-0 z-10"></div>
        <table className="w-full text-left border-collapse mt-1.5 relative z-0">
          <thead>
            <tr className="border-b border-border-color bg-background/50">
              {["Inspection Code", "Commodity", "Category", "Date", "Outcome", "Inspector"].map((header, idx) => (
                <th key={idx} className="px-6 py-4 text-[11px] font-bold text-muted uppercase tracking-widest whitespace-nowrap">
                  <div className="flex items-center gap-2 cursor-pointer hover:text-foreground transition-colors">
                    {header} {header && <ArrowUpDown className="w-3 h-3 opacity-40" />}
                  </div>
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-border-color">
            {filteredData.map((row) => {
              const style = getOutcomeStyles(row.outcome);
              return (
                <tr key={row.code} onClick={() => setSelectedScan(row)} className="hover:bg-background/80 transition-colors group cursor-pointer">
                  <td className="px-6 py-4 whitespace-nowrap font-mono font-bold text-[11px] text-muted tracking-wider group-hover:text-accent transition-colors">{row.code}</td>
                  <td className="px-6 py-4 whitespace-nowrap font-bold text-sm text-foreground">{row.productName}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-xs text-muted font-medium">{row.category}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-xs text-muted font-mono font-bold">{row.createdAt}</td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className={`px-2.5 py-1 rounded-md text-[10px] font-bold border uppercase tracking-wider flex items-center w-max gap-1.5 shadow-sm ${style.bg} ${style.border} ${style.text}`}>
                      <style.icon className="w-3 h-3" /> {row.outcome.replace("_", " ")}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-xs text-muted font-bold">{row.inspectorName}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      {/* Responsive Detail Modal */}
      <AnimatePresence>
        {selectedScan && (
          <motion.div 
            initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}
            className="fixed inset-0 z-50 flex items-end md:items-center justify-center p-0 md:p-4"
          >
            <div className="absolute inset-0 bg-background/80 backdrop-blur-sm" onClick={() => setSelectedScan(null)}></div>
            <motion.div 
              initial={{ y: "100%" }} animate={{ y: 0 }} exit={{ y: "100%" }}
              transition={{ type: "spring", damping: 25, stiffness: 300 }}
              className="relative bg-panel border border-border-color rounded-t-3xl md:rounded-2xl shadow-2xl w-full max-w-2xl overflow-hidden flex flex-col max-h-[85vh]"
            >
              <div className="h-1.5 w-full bg-hazard opacity-90 absolute top-0 left-0 z-10"></div>
              <div className="w-12 h-1.5 bg-border-color rounded-full mx-auto mt-4 md:hidden relative z-20"></div>
              
              <div className="flex justify-between items-center p-5 mt-1.5 border-b border-border-color bg-background/50 relative z-20">
                <div className="flex items-center gap-3">
                  <Box className="w-5 h-5 text-accent" />
                  <h2 className="text-base md:text-lg font-black text-foreground">Inspection Dossier</h2>
                </div>
                <button onClick={() => setSelectedScan(null)} className="p-1.5 bg-background border border-border-color hover:bg-border-color/50 rounded-full text-muted transition-colors active:scale-95">
                  <X className="w-5 h-5" />
                </button>
              </div>
              
              <div className="p-6 overflow-y-auto custom-scrollbar relative z-20">
                <div className="mb-6">
                  <p className="text-[10px] text-muted uppercase tracking-widest font-bold mb-1">Commodity Identity</p>
                  <p className="text-xl font-black text-foreground">{selectedScan.productName}</p>
                  <p className="text-sm font-mono font-bold text-accent mt-1">{selectedScan.code}</p>
                </div>

                <div className="p-4 bg-background rounded-xl border border-border-color shadow-sm mb-6">
                  <p className="text-[10px] text-muted uppercase tracking-widest font-bold mb-3">Rule Engine Outcome</p>
                  <div className="flex items-start gap-3">
                    {React.createElement(getOutcomeStyles(selectedScan.outcome).icon, { className: `w-5 h-5 mt-0.5 ${getOutcomeStyles(selectedScan.outcome).text}` })}
                    <div>
                      <span className={`text-xs font-bold tracking-wider uppercase ${getOutcomeStyles(selectedScan.outcome).text}`}>
                        {selectedScan.outcome.replace("_", " ")}
                      </span>
                      <p className="text-sm text-foreground font-medium mt-1.5 leading-relaxed">{selectedScan.notes}</p>
                    </div>
                  </div>
                </div>

                <button className="w-full py-3.5 bg-foreground hover:bg-foreground/90 text-background rounded-xl font-bold text-sm transition-all shadow-md active:scale-95 border border-transparent">
                  View Full Evidence Package
                </button>
              </div>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}