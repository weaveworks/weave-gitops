import "jest-styled-components";
import React from "react";
import renderer from "react-test-renderer";
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
      const tree = renderer
        .create(
          withTheme(
            withContext(
              <CoreClientContext.Provider value={mockContext}>
                <SyncControls onSyncClick={() => {}} />
              </CoreClientContext.Provider>,
              "/",
              {}
            )
          )
        )
        .toJSON();
      expect(tree).toMatchSnapshot();
    });
    it("allButtonsDisabled", () => {
      const tree = renderer
        .create(
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
              {}
            )
          )
        )
        .toJSON();
      expect(tree).toMatchSnapshot();
    });
    it("hideSyncOptions", () => {
      const tree = renderer
        .create(
          withTheme(
            withContext(
              <CoreClientContext.Provider value={mockContext}>
                <SyncControls hideSyncOptions onSyncClick={() => {}} />
              </CoreClientContext.Provider>,
              "/",
              {}
            )
          )
        )
        .toJSON();
      expect(tree).toMatchSnapshot();
    });
    it("hideSuspend", () => {
      const tree = renderer
        .create(
          withTheme(
            withContext(
              <CoreClientContext.Provider value={mockContext}>
                <SyncControls hideSuspend onSyncClick={() => {}} />
              </CoreClientContext.Provider>,
              "/",
              {}
            )
          )
        )
        .toJSON();
      expect(tree).toMatchSnapshot();
    });
    it("hasTooltipSuffix", () => {
      const tree = renderer
        .create(
          withTheme(
            withContext(
              <CoreClientContext.Provider value={mockContext}>
                <SyncControls
                  tooltipSuffix="test suffix"
                  onSyncClick={() => {}}
                />
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
