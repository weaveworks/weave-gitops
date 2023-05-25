import { render } from "@testing-library/react";
import "jest-styled-components";
import React from "react";
import { BrowserRouter } from "react-router-dom";
import { withTheme } from "../../lib/test-utils";
import { V2Routes } from "../../lib/types";
import Breadcrumbs, { Breadcrumb } from "../Breadcrumbs";

function checkPath(path: Breadcrumb[]) {
  render(
    withTheme(
      <BrowserRouter>
        <Breadcrumbs path={path} />
      </BrowserRouter>
    )
  );

  path.forEach(({ url, label }) => {
    if (url) {
      const ele = document.querySelector(
        `[data-testid="link-${label}"]`
      ) as HTMLLinkElement;
      expect(ele).toBeTruthy();
      expect(ele.textContent).toEqual(label);
      expect(ele.href).toEqual(`${location.origin}${url}`);
    } else {
      const ele = document.querySelector(
        `[data-testid="text-${label}"]`
      ) as HTMLElement;
      expect(ele).toBeTruthy();
      expect(ele.textContent).toEqual(label);
    }
  });
}

describe("Breadcrumbs", () => {
  it("check different routes", async () => {
    [
      [{ label: "Applications" }],
      [
        { label: "Applications", url: V2Routes.Automations },
        { label: "flux-system" },
      ],
      [
        { label: "Applications", url: V2Routes.Automations },
        {
          label: "flux-system",
          url: "/kustomization/details?clusterName=Default&name=flux-system&namespace=flux-system",
        },
        { label: "violation-message" },
      ],
    ].forEach((s) => checkPath(s));
  });
});
