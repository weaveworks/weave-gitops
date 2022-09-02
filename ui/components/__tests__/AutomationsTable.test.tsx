import "jest-styled-components";
import React from "react";
import {withContext, withTheme} from "../../lib/test-utils";
import AutomationsTable from "../../components/AutomationsTable";

import {HelmRelease} from "../../lib/api/core/types.pb";
import {render, screen} from "@testing-library/react";
import Automations from "../../pages/v2/Automations";
import {Automation} from "../../hooks/automations";
import DataTable from "../DataTable";

describe("AutomationsTable", () => {
  it("renders", () => {
    const helmReleaseAsJson = `
    {
          "kind": "KindHelmRelease",
          "releaseName": "",
          "namespace": "flux-system",
          "name": "weave-policy-agent",
          "interval": {
              "hours": "0",
              "minutes": "1",
              "seconds": "0"
          },
          "helmChart": {
              "namespace": "flux-system",
              "name": "flux-system-weave-policy-agent",
              "sourceRef": {
                  "kind": "KindHelmRepository",
                  "name": "weaveworks-charts",
                  "namespace": "flux-system"
              },
              "chart": "weave-policy-agent",
              "version": "0.4.0",
              "interval": null,
              "conditions": [],
              "suspended": false,
              "lastUpdatedAt": "",
              "clusterName": "",
              "apiVersion": "",
              "tenant": ""
          },
          "conditions": [
              {
                  "type": "Ready",
                  "status": "False",
                  "reason": "UpgradeFailed",
                  "message": "Helm upgrade failed: another operation (install/upgrade/rollback) is in progress    Last Helm logs:    checking 1 resources for changes    Replaced 'policies.pac.weave.works' with kind  for kind CustomResourceDefinition    Clearing discovery cache    beginning wait for 1 resources with timeout of 1m0s    preparing upgrade for policy-system-weave-policy-agent",             
                  "timestamp": "2022-09-02T15:07:49Z"
              }
          ],
          "inventory": [],
          "suspended": false,
          "clusterName": "default/dev24",
          "helmChartName": "flux-system/flux-system-weave-policy-agent",
          "lastAppliedRevision": "",
          "lastAttemptedRevision": "0.4.0",
          "apiVersion": "helm.toolkit.fluxcd.io/v2beta1",
          "tenant": ""
    }`;

    const helmRelease:Automation = JSON.parse(helmReleaseAsJson);
    render(
        withTheme(
            withContext(
                <AutomationsTable  automations={[helmRelease] as Automation[]}/>,
                "/applications",
                {}
            )
        )
    );

    const firstRow = screen.getAllByRole("row")[1];
    expect(firstRow.innerHTML).toMatch(/preparing upgrade for policy-system-weave-policy-agent/);
  });
});


