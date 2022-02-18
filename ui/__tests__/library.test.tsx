import { render, screen } from "@testing-library/react";
import "jest-styled-components";
import * as React from "react";
import { BrowserRouter as Router } from "react-router-dom";
import { ThemeProvider } from "styled-components";
import { defaultLinkResolver } from "../contexts/AppContext";
import {
  AppContextProvider,
  ApplicationDetail,
  Applications,
  theme,
} from "../index";
import { GitProvider } from "../lib/api/applications/applications.pb";
import { createMockClient } from "../lib/test-utils";

describe("Example Library App", () => {
  const name = "some app";
  const mockResponses = {
    ListApplications: () => ({ applications: [{ name }] }),
    GetApplication: () => ({ application: { name } }),
    GetReconciledObjects: () => ({ objects: [] }),
    GetChildObjects: () => ({ objects: [] }),
    ListCommits: () => ({ commits: [] }),
    ParseRepoURL: () => ({ provider: GitProvider.GitHub }),
  };

  const wrap = (Component) => (
    <div>
      <ThemeProvider theme={theme}>
        <h3>My custom App!!</h3>
        <Router>
          <AppContextProvider
            linkResolver={defaultLinkResolver}
            applicationsClient={createMockClient(mockResponses)}
          >
            <Component />
          </AppContextProvider>
        </Router>
      </ThemeProvider>
    </div>
  );

  it("renders <Applications />", async () => {
    render(wrap(Applications));
    expect((await screen.findByText(name)).textContent).toEqual(name);
  });
  it("renders <ApplicationDetail />", async () => {
    render(wrap(ApplicationDetail));
    expect(await screen.findAllByText(name)).toBeTruthy();
  });
});
