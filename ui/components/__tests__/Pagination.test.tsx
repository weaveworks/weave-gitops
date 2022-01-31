import { fireEvent, render, screen, within } from "@testing-library/react";
import "jest-styled-components";
import React from "react";
import { withTheme } from "../../lib/test-utils";
import Pagination from "../Pagination";

describe("Pagination", () => {
  const onForward = jest.fn(() => "");
  const onSkipForward = jest.fn(() => "");
  const onBack = jest.fn(() => "");
  const onSkipBack = jest.fn(() => "");
  const onSelect = jest.fn(() => "");
  it("has functioning navigation buttons and display text", () => {
    render(
      withTheme(
        <Pagination
          onForward={onForward}
          onSkipForward={onSkipForward}
          onBack={onBack}
          onSkipBack={onSkipBack}
          onSelect={onSelect}
          index={5}
          length={1}
          totalObjects={10}
        />
      )
    );
    const back = screen.getByLabelText("back one page");
    const skipBack = screen.getByLabelText("skip to first page");
    const forward = screen.getByLabelText("forward one page");
    const skipForward = screen.getByLabelText("skip to last page");
    const displayText = screen.getByText(/6 - 6 out of 10/);
    fireEvent.click(forward);
    expect(onForward).toHaveBeenCalledTimes(1);
    fireEvent.click(skipForward);
    expect(onSkipForward).toHaveBeenCalledTimes(1);
    fireEvent.click(skipBack);
    expect(onSkipBack).toHaveBeenCalledTimes(1);
    fireEvent.click(back);
    expect(onBack).toHaveBeenCalledTimes(1);
    expect(displayText).toBeTruthy();
  });
  it("disables buttons based on page location", () => {
    render(
      withTheme(
        <Pagination
          onForward={onForward}
          onSkipForward={onSkipForward}
          onBack={onBack}
          onSkipBack={onSkipBack}
          onSelect={onSelect}
          index={0}
          length={1}
          totalObjects={3}
        />
      )
    );
    const back = screen.getByLabelText("back one page");
    const skipBack = screen.getByLabelText("skip to first page");
    const forward = screen.getByLabelText("forward one page");
    const skipForward = screen.getByLabelText("skip to last page");
    expect(back.hasAttribute("disabled")).toBeTruthy();
    expect(skipBack.hasAttribute("disabled")).toBeTruthy();
    expect(forward.hasAttribute("disabled")).toBeFalsy();
    expect(skipForward.hasAttribute("disabled")).toBeFalsy();
  });
  it("has a functioning select", () => {
    render(
      withTheme(
        <Pagination
          onForward={onForward}
          onSkipForward={onSkipForward}
          onBack={onBack}
          onSkipBack={onSkipBack}
          onSelect={onSelect}
          index={0}
          length={1}
          totalObjects={3}
        />
      )
    );
    let select;
    screen.getAllByRole("button").forEach((button) => {
      if (button.getAttribute("aria-haspopup")) select = button;
    });
    fireEvent.mouseDown(select);
    const listbox = within(screen.getByRole("listbox"));
    fireEvent.click(listbox.getByText("50"));
    expect(select.innerHTML).toEqual("50");
  });
});
