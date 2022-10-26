import { render, screen, waitFor } from "@testing-library/react";
import * as React from "react";
import {
  createCoreMockClient,
  withContext,
  withTheme,
} from "../../lib/test-utils";
import Automations from "../v2/Automations";
describe("Automations", () => {
  const response = {
    objects: [
      {
        payload:
          '{"apiVersion":"kustomize.toolkit.fluxcd.io/v1beta2","kind":"Kustomization","metadata":{"creationTimestamp":"2022-09-14T16:49:20Z","finalizers":["finalizers.fluxcd.io"],"generation":1,"labels":{"kustomize.toolkit.fluxcd.io/name":"kustomization-testdata","kustomize.toolkit.fluxcd.io/namespace":"flux-system"},"name":"backend","namespace":"default","resourceVersion":"293089","uid":"907093ec-5471-46f9-9953-b5b36f9f7859"},"spec":{"dependsOn":[{"name":"common"}],"force":false,"healthChecks":[{"kind":"Deployment","name":"backend","namespace":"webapp"}],"interval":"5m","path":"./deploy/webapp/backend/","prune":true,"sourceRef":{"kind":"GitRepository","name":"webapp"},"timeout":"2m","validation":"server"},"status":{"conditions":[{"lastTransitionTime":"2022-09-15T12:40:22Z","message":"Applied revision: 6.2.0/79f81383288bf6542fcb5bdd8144b826b33b36e7","reason":"ReconciliationSucceeded","status":"True","type":"Ready"},{"lastTransitionTime":"2022-09-15T12:40:22Z","message":"ReconciliationSucceeded","reason":"ReconciliationSucceeded","status":"True","type":"Healthy"}],"inventory":{"entries":[{"id":"webapp_backend__Service","v":"v1"},{"id":"webapp_backend_apps_Deployment","v":"v1"},{"id":"webapp_backend_autoscaling_HorizontalPodAutoscaler","v":"v2beta2"}]},"lastAppliedRevision":"6.2.0/79f81383288bf6542fcb5bdd8144b826b33b36e7","lastAttemptedRevision":"6.2.0/79f81383288bf6542fcb5bdd8144b826b33b36e7","observedGeneration":1}}\n',
        clusterName: "Default",
        tenant: "",
        uid: "907093ec-5471-46f9-9953-b5b36f9f7859",
        inventory: [],
      },
      {
        payload:
          '{"apiVersion":"helm.toolkit.fluxcd.io/v2beta1","kind":"HelmRelease","metadata":{"annotations":{"reconcile.fluxcd.io/requestedAt":"2022-09-14T14:16:56.304148696Z"},"creationTimestamp":"2022-09-14T14:14:46Z","finalizers":["finalizers.fluxcd.io"],"generation":3,"managedFields":[{"apiVersion":"helm.toolkit.fluxcd.io/v2beta1","fieldsType":"FieldsV1","fieldsV1":{"f:spec":{".":{},"f:chart":{".":{},"f:spec":{".":{},"f:chart":{},"f:reconcileStrategy":{},"f:sourceRef":{".":{},"f:kind":{},"f:name":{}},"f:version":{}}},"f:interval":{},"f:targetNamespace":{}}},"manager":"flux","operation":"Update","time":"2022-09-14T14:14:46Z"},{"apiVersion":"helm.toolkit.fluxcd.io/v2beta1","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:finalizers":{".":{},"v:\\"finalizers.fluxcd.io\\"":{}}}},"manager":"helm-controller","operation":"Update","time":"2022-09-14T14:14:46Z"},{"apiVersion":"helm.toolkit.fluxcd.io/v2beta1","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:reconcile.fluxcd.io/requestedAt":{}}}},"manager":"gitops-server","operation":"Update","time":"2022-09-14T14:17:13Z"},{"apiVersion":"helm.toolkit.fluxcd.io/v2beta1","fieldsType":"FieldsV1","fieldsV1":{"f:status":{"f:conditions":{},"f:helmChart":{},"f:lastAttemptedRevision":{},"f:lastAttemptedValuesChecksum":{},"f:lastHandledReconcileAt":{},"f:lastReleaseRevision":{},"f:observedGeneration":{}}},"manager":"helm-controller","operation":"Update","subresource":"status","time":"2022-09-14T14:17:20Z"}],"name":"ww-gitops","namespace":"flux-system","resourceVersion":"17512","uid":"2dd24865-4ae4-4a0e-9c78-3204a470be9f"},"spec":{"chart":{"spec":{"chart":"weave-gitops","reconcileStrategy":"ChartVersion","sourceRef":{"kind":"HelmRepository","name":"ww-gitops"},"version":"*"}},"interval":"1m0s","targetNamespace":"weave-gitops"},"status":{"conditions":[{"lastTransitionTime":"2022-09-14T14:17:20Z","message":"Reconciliation in progress","reason":"Progressing","status":"Unknown","type":"Ready"}],"helmChart":"flux-system/flux-system-ww-gitops","lastAttemptedRevision":"4.0.0","lastAttemptedValuesChecksum":"da39a3ee5e6b4b0d3255bfef95601890afd80709","lastHandledReconcileAt":"2022-09-14T14:16:56.304148696Z","lastReleaseRevision":1,"observedGeneration":3}}\n',
        clusterName: "Default",
        tenant: "",
        uid: "2dd24865-4ae4-4a0e-9c78-3204a470be9f",
      },
    ],
    errors: [],
  };

  it("should list automations", async () => {
    const client = createCoreMockClient({
      ListObjects: () => {
        return response;
      },
      GetVersion: () => {
        return {
          semver: "",
          commit: "",
          branch: "",
          buildTime: "",
          fluxVersion: "",
          kubeVersion: "",
        };
      },
    });
    render(
      withTheme(withContext(<Automations />, "/automations", { api: client }))
    );
    await waitFor(() =>
      expect(screen.getAllByText("Kustomization").length).toBeTruthy()
    );
    await waitFor(() =>
      expect(screen.getAllByText("HelmRelease").length).toBeTruthy()
    );
  });
  it("should handle undefined response", async () => {
    const client2 = createCoreMockClient({
      ListObjects: () => {
        return { objects: undefined, errors: undefined };
      },
      GetVersion: () => {
        return {
          semver: "",
          commit: "",
          branch: "",
          buildTime: "",
          fluxVersion: "",
          kubeVersion: "",
        };
      },
    });
    render(
      withTheme(withContext(<Automations />, "/automations", { api: client2 }))
    );
    await waitFor(() => expect(screen.getByText("No data")).toBeTruthy());
  });
});
