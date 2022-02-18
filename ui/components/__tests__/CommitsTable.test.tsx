import { render, screen } from "@testing-library/react";
import "jest-styled-components";
import * as React from "react";
import { unmountComponentAtNode } from "react-dom";
import { act } from "react-dom/test-utils";
import { ListCommitsResponse } from "../../lib/api/applications/applications.pb";
import { createMockClient, withContext, withTheme } from "../../lib/test-utils";
import CommitsTable from "../CommitsTable";

describe("CommitsTable", () => {
  let container = null;
  beforeEach(() => {
    // setup a DOM element as a render target
    container = document.createElement("div");
    document.body.appendChild(container);
  });

  afterEach(() => {
    // cleanup on exiting
    unmountComponentAtNode(container);
    container.remove();
    container = null;
  });
  describe("snapshots", () => {
    it("renders", async () => {
      const ovr = {
        ListCommits: (): ListCommitsResponse => ({
          commits: [
            {
              hash: "123abc",
              author: "Example User",
              date: "2021-09-10T23:45:09Z",
            },
          ],
        }),
      };
      const app = { name: "my-app", namespace: "my-ns" };
      let c;
      await act(async () => {
        const { container: div } = render(
          withTheme(
            withContext(
              <CommitsTable app={app} authSuccess={true} provider={"GitHub"} />,
              "/",
              {
                applicationsClient: createMockClient(ovr),
              }
            )
          ),
          container
        );
        c = div;
      });
      expect(c.firstChild).toMatchSnapshot();
    });
  });
  it("shows commits", async () => {
    const hash = "123abc";
    const commits = [
      { hash, author: "Example User", date: "2021-09-10T23:45:09Z" },
    ];
    const ovr = {
      ListCommits: (): ListCommitsResponse => ({
        commits,
      }),
    };
    const app = { name: "my-app", namespace: "my-ns" };

    await act(async () => {
      render(
        withTheme(
          withContext(
            <CommitsTable app={app} authSuccess={true} provider={"GitHub"} />,
            "/",
            {
              applicationsClient: createMockClient(ovr),
            }
          )
        ),
        container
      );
    });
    const row = await screen.findByText(hash);
    expect(row.innerHTML).toEqual(hash);
  });
});
