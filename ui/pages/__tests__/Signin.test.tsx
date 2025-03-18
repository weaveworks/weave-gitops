import { render, screen, fireEvent } from "@testing-library/react";
import * as React from "react";
import { withContext, withTheme } from "../../lib/test-utils";
import SignIn from "../SignIn";

// from the code
const defaultButtonLabel = "LOGIN WITH OIDC PROVIDER";

const renderSignIn = (featureFlags: Record<string, string>) => {
  render(
    withTheme(
      withContext(<SignIn />, "/sign_in", {
        featureFlags,
      }),
    ),
  );
};

describe("SignIn", () => {
  const location = window.location;

  beforeEach(() => {
    Object.defineProperty(window, "location", {
      writable: true,
      value: { assign: jest.fn() },
    });
  });

  afterEach(() => {
    window.location.assign(location.toString());
  });

  it("should show no buttons or user/password fields with no flags set", async () => {
    renderSignIn({});
    expect(screen.queryByText(defaultButtonLabel)).toBeNull();
    expect(screen.queryByPlaceholderText("Username")).toBeNull();
    expect(screen.queryByPlaceholderText("Password")).toBeNull();
  });

  it("should show user/password fields if CLUSTER_USER_AUTH feature flag is set", async () => {
    renderSignIn({ CLUSTER_USER_AUTH: "true" });
    expect(screen.queryByPlaceholderText("Username")).toBeTruthy();
    expect(screen.queryByPlaceholderText("Password")).toBeTruthy();
  });

  it("should show OIDC button if OIDC_AUTH feature flag is set", async () => {
    renderSignIn({ OIDC_AUTH: "true" });
    expect(screen.queryByText(defaultButtonLabel)).toBeTruthy();
  });

  it("should redirect to the oauth2 endpoint with a relative URL to support running under a subpath", async () => {
    renderSignIn({ OIDC_AUTH: "true" });
    fireEvent.click(screen.queryByText(defaultButtonLabel));

    expect(window.location.href).toEqual(
      "/oauth2?return_url=http%3A%2F%2Flocalhost",
    );
  });

  it("should redirect to the oauth2 endpoint with an absolute URL when baseHref is set", async () => {
    const signInWithBaseTag = (
      <>
        <base href="/wego/" />
        <SignIn />
      </>
    );

    render(
      withTheme(
        withContext(signInWithBaseTag, "/sign_in", {
          featureFlags: { OIDC_AUTH: "true" },
        }),
      ),
    );

    fireEvent.click(screen.queryByText(defaultButtonLabel));
    expect(window.location.href).toEqual(
      "/wego/oauth2?return_url=http%3A%2F%2Flocalhost%2Fwego",
    );
  });

  it("should show both buttons if both flags are set", async () => {
    renderSignIn({
      OIDC_AUTH: "true",
      CLUSTER_USER_AUTH: "true",
    });

    expect(screen.queryByText(defaultButtonLabel)).toBeTruthy();
    expect(screen.queryByPlaceholderText("Username")).toBeTruthy();
    expect(screen.queryByPlaceholderText("Password")).toBeTruthy();
  });

  it("should show the custom button label if feature flag is set", async () => {
    const customLabel = "Login with SSO";
    renderSignIn({
      OIDC_AUTH: "true",
      WEAVE_GITOPS_FEATURE_OIDC_BUTTON_LABEL: customLabel,
    });

    expect(screen.queryByText(defaultButtonLabel)).toBeNull();
    expect(screen.queryByText(customLabel)).toBeTruthy();
  });
});
