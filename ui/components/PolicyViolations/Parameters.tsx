import React from "react";
import { PolicyValidationParam } from "../../lib/api/core/core.pb";
import { ParameterCell, ParameterWrapper, parseValue } from "./PolicyUtilis";

const Parameters = ({
  parameters,
}: {
  parameters: PolicyValidationParam[];
}) => {
  return (
    <>
      {parameters?.map((parameter) => (
        <ParameterWrapper key={parameter.name} id={parameter.name} wide gap="8">
          <ParameterCell label="Name" value={parameter.name}></ParameterCell>
          <ParameterCell
            label="Value"
            value={parseValue(parameter)}
          ></ParameterCell>
          <ParameterCell
            label="Policy Config Name"
            value={parameter.configRef || "-"}
          ></ParameterCell>
        </ParameterWrapper>
      ))}
    </>
  );
};

export default Parameters;
