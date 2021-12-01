import { render, screen } from "@testing-library/react";
import * as React from "react";
import { createMockClient, withContext } from "../../lib/test-utils";
import useAuth from "../auth";

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
