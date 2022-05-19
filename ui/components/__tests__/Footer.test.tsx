import "jest-styled-components";
import "jest-canvas-mock";
import React from "react";
import { render, screen } from "@testing-library/react";
import { act } from "react-dom/test-utils";
import {
  createCoreMockClient,
  withContext,
  withTheme,
} from "../../lib/test-utils";
import Footer from "../Footer";
import { CoreClientContext } from "../../contexts/CoreClientContext";

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
                value={{ api: createCoreMockClient({}) }}
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
