import "jest-styled-components";
import React from "react";
import renderer from "react-test-renderer";
import { CoreClientContext } from "../../contexts/CoreClientContext";
import {
  createCoreMockClient,
  withContext,
  withTheme,
} from "../../lib/test-utils";
import SyncActions from "../SyncActions";

describe("SyncActions", () => {
  describe("snapshots", () => {
    const mockContext = { api: createCoreMockClient({}), featureFlags: {} };

    it("non-suspended", () => {
      const tree = renderer
        .create(
          withTheme(
            withContext(
              <CoreClientContext.Provider value={mockContext}>
                <SyncActions />
              </CoreClientContext.Provider>,
              "/",
              {}
            )
          )
        )
        .toJSON();
      expect(tree).toMatchSnapshot();
    });
    it("suspended", () => {
      const tree = renderer
        .create(
          withTheme(
            withContext(
              <CoreClientContext.Provider value={mockContext}>
                <SyncActions suspended />
              </CoreClientContext.Provider>,
              "/",
              {}
            )
          )
        )
        .toJSON();
      expect(tree).toMatchSnapshot();
    });
    it("hideDropdown", () => {
      const tree = renderer
        .create(
          withTheme(
            withContext(
              <CoreClientContext.Provider value={mockContext}>
                <SyncActions hideDropdown />
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
