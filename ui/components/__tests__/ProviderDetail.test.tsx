import { randomUUID } from "crypto";
import { render, waitFor } from "@testing-library/react";
import * as React from "react";

import { Kind } from "../../lib/api/core/types.pb";
import { Provider } from "../../lib/objects";
import {
  createCoreMockClient,
  withContext,
  withTheme,
} from "../../lib/test-utils";
import ProviderDetail from "../ProviderDetail";

describe("ProviderDetail", () => {
  beforeEach(() => {
    jest.spyOn(console, "error").mockImplementation();
  });

  const responseObject = {
    payload:
      '{"apiVersion":"notification.toolkit.fluxcd.io/v1beta3","kind":"Provider","metadata":{"creationTimestamp":"2025-02-14T16:27:55Z","generation":1,"labels":{"kustomize.toolkit.fluxcd.io/name":"flux-system","kustomize.toolkit.fluxcd.io/namespace":"flux-system"},"managedFields":[{"apiVersion":"notification.toolkit.fluxcd.io/v1beta3","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:labels":{"f:kustomize.toolkit.fluxcd.io/name":{},"f:kustomize.toolkit.fluxcd.io/namespace":{}}},"f:spec":{"f:secretRef":{"f:name":{}},"f:type":{}}},"manager":"kustomize-controller","operation":"Apply","time":"2025-02-14T17:19:51Z"}],"name":"discord-bot","namespace":"flux-system","resourceVersion":"189450782","uid":"6ef27ec3-4e7c-45b9-906f-4777de2c0c0f"},"spec":{"secretRef":{"name":"discord-webhook"},"type":"discord"}}\n',
    clusterName: "Default",
    tenant: "",
    uid: randomUUID(),
    inventory: [],
    info: "",
    health: null,
  };

  const provider = new Provider(responseObject);

  const objects = [
    {
      payload:
        '{"apiVersion":"notification.toolkit.fluxcd.io/v1beta3","kind":"Alert","metadata":{"creationTimestamp":"2025-02-14T16:27:55Z","generation":1,"labels":{"kustomize.toolkit.fluxcd.io/name":"flux-system","kustomize.toolkit.fluxcd.io/namespace":"flux-system"},"managedFields":[{"apiVersion":"notification.toolkit.fluxcd.io/v1beta3","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:labels":{"f:kustomize.toolkit.fluxcd.io/name":{},"f:kustomize.toolkit.fluxcd.io/namespace":{}}},"f:spec":{"f:eventMetadata":{"f:cluster":{},"f:env":{},"f:region":{}},"f:eventSeverity":{},"f:eventSources":{},"f:inclusionList":{},"f:providerRef":{"f:name":{}},"f:summary":{}}},"manager":"kustomize-controller","operation":"Apply","time":"2025-02-14T17:19:51Z"}],"name":"discord-error","namespace":"flux-system","resourceVersion":"189450780","uid":"7edf429a-5550-47d4-affc-f512fda2d057"},"spec":{"eventSeverity":"info","eventSources":[{"kind":"GitRepository","name":"*"},{"kind":"Kustomization","name":"*"}],"inclusionList":[".*failed.*"],"providerRef":{"name":"discord-bot"},"summary":"An error occurred during reconciliation"}}\n',
      clusterName: "Default",
      tenant: "",
      uid: randomUUID(),
      inventory: [],
      info: "",
      health: null,
    },
    {
      payload:
        '{"apiVersion":"notification.toolkit.fluxcd.io/v1beta3","kind":"Alert","metadata":{"creationTimestamp":"2025-02-14T16:27:55Z","generation":1,"labels":{"kustomize.toolkit.fluxcd.io/name":"flux-system","kustomize.toolkit.fluxcd.io/namespace":"flux-system"},"managedFields":[{"apiVersion":"notification.toolkit.fluxcd.io/v1beta3","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:labels":{"f:kustomize.toolkit.fluxcd.io/name":{},"f:kustomize.toolkit.fluxcd.io/namespace":{}}},"f:spec":{"f:eventMetadata":{"f:cluster":{},"f:env":{},"f:region":{}},"f:eventSeverity":{},"f:eventSources":{},"f:inclusionList":{},"f:providerRef":{"f:name":{}},"f:summary":{}}},"manager":"kustomize-controller","operation":"Apply","time":"2025-02-14T17:19:51Z"}],"name":"discord-info","namespace":"flux-system","resourceVersion":"189450781","uid":"c82c6df2-ee90-4258-97f8-bd1a8861478c"},"spec":{"eventSeverity":"info","eventSources":[{"kind":"Kustomization","name":"*"}],"inclusionList":[".*passed.*"],"providerRef":{"name":"discord-bot"},"summary":"A reconciliation has been completed successfully"}}\n',
      clusterName: "Default",
      tenant: "",
      uid: randomUUID(),
      inventory: [],
      info: "",
      health: null,
    },
  ];

  it("renders", async () => {
    const client = createCoreMockClient({
      ListObjects: ({ kind }) => {
        const fullResponse = { objects: [], errors: [] };
        if (kind === Kind.Alert) fullResponse.objects = objects;
        return fullResponse;
      },
    });
    render(
      withTheme(
        withContext(<ProviderDetail provider={provider} />, "/alerts", {
          api: client,
        }),
      ),
    );
    await waitFor(() => {
      const rows = document.querySelectorAll("tbody tr");
      expect(rows.length).toEqual(objects.length);
    });
  });
});
