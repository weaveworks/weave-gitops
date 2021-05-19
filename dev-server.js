const express = require("express");
const Bundler = require("parcel-bundler");
const httpProxy = require("http-proxy");

const bundler = new Bundler("ui/index.html", { outDir: "./dist/dev" });
const server = httpProxy.createProxyServer({});

const app = express();

const API_BACKEND = "http://localhost:9001/api/";

const port = 4567;

const proxy = (url) => {
  return (req, res) => {
    server.web(
      req,
      res,
      {
        target: url,
        ws: true,
      },
      (e) => {
        console.error(e);
        res.status(500).json({ msg: e.message });
      }
    );
  };
};

app.use("/api", proxy(API_BACKEND));

app.use(bundler.middleware());

app.listen(port, () => console.log(`Dev server started on :${port}`));
