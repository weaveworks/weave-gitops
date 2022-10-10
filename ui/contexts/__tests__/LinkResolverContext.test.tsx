import { render, screen } from "@testing-library/react";
import * as React from "react";
import { LinkResolverProvider, useLinkResolver } from "../LinkResolverContext";

describe("LinkResolverProvider", () => {
  it("sets a link address via context", () => {
    const MyComponent = () => {
      const linkResolver = useLinkResolver();

      return (
        <a data-testid="link" href={linkResolver("")}>
          My Link
        </a>
      );
    };

    render(
      <LinkResolverProvider
        resolver={() => {
          return "/some/path";
        }}
      >
        <MyComponent />
      </LinkResolverProvider>
    );

    const link = screen.getByTestId("link");

    expect(link.getAttribute("href")).toEqual("/some/path");
  });
});
