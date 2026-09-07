/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{svelte,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        cream: "#F8F5EE",
        bone: "#EDE8DC",
        char: "#0C0C0E",
        "char-card": "#161619",
        "char-card-hover": "#202024",
        ember: "#C1121F",
        "ember-light": "#E63946",
        gold: "#C5A059",
        "gold-light": "#D4AF37",
        ash: "#8E8E93",
      },
      fontFamily: {
        display: ["'Satoshi'", "'Plus Jakarta Sans'", "sans-serif"],
        editorial: ["'Cabinet Grotesk'", "'Satoshi'", "sans-serif"],
        flourish: ["'Playfair Display'", "'Cormorant Garamond'", "serif"],
        script: ["'Pinyon Script'", "'Italianno'", "'Alex Brush'", "cursive"],
        serif: ["'Italiana'", "'Cormorant Garamond'", "Georgia", "serif"],
        sans: ["'Satoshi'", "'Plus Jakarta Sans'", "sans-serif"],
        mono: ["'JetBrains Mono'", "monospace"],
      },
      boxShadow: {
        lux: "0 4px 20px -2px rgba(0, 0, 0, 0.25)",
        gold: "0 0 25px -5px rgba(197, 160, 89, 0.3)",
        ember: "0 0 25px -5px rgba(193, 18, 31, 0.4)",
      },
      borderRadius: {
        card: "42px",
        "card-lg": "48px",
      },
    },
  },
  plugins: [],
}
