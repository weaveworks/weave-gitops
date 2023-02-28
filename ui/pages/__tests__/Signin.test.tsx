// Mock out the bg animation component
jest.mock("../../components/Animations/SignInBackground", () => () => null);

import { render, screen } from "@testing-library/react";
import * as React from "react";
import { withContext, withTheme } from "../../lib/test-utils";
import SignIn from "../SignIn";
describe("SignIn", () => {
  it("should show OIDC button if OIDC_AUTH feature flag is set", async () => {
    render(
      withTheme(
        withContext(<SignIn />, "/sign_in", {
          featureFlags: { OIDC_AUTH: "true" },
        })
      )
    );

    expect(screen.queryByText("LOGIN WITH OIDC PROVIDER")).toBeTruthy();
  });

  it("should show the custom button label if feature flag is set", async () => {
    const customLabel = "Login with SSO";
    render(
      withTheme(
        withContext(<SignIn />, "/sign_in", {
          featureFlags: {
            OIDC_AUTH: "true",
            WEAVE_GITOPS_FEATURE_OIDC_BUTTON_LABEL: customLabel,
          },
        })
      )
    );

    expect(screen.queryByText(customLabel)).toBeTruthy();
  });
});
