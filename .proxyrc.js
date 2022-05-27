const { createProxyMiddleware } = require("http-proxy-middleware");

// 9001 is the default port that tilt starts the application on
const DEFAULT_PROXY_HOST = "http://localhost:9001/";
const proxyHost = process.env.PROXY_HOST || DEFAULT_PROXY_HOST;

// Localhost is running tls by default now
const secure = process.env.PROXY_SECURE === "true";

module.exports = function (app) {
  app.use(
    "/v1",
    createProxyMiddleware({
      target: proxyHost,
      secure,
      changeOrigin: true,
    })
  );
  app.use(
    "/oauth2",
    createProxyMiddleware({
      target: proxyHost,
      secure,
      changeOrigin: true,
    })
  );
};
