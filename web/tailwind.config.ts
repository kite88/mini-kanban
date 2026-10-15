import type { Config } from "tailwindcss";

const config: Config = {
  darkMode: "class",
  content: ["./app/**/*.{ts,tsx}", "./components/**/*.{ts,tsx}"],
  theme: {
    extend: {
      boxShadow: {
        card: "0 1px 2px rgba(16,24,40,.06), 0 4px 16px rgba(16,24,40,.06)",
      },
      // 全局确认框的入场动画（components/ConfirmProvider.tsx）
      keyframes: {
        "confirm-overlay-in": {
          from: { opacity: "0" },
          to: { opacity: "1" },
        },
        "confirm-dialog-in": {
          from: { opacity: "0", transform: "translateY(8px) scale(.97)" },
          to: { opacity: "1", transform: "translateY(0) scale(1)" },
        },
      },
      animation: {
        "confirm-overlay": "confirm-overlay-in .15s ease-out",
        "confirm-dialog": "confirm-dialog-in .18s cubic-bezier(.2,.9,.3,1)",
      },
    },
  },
  plugins: [],
};

export default config;
