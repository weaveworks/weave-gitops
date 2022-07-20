import React from "react";
import { MemoryRouter } from "react-router-dom";
import { render } from "@testing-library/react";
import { withTheme } from "../../lib/test-utils";
import Link from "../Link";

describe("Link", () => {
  it("doesn't create a link for oci links", () => {
    const { container: link } = render(
      withTheme(<Link href="oci://ghcr.io/some/chart">Text</Link>)
    );
    const a = link.querySelector("a");
    expect(a).toBe(null);
    expect(link.textContent).toBe("Text");
  });
  it("creates a link for http links", () => {
    const { container: link } = render(
      withTheme(<Link href="http://google.com">Text</Link>)
    );
    const a = link.querySelector("a");
    expect(a).not.toBe(null);
    expect(a.href).toBe("http://google.com/");
    expect(a.textContent).toBe("Text");
  });
  it("creates a link for relative links", () => {
    const { container: link } = render(
      withTheme(<Link href="/some-page">Text</Link>)
    );
    const a = link.querySelector("a");
    expect(a).not.toBe(null);
    expect(a.href).toBe("http://localhost/some-page");
    expect(a.textContent).toBe("Text");
  });
  it("creates a router link for to links", () => {
    const { container: link } = render(
      withTheme(
        <MemoryRouter>
          <Link to="/some-page">Text</Link>
        </MemoryRouter>
      )
    );
    const a = link.querySelector("a");
    expect(a).not.toBe(null);
    expect(a.href).toBe("http://localhost/some-page");
    expect(a.textContent).toBe("Text");
  });
  it("makes to links relative when specifying an absolute link", () => {
    const { container: link } = render(
      withTheme(
        <MemoryRouter>
          <Link to="http://google.com">Text</Link>
        </MemoryRouter>
      )
    );
    const a = link.querySelector("a");
    expect(a).not.toBe(null);
    expect(a.href).toBe("http://localhost/http://google.com");
    expect(link.textContent).toBe("Text");
  });
  it("makes to links relative when specifying an oic link", () => {
    const { container: link } = render(
      withTheme(
        <MemoryRouter>
          <Link to="oci://ghcr.io/some/chart">Text</Link>
        </MemoryRouter>
      )
    );
    const a = link.querySelector("a");
    expect(a).not.toBe(null);
    expect(a.href).toBe("http://localhost/oci://ghcr.io/some/chart");
    expect(link.textContent).toBe("Text");
  });
});
