import { render, screen } from "@testing-library/react";
import * as React from "react";
import { GitProvider } from "../../lib/api/applications/applications.pb";
import { createMockClient, withContext } from "../../lib/test-utils";
import { GrpcErrorCodes } from "../../lib/types";
import useAuth, { useIsAuthenticated } from "../gitprovider";

describe("useAuth", () => {
  let container;
  beforeEach(() => {
    container = document.createElement("div");
    document.body.appendChild(container);
  });
  afterEach(() => {
    document.body.removeChild(container);
    container = null;
  });
  it("returns a github device code", async () => {
    const userCode = "123-456";
    const ovr = {
      GetGithubDeviceCode: () => ({
        deviceCode: "acb123def456",
        userCode,
      }),
    };
    const id = "code";

    const TestComponent = () => {
      const { getGithubDeviceCode } = useAuth();
      const [userCode, setUserCode] = React.useState<string>("");

      React.useEffect(() => {
        getGithubDeviceCode().then((res) => setUserCode(res.userCode));
      }, []);

      return <div data-testid={id}>{userCode}</div>;
    };
    render(
      withContext(TestComponent, `/`, {
        applicationsClient: createMockClient(ovr),
      })
    );
    expect((await screen.findByTestId(id)).textContent).toEqual(userCode);
  });
  it("returns the github device auth status", async () => {
    const accessToken = "sometoken123456";
    const ovr = {
      GetGithubAuthStatus: () => ({
        accessToken,
      }),
    };
    const id = "token";

    const TestComponent = () => {
      const { getGithubAuthStatus } = useAuth();
      const [accessToken, setAccessToken] = React.useState<string>("");

      React.useEffect(() => {
        const { promise, cancel } = getGithubAuthStatus({
          deviceCode: "abc123def",
        });
        promise.then((res) => setAccessToken(res.accessToken));

        return cancel;
      }, []);

      return <div data-testid={id}>{accessToken}</div>;
    };
    render(
      withContext(TestComponent, `/`, {
        applicationsClient: createMockClient(ovr),
      })
    );
    expect((await screen.findByTestId(id)).textContent).toEqual(accessToken);
  });
});

describe("useIsAuthenticated", () => {
  it("returns whether a git provider token is valid", async () => {
    const id = "auth";
    const ovr = {
      ValidateProviderToken: () => ({ valid: true }),
    };
    const TestComponent = () => {
      const { isAuthenticated, req } = useIsAuthenticated();

      React.useEffect(() => {
        req(GitProvider.GitHub);
      }, []);

      return (
        <div>
          <div data-testid={id}>{isAuthenticated && "Authenticated!"}</div>
        </div>
      );
    };

    render(
      withContext(TestComponent, "", {
        applicationsClient: createMockClient(ovr),
      })
    );

    expect((await screen.findByTestId(id)).textContent).toEqual(
      "Authenticated!"
    );
  });
  it("should return unathenticated when an error occurs", async () => {
    const id = "auth";
    const ovr = {
      ValidateProviderToken: () => ({ valid: true }),
    };

    const client = createMockClient(ovr);
    client.ValidateProviderToken = () =>
      new Promise((_, reject) =>
        reject({ code: GrpcErrorCodes.Unauthenticated, message: "nah fam" })
      );

    const TestComponent = () => {
      const { isAuthenticated, req } = useIsAuthenticated();

      React.useEffect(() => {
        req(GitProvider.GitHub);
      }, []);

      return (
        <div>
          <div data-testid={id}>
            {/* Note that this is strict equals `false` instead of false, as `null` would mean we haven't tried to validate yet */}
            {isAuthenticated === false && "Unauthorized!"}
          </div>
        </div>
      );
    };

    render(
      withContext(TestComponent, "", {
        applicationsClient: client,
      })
    );

    expect((await screen.findByTestId(id)).textContent).toEqual(
      "Unauthorized!"
    );
  });
});
