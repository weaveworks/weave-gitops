import { render, screen } from "@testing-library/react";
import "jest-styled-components";
import React from "react";
import renderer from "react-test-renderer";
import { withTheme } from "../../lib/test-utils";
import KubeStatusIndicator from "../KubeStatusIndicator";

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
    render(withTheme(<KubeStatusIndicator conditions={conditions} />));

    const msg = screen.getByText(conditions[0].message);
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

    render(withTheme(<KubeStatusIndicator conditions={conditions} />));

    const msg = screen.getByText(conditions[0].message);
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
        suspend: true,
      },
    ];

    render(withTheme(<KubeStatusIndicator conditions={conditions} />));
    const msg = screen.getByText("Suspended");
    expect(msg).toBeTruthy();
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
        .create(withTheme(<KubeStatusIndicator conditions={conditions} />))
        .toJSON();
      expect(tree).toMatchSnapshot();
    });
  });
});
