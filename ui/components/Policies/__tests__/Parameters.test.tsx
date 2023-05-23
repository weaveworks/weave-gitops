import { render } from "@testing-library/react";
import React from "react";
import { withTheme } from "../../../lib/test-utils";
import Parameters from "../Utilis/Parameters";
import { parseValue } from "../Utilis/PolicyUtilis";

const parameters = [
  {
    name: "automount_token",
    type: "boolean",
    value: {
      "@type": "type.googleapis.com/google.protobuf.BoolValue",
      value: false,
    },
    required: true,
    configRef: "",
  },
  {
    name: "exclude_namespace",
    type: "string",
    value: null,
    required: false,
    configRef: "",
  },
  {
    name: "exclude_label_key",
    type: "string",
    value: null,
    required: false,
    configRef: "",
  },
  {
    name: "exclude_label_value",
    type: "string",
    value: null,
    required: false,
    configRef: "",
  },
];

describe("Parameters", () => {
  it("validate Violation parameters display", async () => {
    render(withTheme(<Parameters parameters={parameters} />));
    parameters.forEach((p) => {
      const row = document.querySelector(`.${p.name}`);

      const name = row.querySelector(`[data-testid="Name"]`);
      expect(name.textContent).toEqual(`Name${p.name}`);

      const value = row.querySelector(`[data-testid="Value"]`);
      const parsedVal = parseValue(p);

      if (p.value) {
        expect(value.textContent).toContain(`${parsedVal}`);
      } else {
        expect(value.textContent).toContain("undefined");
      }

      const policyConfig = row.querySelector(
        `[data-testid="Policy Config Name"]`
      );
      expect(policyConfig.textContent).toEqual(
        `Policy Config Name${p.configRef || "-"}`
      );
    });
  });

  it("validate Policy parameters display", async () => {
    render(
      withTheme(<Parameters parameters={parameters} parameterType="policy" />)
    );
    parameters.forEach((p) => {
      const row = document.querySelector(`.${p.name}`);

      const name = row.querySelector(`[data-testid="Name"]`);
      expect(name.textContent).toEqual(`Name${p.name}`);

      const pType = row.querySelector(`[data-testid="Type"]`);
      expect(pType.textContent).toEqual(`Type${p.type || "-"}`);

      const value = row.querySelector(`[data-testid="Value"]`);
      const parsedVal = parseValue(p);

      if (p.value) {
        expect(value.textContent).toContain(`${parsedVal}`);
      } else {
        expect(value.textContent).toContain("undefined");
      }

      const requiredParam = row.querySelector(`[data-testid="Required"]`);
      expect(requiredParam.textContent).toEqual(
        `Required${p.required ? "True" : "False"}`
      );
    });
  });
});
