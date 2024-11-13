import { fireEvent, render, screen } from "@testing-library/react";
import "jest-styled-components";
import _ from "lodash";
import React from "react";
import { MemoryRouter } from "react-router-dom";
import { LinkResolverProvider } from "../../contexts/LinkResolverContext";
import { convertResponse } from "../../hooks/objects";
import { objects } from "../../lib/fixtures/objects_table";
import { FluxObject } from "../../lib/objects";
import { withContext, withTheme } from "../../lib/test-utils";
import FluxObjectsTable from "../FluxObjectsTable";

describe("FluxObjectsTable", () => {
  let objs;

  beforeEach(() => {
    objs = _.map(objects, (o) =>
      convertResponse(o.obj.kind, { payload: JSON.stringify(o.obj) })
    );
  });
  it("renders", async () => {
    render(
      withTheme(
        withContext(
          <MemoryRouter>
            <FluxObjectsTable objects={objs} />
          </MemoryRouter>,
          "/",
          {}
        )
      )
    );

    const rows = document.querySelectorAll("tbody tr");
    expect(rows.length).toEqual(objects.length);

    const deploymentName = rows[0].querySelector("td:first-child");
    const link = deploymentName.querySelector("a");

    expect(link).toBeFalsy();
  });
  it("renders when a LinkResolver is present", () => {
    render(
      withTheme(
        withContext(
          <MemoryRouter>
            <LinkResolverProvider
              resolver={(type: string) => {
                if (type === "Deployment") {
                  return "/some-cool-url";
                }
              }}
            >
              <FluxObjectsTable objects={objs} />
            </LinkResolverProvider>
          </MemoryRouter>,
          "/",
          {}
        )
      )
    );
    const rows = document.querySelectorAll("tbody tr");

    // Since our resolver does not specify any behavior for a Service,
    // this should not have a link.
    const serviceName = rows[0].querySelector("td:first-child");
    const serviceLink = serviceName.querySelector("a");

    expect(serviceLink).toBeFalsy();

    const deploymentName = rows[1].querySelector("td:first-child");
    const link = deploymentName.querySelector("a");

    expect(link.href).toEqual("http://localhost/some-cool-url");
  });
  it("runs the onClick handler", () => {
    const onClick = jest.fn();

    render(
      withTheme(
        withContext(
          <MemoryRouter>
            <FluxObjectsTable onClick={onClick} objects={objs} />
          </MemoryRouter>,
          "/",
          {}
        )
      )
    );

    const rows = document.querySelectorAll("tbody tr");
    const deploymentName = rows[0].querySelector("td:first-child");
    const txt = deploymentName.querySelector("span > span");
    fireEvent(txt, new MouseEvent("click", { bubbles: true }));

    expect(onClick).toHaveBeenCalled();
  });
  it("should not run the onClick handler when the object is a Secret", async () => {
    const onClick = jest.fn();

    const secretObj = new FluxObject({
      payload: JSON.stringify({
        apiVersion: "v1",
        kind: "Secret",
        metadata: { name: "my-secret" },
      }),
    });

    render(
      withTheme(
        withContext(
          <MemoryRouter>
            <FluxObjectsTable
              onClick={onClick}
              objects={[...objs, secretObj]}
            />
          </MemoryRouter>,
          "/",
          {}
        )
      )
    );

    const secret = await screen.findByText("my-secret");
    fireEvent(secret, new MouseEvent("click", { bubbles: true }));
    expect(onClick).not.toHaveBeenCalled();
  });
});
