/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        coral: {
          DEFAULT: '#E85D4C',
          light: '#FF7A6B',
          dark: '#C94A3B',
        },
        teal: {
          DEFAULT: '#1A4D4D',
          light: '#2A6B6B',
        },
        gold: {
          DEFAULT: '#F5C542',
          light: '#FFD966',
        },
        cream: {
          DEFAULT: '#FBF8F3',
          dark: '#F5F0E8',
        },
        ink: {
          DEFAULT: '#1A1A1A',
          light: '#4A4A4A',
          muted: '#8A8A8A',
        },
        success: {
          DEFAULT: '#34A853',
          light: '#E6F4EA',
        },
      },
      fontFamily: {
        display: ['Fraunces', 'Georgia', 'serif'],
        body: ['DM Sans', '-apple-system', 'sans-serif'],
      },
      borderRadius: {
        'sm': '8px',
        'md': '12px',
        'lg': '20px',
        'full': '100px',
      },
    },
  },
  plugins: [],
}
