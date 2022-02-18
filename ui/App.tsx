import { MuiThemeProvider } from "@material-ui/core";
import * as React from "react";
import { QueryClient, QueryClientProvider } from "react-query";
import {
  BrowserRouter as Router,
  Redirect,
  Route,
  Switch,
} from "react-router-dom";
import { ToastContainer } from "react-toastify";
import "react-toastify/dist/ReactToastify.css";
import { ThemeProvider } from "styled-components";
import ErrorBoundary from "./components/ErrorBoundary";
import Layout from "./components/Layout";
import AppContextProvider from "./contexts/AppContext";
import { Core } from "./lib/api/core/core.pb";
import Fonts from "./lib/fonts";
import theme, { GlobalStyle, muiTheme } from "./lib/theme";
import { V2Routes } from "./lib/types";
import Error from "./pages/Error";
import Automations from "./pages/v2/Automations";
import FluxRuntime from "./pages/v2/FluxRuntime";
import Sources from "./pages/v2/Sources";

const queryClient = new QueryClient();

export default function App() {
  return (
    <MuiThemeProvider theme={muiTheme}>
      <ThemeProvider theme={theme}>
        <QueryClientProvider client={queryClient}>
          <Fonts />
          <GlobalStyle />
          <Router>
            <AppContextProvider renderFooter coreClient={Core}>
              <Layout>
                <ErrorBoundary>
                  <Switch>
                    <Route
                      exact
                      path={V2Routes.Automations}
                      component={Automations}
                    />
                    <Route exact path={V2Routes.Sources} component={Sources} />
                    <Route
                      exact
                      path={V2Routes.FluxRuntime}
                      component={FluxRuntime}
                    />
                    <Redirect exact from="/" to={V2Routes.Automations} />
                    <Route exact path="*" component={Error} />
                  </Switch>
                </ErrorBoundary>
                <ToastContainer
                  position="top-center"
                  autoClose={5000}
                  newestOnTop={false}
                />
              </Layout>
            </AppContextProvider>
          </Router>
        </QueryClientProvider>
      </ThemeProvider>
    </MuiThemeProvider>
  );
}
