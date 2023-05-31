import React from "react";
import { PolicyValidationParam } from "../../../lib/api/core/core.pb";
import {
  ParameterCell,
  ParameterWrapper,
  SectionWrapper,
  parseValue,
} from "./PolicyUtilis";

const Parameters = ({
  parameters,
  parameterType = "violations",
}: {
  parameters: PolicyValidationParam[];
  parameterType?: string;
}) => {
  return (
    <SectionWrapper title="Parameters Definition:">
      {parameters?.map((parameter) => (
        <ParameterWrapper
          key={parameter.name}
          wide
          gap="8"
          className={parameter.name}
        >
          <ParameterCell label="Name" value={parameter.name} />

          {parameterType !== "violations" && (
            <ParameterCell label="Type" value={parameter.type || "-"} />
          )}

          <ParameterCell label="Value" value={parseValue(parameter)} />

          {parameterType !== "violations" && (
            <ParameterCell
              label="Required"
              value={parameter.required ? "True" : "False"}
            />
          )}

          {parameterType === "violations" && (
            <ParameterCell
              label="Policy Config Name"
              value={parameter.configRef || "-"}
            />
          )}
        </ParameterWrapper>
      ))}
    </SectionWrapper>
  );
};

export default Parameters;
