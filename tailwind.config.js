/** @type {import('tailwindcss').Config} */
export default {
  content: ["public/**/*.css", "templates/**/*.html"],
  theme: {
    extend: {},
  },
  plugins: [require("daisyui")],
};
