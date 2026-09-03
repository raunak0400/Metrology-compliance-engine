import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./src/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  darkMode: "class",
  theme: {
    extend: {
      colors: {
        background: "var(--background)",
        foreground: "var(--foreground)",
        panel: "var(--panel)",
        border: "var(--border)",
        accent: "var(--accent)",
      },
      backgroundImage: {
        'hazard-stripes': "var(--hazard)",
      }
    },
  },
  plugins: [],
};
export default config;