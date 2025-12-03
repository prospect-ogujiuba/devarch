/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        runtime: {
          php: '#8892BF',
          node: '#68A063',
          python: '#3776AB',
          go: '#00ADD8',
          dotnet: '#512BD4',
        },
        status: {
          active: '#eab308',
          ready: '#22c55e',
          stopped: '#94a3b8',
          unknown: '#cbd5e1',
        }
      },
      fontFamily: {
        mono: ['Monaco', 'Menlo', 'Courier New', 'monospace'],
      }
    },
  },
  plugins: [],
}
