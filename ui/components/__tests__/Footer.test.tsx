import { render, screen } from "@testing-library/react";
import "jest-canvas-mock";
import "jest-styled-components";
import React from "react";
import { act } from "react-dom/test-utils";
import { CoreClientContext } from "../../contexts/CoreClientContext";
import {
  createCoreMockClient,
  withContext,
  withTheme,
} from "../../lib/test-utils";
import Footer from "../Footer";

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
            withContext(
              <CoreClientContext.Provider
                value={{
                  api: createCoreMockClient({
                    GetVersion: () => ({
                      version: {
                        version: "v0.0.1",
                        branch: "mybranch",
                        "git-commit": "123abcd",
                      },
                    }),
                  }),
                }}
              >
                <Footer />
              </CoreClientContext.Provider>,
              "/",
              {}
            )
          ),
          container
        );
      });

      const footer = screen.getByRole("footer");
      expect(footer).toMatchSnapshot();
    });
    it("no api version", async () => {
      await act(async () => {
        render(
          withTheme(
            withContext(
              <CoreClientContext.Provider
                value={{
                  api: createCoreMockClient({
                    GetVersion: () => ({}),
                  }),
                }}
              >
                <Footer />
              </CoreClientContext.Provider>,
              "/",
              {}
            )
          ),
          container
        );
      });

      const footer = screen.getByRole("footer");
      expect(footer).toMatchSnapshot();
    });
  });
});
