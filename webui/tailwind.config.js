/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts}'],
  theme: {
    extend: {
      colors: {
        bg: '#080c0e',
        panel: '#0f1519',
        'panel-elevated': '#141c22',
        edge: '#1e2d36',
        'edge-subtle': '#162028',
        accent: '#c8ff3e',
        warm: '#ffa63e',
        danger: '#ff4757',
        ok: '#3effa8'
      },
      boxShadow: {
        panel: '0 12px 40px rgba(0,0,0,0.5)',
        'panel-hover': '0 16px 48px rgba(0,0,0,0.6)'
      },
      fontSize: {
        '2xs': '0.65rem'
      }
    }
  },
  plugins: []
}
