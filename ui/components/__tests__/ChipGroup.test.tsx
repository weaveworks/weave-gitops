import { render, screen } from "@testing-library/react";
import "jest-styled-components";
import React from "react";
import { withTheme } from "../../lib/test-utils";
import ChipGroup from "../ChipGroup";

describe("ChipGroup", () => {
  const setActiveChips = jest.fn();
  const chipList = ["app", "app2", "app3"];

  it("should render chips", () => {
    render(
      withTheme(
        <ChipGroup
          chips={chipList}
          onChipRemove={setActiveChips}
          onClearAll={() => jest.fn()}
        />
      )
    );
    expect(screen.queryByText("app")).toBeTruthy();
    expect(screen.queryByText("app3")).toBeTruthy();
    expect(screen.queryByText("Clear All")).toBeTruthy();
  });
});
