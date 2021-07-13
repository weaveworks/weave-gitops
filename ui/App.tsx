import { MuiThemeProvider } from "@material-ui/core";
import * as React from "react";
import {
  BrowserRouter as Router,
  Redirect,
  Route,
  Switch,
} from "react-router-dom";
import { ToastContainer } from "react-toastify";
import styled, { ThemeProvider } from "styled-components";
import AuthenticatedRoute from "./components/AuthenticatedRoute";
import ErrorBoundary from "./components/ErrorBoundary";
import Flex from "./components/Flex";
import LeftNav from "./components/LeftNav";
import TopToolbar from "./components/TopToolbar";
import AppContextProvider from "./contexts/AppContext";
import { Applications as appsClient } from "./lib/api/applications/applications.pb";
import theme, { GlobalStyle, muiTheme } from "./lib/theme";
import { PageRoute } from "./lib/types";
import ApplicationDetail from "./pages/ApplicationDetail";
import Applications from "./pages/Applications";
import Auth from "./pages/Auth";
import Error from "./pages/Error";
import OAuthCallback from "./pages/OAuthCallback";

const ContentContainer = styled.div`
  width: 100%;
  padding-top: ${(props) => props.theme.spacing.medium};
  padding-bottom: ${(props) => props.theme.spacing.medium};
  padding-right: ${(props) => props.theme.spacing.medium};
`;

const AppContainer = styled.div`
  width: 100%;
  margin: 0 auto;
  padding: 0;
  background-color: ${(props) => props.theme.colors.negativeSpace};
`;

export default function App() {
  return (
    <MuiThemeProvider theme={muiTheme}>
      <ThemeProvider theme={theme}>
        <GlobalStyle />
        <Router>
          <AppContextProvider applicationsClient={appsClient}>
            <AppContainer>
              <ErrorBoundary>
                <TopToolbar />
                <Flex>
                  <LeftNav />
                  <ContentContainer>
                    <Switch>
                      <AuthenticatedRoute
                        exact
                        path={PageRoute.Applications}
                        component={Applications}
                      />
                      <AuthenticatedRoute
                        exact
                        path={PageRoute.ApplicationDetail}
                        component={ApplicationDetail}
                      />
                      <Route exact path={PageRoute.Auth} component={Auth} />
                      <Route
                        exact
                        path={PageRoute.OAuthCallback}
                        component={OAuthCallback}
                      />
                      <Redirect exact from="/" to={PageRoute.Applications} />
                      <Route exact path="*" component={Error} />
                    </Switch>
                  </ContentContainer>
                </Flex>
              </ErrorBoundary>
            </AppContainer>
            <ToastContainer
              position="top-center"
              autoClose={10000}
              newestOnTop={false}
            />
          </AppContextProvider>
        </Router>
      </ThemeProvider>
    </MuiThemeProvider>
  );
}
