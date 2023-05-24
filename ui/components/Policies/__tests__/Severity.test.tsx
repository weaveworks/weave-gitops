import { render } from "@testing-library/react";
import React from "react";
import Severity from "../Utils/Severity";
import { withTheme } from "../../../lib/test-utils";

function checkSeverity(severity: string) {
  render(withTheme(<Severity severity={severity} />));
  const ele = document.querySelector(
    `[data-testid="${severity}"]`
  ) as HTMLElement;
  expect(ele).toBeTruthy();
  expect(ele.textContent).toEqual(severity);
}

describe("Severity", () => {
  it("check severities", async () => {
    ["high", "medium", "low"].forEach((s) => checkSeverity(s));
  });
});
