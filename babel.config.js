module.exports = {
  presets: ['@babel/preset-react'],
  plugins: [
    ['@babel/plugin-transform-runtime', {
      corejs: 3
    }]
  ]
}