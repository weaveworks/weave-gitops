import { addSearchedNamespaces } from "../automations";

describe("addSearchedNamespaces", () => {
  it("calculates searched namespaces for an unique cluster", () => {
    const final = { errors: [], result: [], searchedNamespaces: [] };
    const response = {
      errors: [],
      objects: [],
      searchedNamespaces: {
        Default: {
          namespaces: [
            "default",
            "flux-system",
            "kube-node-lease",
            "kube-public",
          ],
        },
      },
    };
    const finalWithSearchedNamespaces = addSearchedNamespaces(final, response);
    expect(
      finalWithSearchedNamespaces.searchedNamespaces[0]["Default"]
    ).toMatchObject([
      "default",
      "flux-system",
      "kube-node-lease",
      "kube-public",
    ]);
  });

  it("calculates searched namespaces for several clusters", () => {
    const final = { errors: [], result: [], searchedNamespaces: [] };
    const response = {
      errors: [],
      objects: [],
      searchedNamespaces: {
        Default: {
          namespaces: [
            "default",
            "flux-system",
            "kube-node-lease",
            "kube-public",
          ],
        },
        TestCluster: {
          namespaces: [
            "default",
            "flux-system",
            "kube-node-lease",
            "kube-public",
          ],
        },
      },
    };
    const finalWithSearchedNamespaces = addSearchedNamespaces(final, response);
    expect(finalWithSearchedNamespaces.searchedNamespaces).toMatchObject([
      {
        Default: ["default", "flux-system", "kube-node-lease", "kube-public"],
      },
      {
        TestCluster: [
          "default",
          "flux-system",
          "kube-node-lease",
          "kube-public",
        ],
      },
    ]);
  });
});
