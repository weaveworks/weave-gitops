import React from "react";
import ReactDOM from "react-dom";
import App from "./App.tsx";

// eslint-disable-next-line
ReactDOM.render(<App />, document.getElementById("app"));
// eslint-disable-next-line
if (module.hot) {
  // eslint-disable-next-line
  module.hot.accept();
}
