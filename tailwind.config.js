/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./templates/**/*.templ",
    "./templates/**/*_templ.go",
  ],
  theme: {
    // Custom breakpoints for Samsung Fold
    screens: {
      'xs': '400px',      // Small phones
      'sm': '640px',      // Fold outer screen (~344px) + margin
      'md': '520px',      // Fold inner screen landscape / small tablets - SHOW 2 COLUMNS
      'lg': '1024px',     // Larger tablets / desktop
      'xl': '1280px',     // Desktop
      '2xl': '1536px',    // Large desktop
    },
    extend: {
      colors: {
        // Primary - Orange (Retro Anime Style)
        primary: {
          50: '#FFF7ED',
          100: '#FFEDD5',
          200: '#FED7AA',
          300: '#FDBA74',
          400: '#FB923C',
          500: '#F97316',
          600: '#EA580C',
          700: '#C2410C',
          800: '#9A3412',
          900: '#7C2D12',
        },
        // Cream Background
        cream: {
          50: '#FFFBF5',
          100: '#FFF5E6',
          200: '#FFECD1',
          300: '#FFE0B2',
        },
        // Retro Dark Colors
        retro: {
          dark: '#3D2314',
          brown: '#8B4513',
          sepia: '#704214',
        },
        // Accent Colors (Anime-inspired)
        accent: {
          pink: '#FF6B9D',
          teal: '#20B2AA',
          gold: '#FFD700',
          coral: '#FF7F7F',
        },
        // Success Green
        success: {
          DEFAULT: '#059669',
          light: '#34D399',
          dark: '#047857',
        },
      },
      fontFamily: {
        'arabic': ['Tajawal', 'sans-serif'],
      },
      boxShadow: {
        'retro': '6px 6px 0px #EA580C',
        'retro-sm': '4px 4px 0px #EA580C',
        'retro-lg': '8px 8px 0px #EA580C',
        'retro-success': '4px 4px 0px #065F46',
      },
      animation: {
        'bounce-slow': 'bounce 2s infinite',
        'pulse-slow': 'pulse 3s infinite',
      },
    },
  },
  plugins: [],
}
