import { screen, waitFor } from "@testing-library/dom";
import { render } from "@testing-library/react";
import "jest-canvas-mock";
import "jest-styled-components";
import React, { act } from "react";
import {
  createCoreMockClient,
  withContext,
  withTheme,
} from "../../lib/test-utils";
import Footer from "../Footer";

jest.mock(
  "../../../package.json",
  () => ({
    version: "x.y.z",
  }),
  { virtual: true },
);

describe("Footer", () => {
  let container;
  beforeEach(() => {
    container = document.createElement("div");
    document.body.appendChild(container);
  });
  afterEach(() => {
    document.body.removeChild(container);
    container = null;
  });

  describe("snapshots", () => {
    it("default", async () => {
      await act(async () => {
        render(
          withTheme(
            withContext(<Footer />, "/", {
              api: createCoreMockClient({
                GetVersion: () => ({
                  semver: "v0.0.1",
                  branch: "mybranch",
                  commit: "123abcd",
                }),
              }),
            }),
          ),
          container,
        );
      });

      await waitFor(() => expect(screen.getByText("Weave GitOps:")));
      const footer = screen.getByRole("footer");
      expect(footer).toMatchSnapshot();
    });
    it("no api version", async () => {
      await act(async () => {
        render(
          withTheme(
            withContext(<Footer />, "/", {
              api: createCoreMockClient({
                GetVersion: () => ({}),
              }),
            }),
          ),
          container,
        );
      });

      await waitFor(() => expect(screen.getByText("Weave GitOps:")));
      const footer = screen.getByRole("footer");
      expect(footer).toMatchSnapshot();
    });
  });
});
