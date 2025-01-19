import "jest-styled-components";
import { render } from "@testing-library/react";
import React from "react";
import { CoreClientContext } from "../../../contexts/CoreClientContext";
import {
  createCoreMockClient,
  withContext,
  withTheme,
} from "../../../lib/test-utils";
import SyncActions from "../SyncActions";

describe("SyncActions", () => {
  describe("snapshots", () => {
    const mockContext = { api: createCoreMockClient({}), featureFlags: {} };

    it("non-suspended", () => {
      const tree = render(
        withTheme(
          withContext(
            <CoreClientContext.Provider value={mockContext}>
              <SyncActions />
            </CoreClientContext.Provider>,
            "/",
            {},
          ),
        ),
      ).asFragment();
      expect(tree).toMatchSnapshot();
    });
    it("suspended", () => {
      const tree = render(
        withTheme(
          withContext(
            <CoreClientContext.Provider value={mockContext}>
              <SyncActions suspended />
            </CoreClientContext.Provider>,
            "/",
            {},
          ),
        ),
      ).asFragment();
      expect(tree).toMatchSnapshot();
    });
    it("hideSyncOptions", () => {
      const tree = render(
        withTheme(
          withContext(
            <CoreClientContext.Provider value={mockContext}>
              <SyncActions hideSyncOptions />
            </CoreClientContext.Provider>,
            "/",
            {},
          ),
        ),
      ).asFragment();
      expect(tree).toMatchSnapshot();
    });
  });
});
