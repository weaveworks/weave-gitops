import { makeEventSourceLink } from "../AlertsTable";

describe("AlertsTable", () => {
  describe("makeEventSourceLink", () => {
    it("should convert CrossNamespaceRef objects to urls with filters", () => {
      const allNames = {
        apiVersion: "v1",
        name: "*",
        namespace: "space",
        kind: "GitRepository",
        matchLabels: [],
      };
      const sourceLink = makeEventSourceLink(allNames);
      expect(sourceLink.includes("/sources")).toEqual(true);
      expect(
        sourceLink.includes("type") && sourceLink.includes("GitRepository")
      ).toEqual(true);
      console.log(sourceLink);
      expect(sourceLink.includes("*")).toEqual(false);
      const allNamespaces = {
        apiVersion: "v1",
        name: "goose",
        namespace: "*",
        kind: "HelmRelease",
        matchLabels: [],
      };
      const automationLink = makeEventSourceLink(allNamespaces);
      expect(automationLink.includes("/applications")).toEqual(true);
      expect(
        automationLink.includes("name") && automationLink.includes("goose")
      ).toEqual(true);
      expect(automationLink.includes("namespace")).toEqual(false);
    });
  });
});
