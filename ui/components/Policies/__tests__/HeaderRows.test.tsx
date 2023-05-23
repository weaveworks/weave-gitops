import { render } from "@testing-library/react";
import React from "react";
import { withTheme } from "../../../lib/test-utils";
import HeaderRows, { Header } from "../Utilis/HeaderRows";

const headers: Header[] = [
  {
    rowkey: "Policy Name",
    value: "Controller ServiceAccount Tokens Automount",
  },
  {
    rowkey: "Application",
    value: "flux-system/external-secrets",
  },
  {
    rowkey: "Violation Time",
    value: "15 minutes ago",
  },
  {
    rowkey: "Category",
    value: " weave.categories.access-control",
    visible: false,
  },
];

describe("HeaderRows", () => {
  it("validate rows", async () => {
    render(withTheme(<HeaderRows headers={headers} />));
    headers.forEach((h) => {
      const ele = document.querySelector(
        `[data-testid="${h.rowkey}"]`
      ) as HTMLElement;
      if (h.visible !== false) {
        expect(ele).toBeTruthy();
        expect(ele.textContent).toEqual(`${h.rowkey}:${h.value}`);
      } else expect(ele).toBeFalsy();
    });
  });
});
