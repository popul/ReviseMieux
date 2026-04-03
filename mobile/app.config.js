// Dynamic Expo config — reads API_URL from environment for CI/E2E
// Falls back to app.json values for local development
const appJson = require('./app.json');

module.exports = ({ config }) => {
  return {
    ...config,
    extra: {
      ...config.extra,
      // In CI: API_URL=http://10.0.2.2:8080/api/v1 (Android emulator → host)
      // Local: uses app.json default (192.168.x.x)
      apiUrl: process.env.API_URL || config.extra?.apiUrl || 'http://192.168.1.19:8080/api/v1',
    },
  };
};
