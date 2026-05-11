/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Noto Sans Arabic', 'Inter', 'sans-serif'],
      },
      colors: {
        primary: '#1264A3',
        'primary-dark': '#0B4F8A',
        sidebar: '#3F0E40',
        'sidebar-light': '#522653',
        'sidebar-text': '#CFB3CF',
        'sidebar-active': '#1164A3',
      },
    },
  },
  plugins: [],
};
