import "jest-styled-components";
import React from "react";
import renderer from "react-test-renderer";
import { FluxObjectKind } from "../../lib/api/core/types.pb";
import { withContext, withTheme } from "../../lib/test-utils";
import YamlView from "../YamlView";

describe("YamlView", () => {
  describe("snapshots", () => {
    it("renders", () => {
      const tree = renderer
        .create(
          withTheme(
            withContext(
              <YamlView
                object={{
                  kind: FluxObjectKind.KindKustomization,
                  name: "podinfo",
                  namespace: "flux-system",
                }}
                yaml="yaml\nyaml\nyaml\n"
              />,
              "",
              {}
            )
          )
        )
        .toJSON();
      expect(tree).toMatchSnapshot();
    });
  });
});
