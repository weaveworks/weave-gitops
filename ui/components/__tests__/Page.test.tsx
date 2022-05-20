import "jest-canvas-mock";
import "jest-styled-components";
import React from "react";
import renderer from "react-test-renderer";
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
      const tree = renderer
        .create(
          withTheme(
            withContext(
              <CoreClientContext.Provider
                value={{ api: createCoreMockClient({}) }}
              >
                <Page />
              </CoreClientContext.Provider>,
              "/",
              {}
            )
          )
        )
        .toJSON();
      expect(tree).toMatchSnapshot();
    });
  });
});
