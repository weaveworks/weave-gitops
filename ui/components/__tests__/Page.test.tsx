import "jest-canvas-mock";
import "jest-styled-components";
import { render } from "@testing-library/react";
import React from "react";
import { CoreClientContext } from "../../contexts/CoreClientContext";
import {
  createCoreMockClient,
  withContext,
  withTheme,
} from "../../lib/test-utils";
import Page from "../Page";

describe("Page", () => {
  describe("snapshots", () => {
    it("default", () => {
      const tree = render(
        withTheme(
          withContext(
            <CoreClientContext.Provider
              value={{ api: createCoreMockClient({}), featureFlags: {} }}
            >
              <Page path={[{ label: "test" }]} />
            </CoreClientContext.Provider>,
            "/",
            {},
          ),
        ),
      );
      expect(tree.container).toMatchSnapshot();
    });
  });
});
