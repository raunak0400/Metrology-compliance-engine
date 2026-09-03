"use client";

import React from "react";

interface BrandLogoProps {
  size?: "sm" | "md" | "lg";
  showTagline?: boolean;
}

export default function BrandLogo({ size = "md", showTagline = false }: BrandLogoProps) {
  // Dynamically increased sizing for the image and text to make the logo pop
  const styles = {
    sm: { img: "w-8 h-8", text: "text-xl", tag: "text-[8px]" },
    md: { img: "w-10 h-10 lg:w-12 lg:h-12", text: "text-2xl lg:text-3xl", tag: "text-[10px]" },
    lg: { img: "w-16 h-16 lg:w-20 lg:h-20", text: "text-5xl lg:text-6xl", tag: "text-xs lg:text-sm" }
  };

  const s = styles[size];

  return (
    <div className="flex flex-col">
      <div className="flex items-center gap-2 md:gap-3">
        {/* Vector Image Integration */}
        <img src="/logo.png" alt="Niyamit Logo" className={`${s.img} object-contain drop-shadow-md`} />
        
        <div className="flex flex-col justify-center mt-1">
          <span className={`${s.text} font-black tracking-tight leading-none`}>
            {/* Dark green "नि" applied here, adapting slightly for dark mode legibility */}
            <span className="text-emerald-800 dark:text-emerald-500 font-black">नि</span><span className="text-accent">yamit</span>
          </span>
          {showTagline && (
            <span className={`${s.tag} font-bold text-muted uppercase tracking-widest mt-1.5`}>
              Every Label. Every Rule. Verified.
            </span>
          )}
        </div>
      </div>
    </div>
  );
}