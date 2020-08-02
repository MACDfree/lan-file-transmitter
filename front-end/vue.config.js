module.exports = {
  publicPath: process.env.NODE_ENV === 'production' ? './' : '/',
  devServer: {
    proxy: {
      '/rest': {
        target: 'http://localhost:8089'
      }
    }
  }
}