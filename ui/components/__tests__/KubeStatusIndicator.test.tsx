import { render, screen } from "@testing-library/react";
import "jest-styled-components";
import React from "react";
import renderer from "react-test-renderer";
import { withTheme } from "../../lib/test-utils";
import KubeStatusIndicator, {
  createSyntheticConditions,
} from "../KubeStatusIndicator";

describe("KubeStatusIndicator", () => {
  it("renders ready", () => {
    const conditions = [
      {
        type: "Ready",
        status: "True",
        reason: "ReconciliationSucceeded",
        message:
          "Applied revision: main/a3a54ef4a87f8963b14915639f032aa6ec1b8161",
        timestamp: "2022-03-03 17:00:38 +0000 UTC",
      },
    ];
    render(withTheme(<KubeStatusIndicator conditions={conditions} />));

    const msg = screen.getByText(conditions[0].message);
    expect(msg).toBeTruthy();
  });
  it("renders when available", () => {
    const conditions = [
      {
        type: "Available",
        status: "True",
        reason: "ReconciliationSucceeded",
        message:
          "Applied revision: main/a3a54ef4a87f8963b14915639f032aa6ec1b8161",
        timestamp: "2022-03-03 17:00:38 +0000 UTC",
      },
    ];
    render(withTheme(<KubeStatusIndicator conditions={conditions} />));

    const msg = screen.getByText(conditions[0].message);
    expect(msg).toBeTruthy();
  });
  it("renders ready - short", () => {
    const conditions = [
      {
        type: "Ready",
        status: "True",
        reason: "ReconciliationSucceeded",
        message:
          "Applied revision: main/a3a54ef4a87f8963b14915639f032aa6ec1b8161",
        timestamp: "2022-03-03 17:00:38 +0000 UTC",
      },
    ];
    render(withTheme(<KubeStatusIndicator short conditions={conditions} />));

    const ready = screen.getByText("Ready");
    expect(ready).toBeTruthy();
  });
  it("renders an error", () => {
    const conditions = [
      {
        type: "Ready",
        status: "False",
        reason: "BigTrouble",
        message: "There was a problem",
        timestamp: "2022-03-03 17:00:38 +0000 UTC",
      },
    ];
    render(withTheme(<KubeStatusIndicator conditions={conditions} short />));

    const msg = screen.getByText("Not Ready");
    expect(msg).toBeTruthy();
  });
  it("1593 - handles unhealthy", () => {
    const conditions = [
      {
        type: "Ready",
        status: "False",
        reason: "HealthCheckFailed",
        message:
          "Health check failed after 30.004470633s, timeout waiting for: [Deployment/test/backend status: 'Failed']",
        timestamp: "2022-03-03 16:55:29 +0000 UTC",
      },
      {
        type: "Healthy",
        status: "False",
        reason: "HealthCheckFailed",
        message: "HealthCheckFailed",
        timestamp: "2022-03-03 16:55:29 +0000 UTC",
      },
    ];

    render(withTheme(<KubeStatusIndicator conditions={conditions} short />));

    const msg = screen.getByText("Not Ready");
    expect(msg).toBeTruthy();
  });
  it("handles suspended", () => {
    const conditions = [
      {
        type: "Ready",
        status: "True",
        reason: "ReconciliationSucceeded",
        message:
          "Applied revision: main/a3a54ef4a87f8963b14915639f032aa6ec1b8161",
        timestamp: "2022-03-03 17:00:38 +0000 UTC",
      },
    ];

    render(
      withTheme(<KubeStatusIndicator conditions={conditions} suspended short />)
    );
    const msg = screen.getByText("Suspended");
    expect(msg).toBeTruthy();
  });
  it("handles conditions without type: Ready and status: False", () => {
    const notReady = [
      {
        type: "test-condition",
        status: "True",
        reason: "ReconciliationSucceeded",
        message:
          "Applied revision: main/a3a54ef4a87f8963b14915639f032aa6ec1b8161",
        timestamp: "2022-03-03 17:00:38 +0000 UTC",
      },
      {
        type: "other-test-condition",
        status: "False",
        reason: "ReconciliationFailed",
        message:
          "Applied revision: main/a3a54ef4a87f8963b14915639f032aa6ec1b8161",
        timestamp: "2022-03-03 17:00:38 +0000 UTC",
      },
    ];
    render(withTheme(<KubeStatusIndicator conditions={notReady} />));
    expect(screen.getByText(notReady[1].message)).toBeTruthy();
  });
  it("handles conditions without type: Ready and status: True", () => {
    const notReady = [
      {
        type: "test-condition",
        status: "True",
        reason: "ReconciliationSucceeded",
        message:
          "Applied revision: main/a3a54ef4a87f8963b14915639f032aa6ec1b8161",
        timestamp: "2022-03-03 17:00:38 +0000 UTC",
      },
      {
        type: "other-test-condition",
        status: "True",
        reason: "ReconciliationDidItBigTime",
        message:
          "Applied revision: main/a3a54ef4a87f8963b14915639f032aa6ec1b8161",
        timestamp: "2022-03-03 17:00:38 +0000 UTC",
      },
    ];
    render(withTheme(<KubeStatusIndicator conditions={notReady} />));
    expect(screen.getByText(notReady[0].message)).toBeTruthy();
  });
  describe("special objects", () => {
    it("daemonset - not ready", () => {
      const status = {
        currentNumberScheduled: 0,
        desiredNumberScheduled: 2,
        numberMisscheduled: 0,
        numberReady: 0,
        numberUnavailable: 2,
        observedGeneration: 0,
        updatedNumberScheduled: 0,
      };

      const conditions = createSyntheticConditions("Daemonset", status);

      render(withTheme(<KubeStatusIndicator conditions={conditions} />));

      expect(screen.getByText("Not Ready")).toBeTruthy();
    });
    it("daemonset - ready", () => {
      const status = {
        currentNumberScheduled: 0,
        desiredNumberScheduled: 2,
        numberMisscheduled: 0,
        numberReady: 2,
        numberUnavailable: 0,
        observedGeneration: 0,
        updatedNumberScheduled: 0,
      };

      const conditions = createSyntheticConditions("Daemonset", status);

      render(withTheme(<KubeStatusIndicator conditions={conditions} />));

      expect(screen.getByText("Ready")).toBeTruthy();
    });
  });

  describe("snapshots", () => {
    it("renders success", () => {
      const conditions = [
        {
          type: "Ready",
          status: "True",
          reason: "ReconciliationSucceeded",
          message:
            "Applied revision: main/a3a54ef4a87f8963b14915639f032aa6ec1b8161",
          timestamp: "2022-03-03 17:00:38 +0000 UTC",
        },
      ];
      const tree = renderer
        .create(withTheme(<KubeStatusIndicator conditions={conditions} />))
        .toJSON();
      expect(tree).toMatchSnapshot();
    });
    it("renders error", () => {
      const conditions = [
        {
          type: "Ready",
          status: "False",
          reason: "BigTrouble",
          message: "There was a problem",
          timestamp: "2022-03-03 17:00:38 +0000 UTC",
        },
      ];
      const tree = renderer
        .create(
          withTheme(<KubeStatusIndicator conditions={conditions} short />)
        )
        .toJSON();
      expect(tree).toMatchSnapshot();
    });
  });
});
