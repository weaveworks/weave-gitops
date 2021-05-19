import { createBrowserHistory } from "history";
import React from "react";
import ReactDOM from "react-dom";
import App from "./App.tsx";

const history = createBrowserHistory();

// eslint-disable-next-line
ReactDOM.render(<App history={history} />, document.getElementById("app"));
// eslint-disable-next-line
if (module.hot) {
  // eslint-disable-next-line
  module.hot.accept();
}
