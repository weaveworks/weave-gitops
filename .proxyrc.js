const { createProxyMiddleware } = require("http-proxy-middleware");

const DEFAULT_PROXY_HOST = "http://localhost:9001/";
const proxyHost = process.env.PROXY_HOST || DEFAULT_PROXY_HOST;

// Localhost is running tls by default now
const insecure =
  process.env.PROXY_INSECURE === "true" || proxyHost === DEFAULT_PROXY_HOST;

module.exports = function (app) {
  app.use(
    "/v1",
    createProxyMiddleware({
      target: proxyHost,
      secure: !insecure,
    })
  );
  app.use(
    "/oauth2",
    createProxyMiddleware({
      target: proxyHost,
      secure: !insecure,
    })
  );
};
