/** @type {import('tailwindcss').Config} */
export default {
  darkMode: 'class',
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        primary: { DEFAULT: '#4FC08D', dark: '#3aa876' },
        danger: '#e74c3c',
        warning: '#f39c12',
      },
    },
  },
  plugins: [],
}
