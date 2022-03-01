import React from "react";
import renderer from "react-test-renderer";
import { createMockClient, withContext, withTheme } from "../../lib/test-utils";
import Breadcrumbs from "../Breadcrumbs";

describe("snapshots", () => {
  it("renders", () => {
    const tree = renderer
      .create(
        withTheme(
          withContext(<Breadcrumbs />, "/automations", {
            applicationsClient: createMockClient({}),
          })
        )
      )
      .toJSON();
    expect(tree).toMatchSnapshot();
  });
  it("renders child route", () => {
    const tree = renderer
      .create(
        withTheme(
          withContext(<Breadcrumbs />, "/kustomization?name=flux", {
            applicationsClient: createMockClient({}),
          })
        )
      )
      .toJSON();
    expect(tree).toMatchSnapshot();
  });
});
