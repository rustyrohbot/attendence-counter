/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./internal/templates/layout.html",
    "./internal/templates/attendance/*.html",
  ],
  theme: {
    extend: {
      maxWidth: {
        '2000': '2000px',
      }
    },
  },
  plugins: [],
}