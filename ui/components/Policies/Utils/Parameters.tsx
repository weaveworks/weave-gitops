import React from "react";
import styled from "styled-components";
import { PolicyValidationParam } from "../../../lib/api/core/core.pb";
import Flex from "../../Flex";
import Text from "../../Text";
import { parseValue } from "./PolicyUtils";

export const ParameterWrapper = styled(Flex)`
  border: 1px solid ${(props) => props.theme.colors.neutral20};
  box-sizing: border-box;
  border-radius: ${(props) => props.theme.spacing.xxs};
  padding: ${(props) => props.theme.spacing.base};
  width: 100%;
`;
export const ParameterCell = ({
  label,
  value,
}: {
  label: string;
  value: string | undefined;
}) => {
  return (
    <Flex wide column data-testid={label} gap="4">
      <Text color="neutral30">{label}</Text>
      <Text color="black">{value}</Text>
    </Flex>
  );
};

const Parameters = ({
  parameters,
  parameterType = "violations",
}: {
  parameters: PolicyValidationParam[];
  parameterType?: string;
}) => {
  return (
    <>
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
    </>
  );
};

export default Parameters;
