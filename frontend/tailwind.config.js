/** @type {import('tailwindcss').Config} */
export default {
  content: [
    './index.html',
    './src/**/*.{vue,js,ts,jsx,tsx}',
  ],
  theme: {
    extend: {
      colors: {
        primary: {
          DEFAULT: '#4FC08D',
          dark: '#3aa876',
        },
        danger: '#e74c3c',
        warning: '#f39c12',
        surface: '#ffffff',
        muted: '#7f8c8d',
      },
      fontFamily: {
        sans: [
          '-apple-system', 'BlinkMacSystemFont', '"Segoe UI"', 'Roboto',
          '"Helvetica Neue"', 'Arial', '"Noto Sans SC"', 'sans-serif',
        ],
      },
    },
  },
  plugins: [],
}
