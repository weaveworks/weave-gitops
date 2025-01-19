import "jest-styled-components";
import { render } from "@testing-library/react";
import React from "react";
import { CoreClientContext } from "../../../contexts/CoreClientContext";
import {
  createCoreMockClient,
  withContext,
  withTheme,
} from "../../../lib/test-utils";
import SyncControls from "../SyncControls";

describe("SyncControls", () => {
  describe("snapshots", () => {
    const mockContext = { api: createCoreMockClient({}), featureFlags: {} };

    it("non-suspended", () => {
      const tree = render(
        withTheme(
          withContext(
            <CoreClientContext.Provider value={mockContext}>
              <SyncControls onSyncClick={() => {}} />
            </CoreClientContext.Provider>,
            "/",
            {},
          ),
        ),
      ).asFragment();
      expect(tree).toMatchSnapshot();
    });
    it("allButtonsDisabled", () => {
      const tree = render(
        withTheme(
          withContext(
            <CoreClientContext.Provider value={mockContext}>
              <SyncControls
                syncDisabled
                suspendDisabled
                resumeDisabled
                onSyncClick={() => {}}
              />
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
              <SyncControls hideSyncOptions onSyncClick={() => {}} />
            </CoreClientContext.Provider>,
            "/",
            {},
          ),
        ),
      ).asFragment();
      expect(tree).toMatchSnapshot();
    });
    it("hideSuspend", () => {
      const tree = render(
        withTheme(
          withContext(
            <CoreClientContext.Provider value={mockContext}>
              <SyncControls hideSuspend onSyncClick={() => {}} />
            </CoreClientContext.Provider>,
            "/",
            {},
          ),
        ),
      ).asFragment();
      expect(tree).toMatchSnapshot();
    });
    it("hasTooltipSuffix", () => {
      const tree = render(
        withTheme(
          withContext(
            <CoreClientContext.Provider value={mockContext}>
              <SyncControls
                tooltipSuffix="test suffix"
                onSyncClick={() => {}}
              />
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
