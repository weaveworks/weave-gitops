import { render, screen } from "@testing-library/react";
import "jest-styled-components";
import * as React from "react";
import { unmountComponentAtNode } from "react-dom";
import { act } from "react-dom/test-utils";
import { GitProvider } from "../../lib/api/gitauth/gitauth.pb";
import { getProviderToken } from "../../lib/storage";
import {
  ApplicationOverrides,
  createMockClient,
  withContext,
  withTheme,
} from "../../lib/test-utils";
import GithubDeviceAuthModal from "../GithubDeviceAuthModal";

describe("GithubDeviceAuthModal", () => {
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
        GetGithubDeviceCode: () => ({ userCode: "" }),
        GetGithubAuthStatus: () => ({ accessToken: "" }),
      };
      const repoName = "ssh://example.com/user/repo.git";

      await act(async () => {
        render(
          withTheme(
            withContext(
              <GithubDeviceAuthModal
                repoName={repoName}
                open
                onClose={() => null}
                onSuccess={() => null}
              />,
              "/",
              { applicationsClient: createMockClient(ovr) }
            )
          ),
          container
        );
      });

      const modal = screen.getByRole("presentation");
      expect(modal).toMatchSnapshot();
    });
  });
  it("renders the user code", async () => {
    const userCode = "123-456";
    const ovr: ApplicationOverrides = {
      GetGithubDeviceCode: () => ({ userCode }),
      GetGithubAuthStatus: () => ({ error: "some error" }),
    };
    await act(async () => {
      render(
        withTheme(
          withContext(
            <GithubDeviceAuthModal
              repoName="ssh://example.com/user/repo.git"
              open
              onClose={() => null}
              onSuccess={() => null}
            />,
            "/",
            { applicationsClient: createMockClient(ovr) }
          )
        ),
        container
      );
    });

    const code = screen.getByText(userCode);
    expect(code.innerHTML).toEqual(userCode);
  });
  it("stores a token", async () => {
    const accessToken = "sometoken";
    const ovr = {
      GetGithubDeviceCode: () => ({ userCode: "" }),
      GetGithubAuthStatus: () => ({ accessToken }),
    };
    await act(async () => {
      render(
        withTheme(
          withContext(
            <GithubDeviceAuthModal
              repoName="ssh://example.com/user/repo.git"
              open
              onClose={() => null}
              onSuccess={() => null}
            />,
            "/",
            { applicationsClient: createMockClient(ovr) }
          )
        ),
        container
      );
    });

    const token = getProviderToken(GitProvider.GitHub);
    expect(token).toEqual(accessToken);
  });
});
