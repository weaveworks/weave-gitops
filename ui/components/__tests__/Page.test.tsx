import "jest-styled-components";
import "jest-canvas-mock";
import React from "react";
import renderer from "react-test-renderer";
import {
  createCoreMockClient,
  withContext,
  withTheme,
} from "../../lib/test-utils";
import Page from "../Page";
import { CoreClientContext } from "../../contexts/CoreClientContext";

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
