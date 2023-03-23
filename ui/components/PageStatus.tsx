import * as React from "react";
import styled from "styled-components";
import { Condition } from "../lib/api/core/types.pb";
import Button from "./Button";
import Flex from "./Flex";
import Icon from "./Icon";
import { computeMessage, getIndicatorInfo } from "./KubeStatusIndicator";
import { NoDialogDataTable } from "./PodDetail";
import Spacer from "./Spacer";
import Text from "./Text";

const SlideFlex = styled(Flex)<{ open: boolean }>`
  padding-top: ${(props) => props.theme.spacing.medium};
  max-height: ${(props) => (props.open ? "400px" : "0px")};
  transition-property: max-height;
  transition-duration: 0.5s;
  transition-timing-function: ease-in-out;
  overflow: hidden;
  overflow-x: auto;
`;

type StatusProps = {
  conditions: Condition[];
  suspended: boolean;
  showAll?: boolean;
  className?: string;
};

function PageStatus({
  conditions,
  suspended,
  className,
  showAll,
}: StatusProps) {
  let msg = suspended ? "Suspended" : computeMessage(conditions);
  const { icon, color } = getIndicatorInfo(suspended, conditions);
  const [open, setOpen] = React.useState(false);

  return (
    <Flex column wide>
      <Flex align className={className}>
        <Icon type={icon} color={color} size="medium" />
        <Spacer padding="xs" />
        <Text color="neutral30">{msg}</Text>
        {showAll && (
          <Button variant="text" onClick={() => setOpen(!open)}>
            show conditions
          </Button>
        )}
      </Flex>
      {showAll && (
        <SlideFlex open={open} wide>
          <Flex column wide>
            <Text bold size="large" color="neutral30">
              Conditions:
            </Text>
            <NoDialogDataTable
              hideSearchAndFilters
              fields={[
                { label: "Type", value: "type" },
                { label: "Status", value: "status" },
                { label: "Reason", value: "reason" },
                { label: "Message", value: "message" },
              ]}
              rows={conditions}
            />
          </Flex>
        </SlideFlex>
      )}
    </Flex>
  );
}
export default styled(PageStatus).attrs({ className: PageStatus.name })`
  transition-property: max-height;
  transition-duration: 0.5s;
  transition-timing-function: ease-in-out;
  color: ${(props) => props.theme.colors.neutral30};
`;
