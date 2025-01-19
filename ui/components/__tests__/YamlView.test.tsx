import "jest-styled-components";
import { render } from "@testing-library/react";
import React from "react";
import { Kind } from "../../lib/api/core/types.pb";
import { withContext, withTheme } from "../../lib/test-utils";
import { createYamlCommand } from "../../lib/utils";
import YamlView from "../YamlView";

describe("YamlView", () => {
  describe("snapshots", () => {
    it("renders", () => {
      const tree = render(
        withTheme(
          withContext(
            <YamlView
              header={createYamlCommand(
                Kind.Kustomization,
                "podinfo",
                "flux-system",
              )}
              yaml="yaml\nyaml\nyaml\n"
            />,
            "",
            {},
          ),
        ),
      ).asFragment();
      expect(tree).toMatchSnapshot();
    });
  });
});
