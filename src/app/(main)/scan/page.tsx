"use client";

import React, { useState, useRef, useCallback } from "react";
import Webcam from "react-webcam";
import { motion, AnimatePresence } from "framer-motion";
import { Camera, UploadCloud, X, ChevronLeft, AlertCircle, ScanLine, Check, ChevronDown } from "lucide-react";
import Link from "next/link";

const CAPTURE_STEPS = [
  { id: 1, title: "Main Full Product", desc: "Ensure the whole product is in frame." },
  { id: 2, title: "Primary Label", desc: "Focus clearly on the front text." },
  { id: 3, title: "Extra Text / Bottom", desc: "Capture nutritional info or bottom text." },
  { id: 4, title: "Cap / Seal Area", desc: "Capture the top cap or tamper seal." },
  { id: 5, title: "Additional Angle", desc: "Any other relevant metrology details." }
];

const VectorPlaceholder = ({ step }: { step: number }) => {
  switch (step) {
    case 1:
      return <svg className="w-6 h-6 md:w-8 md:h-8 text-current" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1" strokeLinecap="round" strokeLinejoin="round"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"></path><polyline points="3.27 6.96 12 12.01 20.73 6.96"></polyline><line x1="12" y1="22.08" x2="12" y2="12"></line></svg>;
    case 2:
      return <svg className="w-6 h-6 md:w-8 md:h-8 text-current" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1"><rect x="4" y="4" width="16" height="16" rx="2" ry="2"></rect><line x1="8" y1="12" x2="16" y2="12" strokeDasharray="2 2"></line><line x1="8" y1="8" x2="16" y2="8"></line><line x1="8" y1="16" x2="12" y2="16"></line></svg>;
    case 3:
      return <svg className="w-6 h-6 md:w-8 md:h-8 text-current" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1"><path d="M4 5v14M8 5v14M11 5v14M15 5v14M18 5v14M22 5v14" strokeWidth="1.5" strokeDasharray="3 3"/></svg>;
    case 4:
      return <svg className="w-6 h-6 md:w-8 md:h-8 text-current" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1"><ellipse cx="12" cy="7" rx="6" ry="3"></ellipse><path d="M6 7v10c0 1.66 2.69 3 6 3s6-1.34 6-3V7"></path><line x1="6" y1="12" x2="18" y2="12" strokeDasharray="2 2"></line></svg>;
    case 5:
      return <svg className="w-6 h-6 md:w-8 md:h-8 text-current" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1"><circle cx="12" cy="12" r="3"></circle><path d="M3 12h4M17 12h4M12 3v4M12 17v4"></path></svg>;
    default:
      return null;
  }
};

const CustomSelect = ({ label, options, value, onChange, placeholder }: any) => {
  const [isOpen, setIsOpen] = useState(false);
  const selected = options.find((o: any) => o.value === value);

  return (
    // DYNAMIC Z-INDEX: Jumps to front only when open to prevent overlap clipping
    <div className={`space-y-2 relative ${isOpen ? 'z-50' : 'z-10'}`}>
      <label className="text-[10px] font-bold text-muted uppercase tracking-widest pl-1">{label}</label>
      <div className="relative">
        <button
          type="button"
          onClick={() => setIsOpen(!isOpen)}
          className={`w-full bg-background border ${isOpen ? 'border-accent' : 'border-border-color hover:border-accent/50'} rounded-xl p-3.5 flex justify-between items-center text-sm transition-all shadow-inner`}
        >
          <span className={selected ? "text-foreground font-bold" : "text-muted font-medium"}>
            {selected ? selected.label : placeholder}
          </span>
          <ChevronDown className={`w-4 h-4 text-muted transition-transform duration-300 ${isOpen ? "rotate-180 text-accent" : ""}`} />
        </button>

        <AnimatePresence>
          {isOpen && (
            <>
              <div className="fixed inset-0 z-30" onClick={() => setIsOpen(false)}></div>
              <motion.div
                initial={{ opacity: 0, y: -10, scale: 0.95 }} 
                animate={{ opacity: 1, y: 0, scale: 1 }} 
                exit={{ opacity: 0, y: -10, scale: 0.95 }}
                transition={{ duration: 0.15 }}
                className="absolute z-50 w-full mt-2 bg-panel/95 backdrop-blur-xl border border-border-color rounded-xl shadow-2xl flex flex-col overflow-hidden"
              >
                <div className="max-h-56 overflow-y-auto custom-scrollbar">
                  {options.map((opt: any) => (
                    <button
                      key={opt.value}
                      type="button"
                      onClick={() => { onChange(opt.value); setIsOpen(false); }}
                      className={`w-full text-left px-4 py-3 text-sm font-medium hover:bg-background transition-colors ${value === opt.value ? "text-accent bg-accent/10 font-bold border-l-2 border-accent" : "text-foreground border-l-2 border-transparent"}`}
                    >
                      {opt.label}
                    </button>
                  ))}
                </div>
              </motion.div>
            </>
          )}
        </AnimatePresence>
      </div>
    </div>
  );
};

export default function ScanUploadPage() {
  const [formData, setFormData] = useState({
    category: "",
    sale_channel: "",
    weight: "",
    unit: "kg",
    is_imported: false,
  });
  
  const [images, setImages] = useState<string[]>([]);
  const [isCameraActive, setIsCameraActive] = useState(false);
  const [isScanning, setIsScanning] = useState(false);
  const [error, setError] = useState("");
  const [isUnitDropdownOpen, setIsUnitDropdownOpen] = useState(false);
  const webcamRef = useRef<Webcam>(null);

  const videoConstraints = { width: { ideal: 1920 }, height: { ideal: 1080 }, facingMode: "environment" };

  const handleInputChange = (name: string, value: any) => {
    setFormData((prev) => ({ ...prev, [name]: value }));
    setError(""); 
  };

  const capturePhoto = useCallback(() => {
    if (webcamRef.current) {
      setIsScanning(true);
      setTimeout(() => {
        const imageSrc = webcamRef.current?.getScreenshot();
        if (imageSrc && images.length < 5) {
          setImages((prev) => [...prev, imageSrc]);
          setIsScanning(false);
          if (images.length === 4) setIsCameraActive(false); 
        }
      }, 600); 
    }
  }, [webcamRef, images]);

  const removeImage = (index: number) => setImages((prev) => prev.filter((_, i) => i !== index));

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (images.length === 0) return setError("The Main Full Product picture is mandatory.");
    if (!formData.category || !formData.sale_channel || !formData.weight) return setError("Please complete all dropdowns and fields.");

    let weightKg = parseFloat(formData.weight) || 0;
    if (formData.unit === "g" || formData.unit === "ml") weightKg = weightKg / 1000;

    const payload: Record<string, any> = {
      product_id: `SKU-${Math.floor(1000 + Math.random() * 9000)}`,
      category: formData.category,
      sale_channel: formData.sale_channel,
      net_weight_kg: Number(weightKg.toFixed(3)),
      is_imported: formData.is_imported,
    };
    
    images.forEach((img, idx) => { payload[`image${idx + 1}`] = img; });
    console.log("JSON PAYLOAD READY:", JSON.stringify(payload, null, 2));
    alert("Scan captured successfully! Payload generated in console.");
  };

  const currentStep = images.length < 5 ? CAPTURE_STEPS[images.length] : null;

  return (
    <div className="min-h-full bg-background text-foreground pb-24 font-sans selection:bg-accent/30 transition-colors duration-300 relative">
      
      <div className="absolute inset-0 pointer-events-none opacity-5 dark:opacity-10 z-0" 
           style={{ backgroundImage: 'radial-gradient(var(--foreground) 1px, transparent 1px)', backgroundSize: '24px 24px' }} />

      <motion.header initial={{ y: -20, opacity: 0 }} animate={{ y: 0, opacity: 1 }} className="bg-panel/90 backdrop-blur-md border-b border-border-color p-4 flex items-center gap-4 sticky top-0 z-20 shadow-sm transition-colors duration-300">
        <Link href="/dashboard" className="p-2 bg-background rounded-xl border border-border-color hover:bg-border-color/50 transition-colors active:scale-95">
          <ChevronLeft className="w-5 h-5 text-muted hover:text-foreground transition-colors" />
        </Link>
        <div>
          <h1 className="text-lg md:text-xl font-black text-foreground leading-tight tracking-tight">
            Compliance Scanner
          </h1>
          <p className="text-[10px] text-accent font-mono font-bold tracking-widest mt-0.5">ML DATA ENTRY</p>
        </div>
      </motion.header>

      <form onSubmit={handleSubmit} className="p-4 md:p-6 max-w-2xl mx-auto space-y-6 mt-2 relative z-10">
        
        <AnimatePresence>
          {error && (
            <motion.div initial={{ opacity: 0, y: -10 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0, scale: 0.95 }} className="bg-rose-500/10 border border-rose-500/30 text-rose-600 dark:text-rose-400 p-4 rounded-xl flex items-center gap-3 text-sm font-bold shadow-sm">
              <AlertCircle className="w-5 h-5 shrink-0" /> <p>{error}</p>
            </motion.div>
          )}
        </AnimatePresence>

        <div className="bg-panel border border-border-color rounded-3xl p-5 shadow-sm relative overflow-hidden transition-colors z-10">
          <div className="absolute top-0 left-0 w-full h-1.5 bg-background border-b border-border-color">
            <motion.div className="h-full bg-accent" initial={{ width: 0 }} animate={{ width: `${(images.length / 5) * 100}%` }} transition={{ duration: 0.3 }} />
          </div>

          <div className="flex justify-between items-center mb-5 mt-2">
            <div>
              <h2 className="text-sm font-black text-foreground tracking-tight">Evidence Imagery</h2>
              {currentStep ? (
                <p className="text-[11px] text-accent font-bold mt-1 uppercase tracking-widest">Next: {currentStep.title}</p>
              ) : (
                <p className="text-[11px] text-emerald-600 dark:text-emerald-400 font-bold mt-1 flex items-center gap-1 uppercase tracking-widest"><Check className="w-3 h-3"/> Capture Complete</p>
              )}
            </div>
            <span className="text-xs font-mono font-bold bg-background border border-border-color px-2 py-1 rounded-lg text-foreground shadow-sm">{images.length}/5</span>
          </div>

          <div className="grid grid-cols-3 sm:grid-cols-5 gap-3">
            {CAPTURE_STEPS.map((step, idx) => {
              const hasImage = images[idx];
              const isCurrent = images.length === idx;
              
              return (
                <div key={step.id} className="flex flex-col gap-1.5">
                  <div className={`relative aspect-square rounded-2xl border-2 flex items-center justify-center overflow-hidden transition-all ${hasImage ? 'border-accent shadow-sm' : isCurrent ? 'border-accent/50 border-dashed bg-accent/5' : 'border-border-color border-dashed bg-background'}`}>
                    {hasImage ? (
                      <>
                        <img src={images[idx]} alt={step.title} className="w-full h-full object-cover" />
                        <button type="button" onClick={() => removeImage(idx)} className="absolute top-1 right-1 bg-black/60 backdrop-blur text-white p-1 rounded-full active:scale-90 transition-transform">
                          <X className="w-3 h-3" />
                        </button>
                      </>
                    ) : (
                      <>
                        <div className={`transition-opacity duration-300 ${isCurrent ? 'text-accent opacity-100' : 'text-muted opacity-30'}`}>
                           <VectorPlaceholder step={idx + 1} />
                        </div>
                        {isCurrent && (
                           <motion.div animate={{ opacity: [0, 0.2, 0] }} transition={{ repeat: Infinity, duration: 1.5 }} className="absolute inset-0 bg-accent" />
                        )}
                      </>
                    )}
                  </div>
                  <span className={`text-[9px] text-center font-bold leading-tight ${hasImage ? 'text-accent' : isCurrent ? 'text-foreground' : 'text-muted'}`}>
                    {step.title}
                  </span>
                </div>
              );
            })}
          </div>

          {images.length < 5 && (
            <button type="button" onClick={() => setIsCameraActive(true)} className="w-full mt-5 bg-accent/10 hover:bg-accent/20 border border-accent/30 text-accent font-black tracking-tight py-3.5 rounded-xl transition-colors flex items-center justify-center gap-2 active:scale-[0.98]">
              <Camera className="w-5 h-5" /> Launch Camera for: {currentStep?.title}
            </button>
          )}
        </div>

        <div className="bg-panel border border-border-color rounded-3xl p-5 md:p-6 space-y-5 shadow-sm relative flex flex-col transition-colors z-20">
          <div className="h-1.5 w-full bg-hazard opacity-80 absolute top-0 left-0 z-10 rounded-t-3xl"></div>
          
          <div className="pt-2 z-20 space-y-5">
            <CustomSelect 
              label="Category *" 
              placeholder="Select a category"
              value={formData.category}
              onChange={(val: string) => handleInputChange("category", val)}
              options={[
                { value: "GENERAL", label: "General Packaged Commodity" },
                { value: "FOOD", label: "Packaged Food" },
                { value: "COSMETICS", label: "Cosmetics & Toiletries" },
                { value: "HOUSEHOLD", label: "Household & Cleaning" },
                { value: "ELECTRONICS", label: "Consumer Electronics" }
              ]}
            />

            <CustomSelect 
              label="Sale Channel *" 
              placeholder="Select distribution channel"
              value={formData.sale_channel}
              onChange={(val: string) => handleInputChange("sale_channel", val)}
              options={[
                { value: "retail", label: "Retail" },
                { value: "wholesale", label: "Wholesale" },
                { value: "e-commerce", label: "E-Commerce" }
              ]}
            />

            {/* DYNAMIC Z-INDEX applied here to prevent Weight Dropdown clipping */}
            <div className={`space-y-2 relative ${isUnitDropdownOpen ? 'z-50' : 'z-10'}`}>
              <label className="text-[10px] font-bold text-muted uppercase tracking-widest pl-1">Net Weight / Volume *</label>
              <div className="flex gap-2">
                <input required type="number" step="any" name="weight" value={formData.weight} onChange={(e) => handleInputChange("weight", e.target.value)} placeholder="0.00" className="flex-1 bg-background border border-border-color hover:border-accent/50 rounded-xl p-3.5 text-foreground font-bold outline-none focus:border-accent transition-all shadow-inner font-mono text-sm" />
                
                <div className="relative w-24 shrink-0 z-50">
                  <button type="button" onClick={() => setIsUnitDropdownOpen(!isUnitDropdownOpen)} className="w-full h-full bg-background border border-border-color hover:border-accent/50 rounded-xl px-3 flex justify-between items-center transition-all shadow-inner">
                    <span className="text-accent font-black text-sm">{formData.unit}</span>
                    <ChevronDown className={`w-3 h-3 text-accent transition-transform ${isUnitDropdownOpen ? "rotate-180" : ""}`} />
                  </button>

                  <AnimatePresence>
                    {isUnitDropdownOpen && (
                      <>
                        <div className="fixed inset-0 z-30" onClick={() => setIsUnitDropdownOpen(false)}></div>
                        <motion.div initial={{ opacity: 0, y: -10 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0, y: -10 }} className="absolute right-0 z-50 w-full mt-2 bg-panel/95 backdrop-blur-xl border border-border-color rounded-xl shadow-md flex flex-col overflow-hidden">
                          <div className="max-h-40 overflow-y-auto custom-scrollbar">
                            {["kg", "g", "l", "ml"].map((u) => (
                              <button key={u} type="button" onClick={() => { handleInputChange("unit", u); setIsUnitDropdownOpen(false); }} className={`w-full text-center px-4 py-2.5 text-sm font-medium hover:bg-background transition-colors ${formData.unit === u ? "text-accent bg-accent/10 font-bold" : "text-foreground"}`}>
                                {u}
                              </button>
                            ))}
                          </div>
                        </motion.div>
                      </>
                    )}
                  </AnimatePresence>
                </div>
              </div>
            </div>

            <div className="space-y-2 pt-2 z-0 relative">
              <div className="flex justify-between items-center pl-1 mb-3">
                <label className="text-[10px] font-bold text-muted uppercase tracking-widest">Is Imported?</label>
              </div>
              <div className="flex bg-background p-1.5 rounded-xl border border-border-color shadow-inner relative">
                <button type="button" onClick={() => handleInputChange("is_imported", true)} className={`flex-1 py-3 text-sm font-black tracking-tight rounded-lg transition-colors z-10 ${formData.is_imported ? 'text-white dark:text-black' : 'text-muted hover:text-foreground'}`}>
                  Yes, Imported
                </button>
                <button type="button" onClick={() => handleInputChange("is_imported", false)} className={`flex-1 py-3 text-sm font-black tracking-tight rounded-lg transition-colors z-10 ${!formData.is_imported ? 'text-white dark:text-black' : 'text-muted hover:text-foreground'}`}>
                  No, Domestic
                </button>
                <motion.div className="absolute top-1.5 bottom-1.5 w-[calc(50%-6px)] bg-accent rounded-lg shadow-sm z-0" animate={{ left: formData.is_imported ? '6px' : 'calc(50%)' }} transition={{ type: "spring", stiffness: 300, damping: 25 }} />
              </div>
            </div>
          </div>
        </div>

        <button type="submit" className="w-full bg-foreground text-background font-black tracking-tight py-4 rounded-xl shadow-md hover:bg-foreground/90 active:scale-[0.98] transition-all flex items-center justify-center gap-2 text-sm md:text-base relative z-10 mb-10 border border-transparent">
          <UploadCloud className="w-5 h-5" /> Generate ML Payload
        </button>
      </form>

      <AnimatePresence>
        {isCameraActive && currentStep && (
          <motion.div initial={{ opacity: 0, scale: 0.95 }} animate={{ opacity: 1, scale: 1 }} exit={{ opacity: 0, scale: 1.05 }} transition={{ duration: 0.2 }} className="fixed inset-0 z-50 bg-black flex flex-col">
            <div className="absolute top-0 inset-x-0 p-5 flex flex-col items-center z-20 bg-gradient-to-b from-black/90 to-transparent pb-10">
              <div className="w-full flex justify-between items-center mb-4">
                <button type="button" onClick={() => setIsCameraActive(false)} className="bg-zinc-900/80 backdrop-blur-md text-white p-2.5 rounded-full border border-zinc-700 active:scale-90 transition-transform">
                  <X className="w-5 h-5" />
                </button>
                <div className="bg-accent px-3 py-1.5 rounded-full border border-accent/50 flex items-center gap-1 shadow-[0_0_15px_rgba(245,158,11,0.5)]">
                  <span className="text-[10px] font-bold text-white uppercase tracking-widest">{images.length + 1} OF 5</span>
                </div>
              </div>
              <h2 className="text-xl font-bold text-white tracking-wide">{currentStep.title}</h2>
              <p className="text-sm text-zinc-300 mt-1">{currentStep.desc}</p>
            </div>

            <div className="flex-1 relative bg-black overflow-hidden flex items-center justify-center">
              <Webcam audio={false} ref={webcamRef} screenshotFormat="image/jpeg" videoConstraints={videoConstraints} className="w-full h-full object-cover" />
              
              <div className="absolute inset-0 pointer-events-none p-8 md:p-24 flex items-center justify-center opacity-70">
                <div className="w-full aspect-[3/4] sm:aspect-square border-2 border-accent/50 rounded-3xl relative">
                  <div className="absolute top-0 left-0 w-8 h-8 border-t-4 border-l-4 border-accent rounded-tl-3xl -m-0.5" />
                  <div className="absolute top-0 right-0 w-8 h-8 border-t-4 border-r-4 border-accent rounded-tr-3xl -m-0.5" />
                  <div className="absolute bottom-0 left-0 w-8 h-8 border-b-4 border-l-4 border-accent rounded-bl-3xl -m-0.5" />
                  <div className="absolute bottom-0 right-0 w-8 h-8 border-b-4 border-r-4 border-accent rounded-br-3xl -m-0.5" />
                  
                  <svg className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-8 h-8 text-accent/50" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1">
                    <line x1="12" y1="2" x2="12" y2="8"></line>
                    <line x1="12" y1="16" x2="12" y2="22"></line>
                    <line x1="2" y1="12" x2="8" y2="12"></line>
                    <line x1="16" y1="12" x2="22" y2="12"></line>
                  </svg>
                </div>
              </div>

              <AnimatePresence>
                {isScanning && (
                  <motion.div initial={{ top: "10%" }} animate={{ top: "90%" }} transition={{ duration: 0.5, ease: "linear" }} className="absolute left-4 right-4 h-1 bg-emerald-400 shadow-[0_0_20px_rgba(52,211,153,1)] z-30 rounded-full pointer-events-none" />
                )}
              </AnimatePresence>

              <AnimatePresence>
                {isScanning && <motion.div initial={{ opacity: 1 }} exit={{ opacity: 0 }} transition={{ duration: 0.3 }} className="absolute inset-0 bg-white z-20 pointer-events-none" />}
              </AnimatePresence>
            </div>

            <div className="h-36 bg-black flex items-center justify-center pb-8 z-20 relative">
              <button type="button" onClick={capturePhoto} disabled={isScanning} className="w-20 h-20 bg-white rounded-full border-4 border-zinc-300 ring-4 ring-accent/50 active:scale-90 transition-transform flex items-center justify-center shadow-[0_0_30px_rgba(255,255,255,0.3)] disabled:opacity-50">
                <div className="w-16 h-16 rounded-full border border-zinc-300 flex items-center justify-center">
                  <ScanLine className="w-6 h-6 text-zinc-900" />
                </div>
              </button>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}