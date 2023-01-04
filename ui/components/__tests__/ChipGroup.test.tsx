import { render, screen } from "@testing-library/react";
import "jest-styled-components";
import React from "react";
import { withTheme } from "../../lib/test-utils";
import ChipGroup from "../ChipGroup";
import { filterSeparator } from "../FilterDialog";

describe("ChipGroup", () => {
  const setActiveChips = jest.fn();
  const chipList = [
    "app",
    "app2",
    "app3",
    `appapp${filterSeparator}`,
    `app${filterSeparator}app`,
  ];

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
  it("adds 'null' to undefined values", () => {
    render(
      withTheme(
        <ChipGroup
          chips={chipList}
          onChipRemove={setActiveChips}
          onClearAll={() => jest.fn()}
        />
      )
    );
    expect(screen.queryByText(`appapp${filterSeparator}null`)).toBeTruthy();
    expect(screen.queryByText(`app${filterSeparator}app`)).toBeTruthy();
  });
});
